package domain

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const quorumPercent uint32 = 60

type Engine struct {
	mu sync.RWMutex

	directory IdentityDirectory

	goals            map[string]Goal
	goalReservations map[string]string
	proposals        map[string]Proposal
	proposalHistory  map[string]Proposal
	voteRequests     map[string]VoteRequest
	authorizations   map[string]VoteAuthorization
	confirmations    map[string]BallotConfirmation
	ballots          map[string]Ballot
	decisions        map[string]Decision
	events           []Event
	eventCursor      uint64
}

func NewEngine(directory IdentityDirectory) *Engine {
	if directory == nil {
		panic("civilization: identity directory is required")
	}
	return &Engine{
		directory:        directory,
		goals:            make(map[string]Goal),
		goalReservations: make(map[string]string),
		proposals:        make(map[string]Proposal),
		proposalHistory:  make(map[string]Proposal),
		voteRequests:     make(map[string]VoteRequest),
		authorizations:   make(map[string]VoteAuthorization),
		confirmations:    make(map[string]BallotConfirmation),
		ballots:          make(map[string]Ballot),
		decisions:        make(map[string]Decision),
	}
}

func (e *Engine) CreateGoal(actorPraxeonID string, goal Goal, now time.Time) (Goal, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if actorPraxeonID == "" || goal.ID == "" || goal.Title == "" || goal.OwnerPraxeonID != actorPraxeonID {
		return Goal{}, ErrInvalidInput
	}
	if goal.Type == GoalTypeCivilization {
		return Goal{}, ErrCivilizationGoalAuthority
	}
	if goal.Type != GoalTypePraxeon && goal.Type != GoalTypeAgent {
		return Goal{}, ErrInvalidInput
	}
	if !e.activePraxeon(actorPraxeonID) {
		return Goal{}, ErrNotEligible
	}
	if _, ok := e.goals[goal.ID]; ok {
		return Goal{}, ErrAlreadyExists
	}
	if _, reserved := e.goalReservations[goal.ID]; reserved {
		return Goal{}, ErrAlreadyExists
	}
	if err := validateGoalChecks(goal.SuccessChecks); err != nil {
		return Goal{}, err
	}
	if goal.Type == GoalTypeAgent && len(goal.AssignedAgentIDs) == 0 {
		return Goal{}, ErrInvalidInput
	}
	for _, agentID := range goal.AssignedAgentIDs {
		owner, ok := e.directory.PraxeonForAgent(agentID)
		if !ok || owner != actorPraxeonID {
			return Goal{}, ErrAgentOwnership
		}
	}

	goal.CreatedAt = canonicalTime(now)
	goal.Status = GoalStatusActive
	goal.GovernanceProposalID = ""
	goal = cloneGoal(goal)
	e.goals[goal.ID] = goal
	e.emit(EventTypeGoalCreated, goal.ID, actorPraxeonID, "", EventVisibilityPraxeon, goal.DescriptionHash, now)
	return cloneGoal(goal), nil
}

func (e *Engine) FileProposal(actorPraxeonID string, proposal Proposal, now time.Time) (Proposal, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if actorPraxeonID == "" || proposal.ProposerPraxeonID != actorPraxeonID || !e.activePraxeon(actorPraxeonID) {
		return Proposal{}, ErrNotEligible
	}
	if _, ok := e.proposals[proposal.ID]; ok {
		return Proposal{}, ErrAlreadyExists
	}
	if err := validateProposal(proposal); err != nil {
		return Proposal{}, err
	}
	if proposal.Type == ProposalTypeCivilizationGoal {
		if proposal.CivilizationGoalSpec == nil {
			return Proposal{}, ErrInvalidInput
		}
		if err := validateCivilizationGoal(*proposal.CivilizationGoalSpec); err != nil {
			return Proposal{}, err
		}
		if _, exists := e.goals[proposal.CivilizationGoalSpec.ID]; exists {
			return Proposal{}, ErrAlreadyExists
		}
		if _, reserved := e.goalReservations[proposal.CivilizationGoalSpec.ID]; reserved {
			return Proposal{}, ErrAlreadyExists
		}
	}

	proposal.Version = 1
	proposal.SeconderPraxeonID = ""
	proposal.EligiblePraxeonIDs = nil
	proposal.Status = ProposalStatusFiled
	proposal.Approval = requiredApproval(proposal)
	proposal.Schedule = canonicalSchedule(proposal.Schedule)
	proposal = cloneProposal(proposal)
	e.proposals[proposal.ID] = proposal
	e.proposalHistory[proposalVersionKey(proposal.ID, proposal.Version)] = proposal
	if proposal.CivilizationGoalSpec != nil {
		e.goalReservations[proposal.CivilizationGoalSpec.ID] = proposal.ID
	}
	e.emit(EventTypeProposalFiled, proposal.ID, actorPraxeonID, "", EventVisibilityPublic, proposal.Hash, now)
	return cloneProposal(proposal), nil
}

func (e *Engine) SecondProposal(actorPraxeonID, proposalID string, expectedVersion uint64, now time.Time) (Proposal, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	proposal, ok := e.proposals[proposalID]
	if !ok {
		return Proposal{}, ErrNotFound
	}
	if proposal.Version != expectedVersion {
		return Proposal{}, ErrVersionMismatch
	}
	if proposal.Status != ProposalStatusFiled {
		return Proposal{}, ErrInvalidState
	}
	if !canonicalTime(now).Before(proposal.Schedule.NoticeOpensAt) {
		return Proposal{}, ErrInvalidState
	}
	if actorPraxeonID == proposal.ProposerPraxeonID {
		return Proposal{}, ErrSameSeconder
	}
	if !e.activePraxeon(actorPraxeonID) {
		return Proposal{}, ErrNotEligible
	}

	routes := activeRoutes(e.directory.Routes())
	if len(routes) == 0 {
		return Proposal{}, ErrNotEligible
	}
	proposal.SeconderPraxeonID = actorPraxeonID
	proposal.Status = ProposalStatusNoticed
	proposal.EligiblePraxeonIDs = make([]string, 0, len(routes))
	deliveredAt := canonicalTime(now)
	for _, route := range routes {
		proposal.EligiblePraxeonIDs = append(proposal.EligiblePraxeonIDs, route.PraxeonID)
		recipient := route.GovernanceAgentID
		if recipient == "" {
			recipient = route.PrimaryAgentID
		}
		request := VoteRequest{
			ID:                voteRequestID(proposal.ID, proposal.Version, route.PraxeonID),
			ProposalID:        proposal.ID,
			ProposalVersion:   proposal.Version,
			ProposalHash:      proposal.Hash,
			EligiblePraxeonID: route.PraxeonID,
			RecipientAgentID:  recipient,
			BallotDeadline:    proposal.Schedule.BallotClosesAt,
		}
		if recipient == "" {
			request.DeliveryStatus = VoteRequestDeliveryStatusFailed
			request.DeliveryFailure = "no_active_agent_route"
		} else {
			request.DeliveryStatus = VoteRequestDeliveryStatusDelivered
			request.DeliveredAt = deliveredAt
		}
		e.voteRequests[request.ID] = request
	}
	sort.Strings(proposal.EligiblePraxeonIDs)
	proposal.EligibilityHash = eligibilitySnapshotHash(proposal.EligiblePraxeonIDs)
	proposal.EligibilityFrozenAt = deliveredAt
	e.proposals[proposal.ID] = cloneProposal(proposal)
	e.proposalHistory[proposalVersionKey(proposal.ID, proposal.Version)] = cloneProposal(proposal)
	e.emit(EventTypeProposalNoticed, proposal.ID, actorPraxeonID, "", EventVisibilityPublic, proposal.Hash, now)
	return cloneProposal(proposal), nil
}

func (e *Engine) AmendProposal(actorPraxeonID string, expectedVersion uint64, replacement Proposal, now time.Time) (Proposal, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	current, ok := e.proposals[replacement.ID]
	if !ok {
		return Proposal{}, ErrNotFound
	}
	if current.Version != expectedVersion {
		return Proposal{}, ErrVersionMismatch
	}
	if current.Status.Terminal() {
		return Proposal{}, ErrInvalidState
	}
	if !canonicalTime(now).Before(current.Schedule.BallotClosesAt) {
		return Proposal{}, ErrInvalidState
	}
	if actorPraxeonID != current.ProposerPraxeonID || replacement.ProposerPraxeonID != current.ProposerPraxeonID {
		return Proposal{}, ErrNotEligible
	}
	if replacement.Type != current.Type {
		return Proposal{}, ErrInvalidInput
	}
	if err := validateProposal(replacement); err != nil {
		return Proposal{}, err
	}
	if replacement.Type == ProposalTypeCivilizationGoal {
		if replacement.CivilizationGoalSpec == nil {
			return Proposal{}, ErrInvalidInput
		}
		if err := validateCivilizationGoal(*replacement.CivilizationGoalSpec); err != nil {
			return Proposal{}, err
		}
		if replacement.CivilizationGoalSpec.ID != current.CivilizationGoalSpec.ID {
			return Proposal{}, ErrInvalidInput
		}
	}

	current.Status = ProposalStatusInvalid
	e.proposalHistory[proposalVersionKey(current.ID, current.Version)] = cloneProposal(current)
	for key, request := range e.voteRequests {
		if request.ProposalID == current.ID && request.ProposalVersion == current.Version {
			request.Invalidated = true
			e.voteRequests[key] = request
		}
	}
	for key, ballot := range e.ballots {
		if ballot.ProposalID == current.ID && ballot.ProposalVersion == current.Version {
			ballot.Invalidated = true
			e.ballots[key] = ballot
		}
	}

	replacement.Version = current.Version + 1
	replacement.SeconderPraxeonID = ""
	replacement.EligiblePraxeonIDs = nil
	replacement.Status = ProposalStatusFiled
	replacement.Approval = requiredApproval(replacement)
	replacement.Schedule = canonicalSchedule(replacement.Schedule)
	e.proposals[replacement.ID] = cloneProposal(replacement)
	e.proposalHistory[proposalVersionKey(replacement.ID, replacement.Version)] = cloneProposal(replacement)
	e.emit(EventTypeProposalAmended, replacement.ID, actorPraxeonID, "", EventVisibilityPublic, replacement.Hash, now)
	return cloneProposal(replacement), nil
}

func (e *Engine) AcknowledgeVoteRequest(actorAgentID, requestID, proposalHash string, now time.Time) (VoteRequest, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	request, ok := e.voteRequests[requestID]
	if !ok {
		return VoteRequest{}, ErrNotFound
	}
	if request.RecipientAgentID != actorAgentID {
		return VoteRequest{}, ErrAgentOwnership
	}
	if request.DeliveryStatus != VoteRequestDeliveryStatusDelivered || request.Invalidated {
		return VoteRequest{}, ErrInvalidState
	}
	proposal, ok := e.proposals[request.ProposalID]
	if !ok || proposal.Version != request.ProposalVersion || proposal.Status.Terminal() {
		return VoteRequest{}, ErrInvalidState
	}
	if proposalHash == "" || proposalHash != request.ProposalHash || proposalHash != proposal.Hash {
		return VoteRequest{}, ErrProposalHashMismatch
	}
	if !canonicalTime(now).Before(request.BallotDeadline) {
		return VoteRequest{}, ErrProposalClosed
	}
	if !request.AcknowledgedAt.IsZero() {
		return request, nil
	}
	request.AcknowledgedAt = canonicalTime(now)
	e.voteRequests[request.ID] = request
	e.emit(EventTypeVoteRequestAcknowledged, request.ID, request.EligiblePraxeonID, actorAgentID, EventVisibilityPublic, proposalHash, now)
	return request, nil
}

func (e *Engine) SetVoteAuthorization(actorPraxeonID string, authorization VoteAuthorization, now time.Time) (VoteAuthorization, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if actorPraxeonID == "" || authorization.ID == "" || authorization.PraxeonID != actorPraxeonID {
		return VoteAuthorization{}, ErrInvalidInput
	}
	if _, exists := e.authorizations[authorization.ID]; exists {
		return VoteAuthorization{}, ErrAlreadyExists
	}
	owner, ok := e.directory.PraxeonForAgent(authorization.AgentID)
	if !ok || owner != actorPraxeonID {
		return VoteAuthorization{}, ErrAgentOwnership
	}
	if authorization.Mode != VoteModeNotifyOnly && authorization.Mode != VoteModeConfirmEach && authorization.Mode != VoteModePolicyVote {
		return VoteAuthorization{}, ErrInvalidInput
	}
	if !authorization.ValidFrom.Before(authorization.ExpiresAt) || authorization.MaxRisk == RiskLevelUnspecified {
		return VoteAuthorization{}, ErrInvalidInput
	}
	if len(authorization.AllowedProposalTypes) == 0 || len(authorization.AllowedDomains) == 0 {
		return VoteAuthorization{}, ErrInvalidInput
	}
	if authorization.Mode == VoteModePolicyVote && authorization.DecisionPolicyHash == "" {
		return VoteAuthorization{}, ErrInvalidInput
	}
	authorization.ValidFrom = canonicalTime(authorization.ValidFrom)
	authorization.ExpiresAt = canonicalTime(authorization.ExpiresAt)
	authorization.Revoked = false
	authorization.Hash = voteAuthorizationHash(authorization)
	authorization = cloneAuthorization(authorization)
	e.authorizations[authorization.ID] = authorization
	e.emit(EventTypeVoteAuthorizationChanged, authorization.ID, actorPraxeonID, authorization.AgentID, EventVisibilityPraxeon, authorization.Hash, now)
	return cloneAuthorization(authorization), nil
}

func (e *Engine) RevokeVoteAuthorization(actorPraxeonID, authorizationID string, now time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	authorization, ok := e.authorizations[authorizationID]
	if !ok {
		return ErrNotFound
	}
	if authorization.PraxeonID != actorPraxeonID {
		return ErrNotEligible
	}
	authorization.Revoked = true
	e.authorizations[authorizationID] = authorization
	e.emit(EventTypeVoteAuthorizationChanged, authorization.ID, actorPraxeonID, authorization.AgentID, EventVisibilityPraxeon, authorization.Hash, now)
	return nil
}

func (e *Engine) ConfirmBallot(actorPraxeonID string, confirmation BallotConfirmation, now time.Time) (BallotConfirmation, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if actorPraxeonID == "" || confirmation.ID == "" || confirmation.PraxeonID != actorPraxeonID || !confirmation.Choice.Valid() {
		return BallotConfirmation{}, ErrInvalidInput
	}
	if _, exists := e.confirmations[confirmation.ID]; exists {
		return BallotConfirmation{}, ErrAlreadyExists
	}
	owner, ok := e.directory.PraxeonForAgent(confirmation.AgentID)
	if !ok || owner != actorPraxeonID {
		return BallotConfirmation{}, ErrAgentOwnership
	}
	proposal, ok := e.proposals[confirmation.ProposalID]
	if !ok {
		return BallotConfirmation{}, ErrNotFound
	}
	if proposal.Version != confirmation.ProposalVersion {
		return BallotConfirmation{}, ErrVersionMismatch
	}
	if !contains(proposal.EligiblePraxeonIDs, actorPraxeonID) {
		return BallotConfirmation{}, ErrNotEligible
	}
	if !canonicalTime(now).Before(confirmation.ExpiresAt) || confirmation.ExpiresAt.After(proposal.Schedule.BallotClosesAt) {
		return BallotConfirmation{}, ErrInvalidInput
	}
	confirmation.ExpiresAt = canonicalTime(confirmation.ExpiresAt)
	confirmation.Used = false
	e.confirmations[confirmation.ID] = confirmation
	e.emit(EventTypeBallotConfirmed, confirmation.ID, actorPraxeonID, confirmation.AgentID, EventVisibilityPraxeon, confirmation.ProposalID, now)
	return confirmation, nil
}

func (e *Engine) SubmitBallot(actorID string, submitted Ballot, expectedSequence uint64, now time.Time) (Ballot, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now = canonicalTime(now)
	proposal, ok := e.proposals[submitted.ProposalID]
	if !ok {
		return Ballot{}, ErrNotFound
	}
	if proposal.Version != submitted.ProposalVersion {
		return Ballot{}, ErrVersionMismatch
	}
	if now.Before(proposal.Schedule.BallotOpensAt) {
		return Ballot{}, ErrProposalNotOpen
	}
	if !now.Before(proposal.Schedule.BallotClosesAt) {
		return Ballot{}, ErrProposalClosed
	}
	if !submitted.Choice.Valid() || !contains(proposal.EligiblePraxeonIDs, submitted.VoterPraxeonID) {
		return Ballot{}, ErrNotEligible
	}

	key := ballotKey(proposal.ID, proposal.Version, submitted.VoterPraxeonID)
	current, hasCurrent := e.ballots[key]
	if hasCurrent && current.Invalidated {
		hasCurrent = false
		current = Ballot{}
	}
	currentSequence := uint64(0)
	if hasCurrent {
		currentSequence = current.Sequence
	}
	if currentSequence != expectedSequence {
		return Ballot{}, ErrSequenceMismatch
	}

	accepted := Ballot{
		ProposalID:      proposal.ID,
		ProposalVersion: proposal.Version,
		ProposalHash:    proposal.Hash,
		VoterPraxeonID:  submitted.VoterPraxeonID,
		Choice:          submitted.Choice,
		SubmittedBy:     submitted.SubmittedBy,
		SignerID:        actorID,
		AuthorizationID: submitted.AuthorizationID,
		ConfirmationID:  submitted.ConfirmationID,
		ReasonHash:      submitted.ReasonHash,
		Sequence:        currentSequence + 1,
		SubmittedAt:     now,
	}

	switch submitted.SubmittedBy {
	case SubmitterTypePraxeon:
		if actorID != submitted.VoterPraxeonID {
			return Ballot{}, ErrNotEligible
		}
		accepted.PraxeonLock = true
	case SubmitterTypeAgent:
		if hasCurrent && current.PraxeonLock {
			return Ballot{}, ErrPraxeonLocked
		}
		owner, ok := e.directory.PraxeonForAgent(actorID)
		if !ok || owner != submitted.VoterPraxeonID {
			return Ballot{}, ErrAgentOwnership
		}
		if submitted.AuthorizationID == "" {
			confirmation, err := e.validConfirmation(submitted.ConfirmationID, submitted.VoterPraxeonID, actorID, proposal, submitted.Choice, now)
			if err != nil {
				return Ballot{}, err
			}
			confirmation.Used = true
			e.confirmations[confirmation.ID] = confirmation
		} else {
			authorization, err := e.validAuthorization(submitted.AuthorizationID, submitted.VoterPraxeonID, actorID, proposal, now)
			if err != nil {
				return Ballot{}, err
			}
			if submitted.Choice == BallotChoiceAbstain && !authorization.CanAbstain {
				return Ballot{}, ErrAuthorizationInvalid
			}
			if authorization.RequiresReason && submitted.ReasonHash == "" {
				return Ballot{}, ErrAuthorizationInvalid
			}
			accepted.AuthorizationHash = authorization.Hash
			switch authorization.Mode {
			case VoteModeNotifyOnly:
				return Ballot{}, ErrNotifyOnly
			case VoteModeConfirmEach:
				confirmation, err := e.validConfirmation(submitted.ConfirmationID, submitted.VoterPraxeonID, actorID, proposal, submitted.Choice, now)
				if err != nil {
					return Ballot{}, err
				}
				confirmation.Used = true
				e.confirmations[confirmation.ID] = confirmation
			case VoteModePolicyVote:
				// Scope and policy hash were verified when the authorization was loaded.
			default:
				return Ballot{}, ErrAuthorizationInvalid
			}
		}
	default:
		return Ballot{}, ErrInvalidInput
	}

	accepted.ID = fmt.Sprintf("ballot:%s:%d:%s:%d", proposal.ID, proposal.Version, accepted.VoterPraxeonID, accepted.Sequence)
	e.ballots[key] = accepted
	e.emit(EventTypeBallotSubmitted, accepted.ID, accepted.VoterPraxeonID, actorID, EventVisibilityPublic, accepted.ProposalID, now)
	return accepted, nil
}

func (e *Engine) FinalizeProposal(proposalID string, expectedVersion uint64, now time.Time) (Decision, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.decisions[proposalID]; exists {
		return Decision{}, ErrFinalized
	}
	proposal, ok := e.proposals[proposalID]
	if !ok {
		return Decision{}, ErrNotFound
	}
	if proposal.Version != expectedVersion {
		return Decision{}, ErrVersionMismatch
	}
	if len(proposal.EligiblePraxeonIDs) == 0 {
		return Decision{}, ErrInvalidState
	}
	now = canonicalTime(now)
	if now.Before(proposal.Schedule.BallotClosesAt) {
		return Decision{}, ErrProposalNotClosed
	}

	decision := Decision{
		ID:              fmt.Sprintf("decision:%s:%d", proposal.ID, proposal.Version),
		ProposalID:      proposal.ID,
		ProposalVersion: proposal.Version,
		Eligible:        uint32(len(proposal.EligiblePraxeonIDs)),
		Approval:        proposal.Approval,
		FinalizedAt:     now,
	}
	for _, voterID := range proposal.EligiblePraxeonIDs {
		request, requestExists := e.voteRequests[voteRequestID(proposal.ID, proposal.Version, voterID)]
		if requestExists && request.DeliveryStatus == VoteRequestDeliveryStatusDelivered {
			decision.Delivered++
		}
		if requestExists && !request.AcknowledgedAt.IsZero() {
			decision.Acknowledged++
		}
		ballot, exists := e.ballots[ballotKey(proposal.ID, proposal.Version, voterID)]
		if !exists || ballot.Invalidated {
			continue
		}
		switch ballot.Choice {
		case BallotChoiceYes:
			decision.Yes++
		case BallotChoiceNo:
			decision.No++
		case BallotChoiceAbstain:
			decision.Abstain++
		}
	}
	participating := decision.Yes + decision.No + decision.Abstain
	decision.NotVoted = decision.Eligible - participating
	if participating*100 < decision.Eligible*quorumPercent {
		decision.Outcome = ProposalStatusExpiredNoQuorum
	} else if approved(decision.Yes, decision.No, proposal.Approval) {
		decision.Outcome = ProposalStatusRatified
	} else {
		decision.Outcome = ProposalStatusRejected
	}

	if decision.Outcome == ProposalStatusRatified && proposal.Type == ProposalTypeCivilizationGoal {
		spec := proposal.CivilizationGoalSpec
		if spec == nil {
			return Decision{}, ErrInvalidInput
		}
		if _, exists := e.goals[spec.ID]; exists {
			return Decision{}, ErrAlreadyExists
		}
		goal := Goal{
			ID:                   spec.ID,
			Type:                 GoalTypeCivilization,
			Title:                spec.Title,
			DescriptionHash:      spec.DescriptionHash,
			GovernanceProposalID: proposal.ID,
			AssignedAgentIDs:     append([]string(nil), spec.AssignedAgentIDs...),
			Status:               GoalStatusActive,
			SuccessChecks:        cloneGoalChecks(spec.SuccessChecks),
			CreatedAt:            now,
			TargetAt:             canonicalTime(spec.TargetAt),
		}
		e.goals[goal.ID] = goal
		e.emit(EventTypeCivilizationGoalActivated, goal.ID, "", "", EventVisibilityPublic, proposal.Hash, now)
	}
	if proposal.CivilizationGoalSpec != nil {
		delete(e.goalReservations, proposal.CivilizationGoalSpec.ID)
	}

	proposal.Status = decision.Outcome
	e.proposals[proposal.ID] = cloneProposal(proposal)
	e.proposalHistory[proposalVersionKey(proposal.ID, proposal.Version)] = cloneProposal(proposal)
	e.decisions[proposal.ID] = decision
	e.emit(EventTypeDecisionFinalized, decision.ID, "", "", EventVisibilityPublic, proposal.Hash, now)
	return decision, nil
}

func (e *Engine) AgentSnapshot(agentID string, now time.Time) (AgentSnapshot, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	praxeonID, ok := e.directory.PraxeonForAgent(agentID)
	if !ok {
		return AgentSnapshot{}, ErrAgentOwnership
	}
	now = canonicalTime(now)
	snapshot := AgentSnapshot{
		ID:          fmt.Sprintf("snapshot:%s:%d", agentID, e.eventCursor),
		AgentID:     agentID,
		PraxeonID:   praxeonID,
		EventCursor: e.eventCursor,
		GeneratedAt: now,
	}
	for _, goal := range e.goals {
		if !goalVisibleToAgent(goal, agentID) {
			continue
		}
		if goal.Status == GoalStatusActive || goal.Status == GoalStatusBlocked || goal.Status == GoalStatusPaused {
			snapshot.ActiveGoals = append(snapshot.ActiveGoals, cloneGoal(goal))
		}
	}
	for _, request := range e.voteRequests {
		if request.Invalidated || request.RecipientAgentID != agentID {
			continue
		}
		proposal, exists := e.proposals[request.ProposalID]
		if !exists || proposal.Version != request.ProposalVersion || proposal.Status.Terminal() {
			continue
		}
		snapshot.VoteRequests = append(snapshot.VoteRequests, request)
		ballot, hasBallot := e.ballots[ballotKey(proposal.ID, proposal.Version, praxeonID)]
		if hasBallot && !ballot.Invalidated {
			continue
		}
		if !e.hasPolicyAuthorization(praxeonID, agentID, proposal, now) {
			snapshot.RequiredConfirmationRequestIDs = append(snapshot.RequiredConfirmationRequestIDs, request.ID)
		}
	}
	sort.Slice(snapshot.ActiveGoals, func(i, j int) bool { return snapshot.ActiveGoals[i].ID < snapshot.ActiveGoals[j].ID })
	sort.Slice(snapshot.VoteRequests, func(i, j int) bool { return snapshot.VoteRequests[i].ID < snapshot.VoteRequests[j].ID })
	sort.Strings(snapshot.RequiredConfirmationRequestIDs)
	return snapshot, nil
}

func (e *Engine) Goal(goalID string) (Goal, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	goal, ok := e.goals[goalID]
	if !ok {
		return Goal{}, ErrNotFound
	}
	return cloneGoal(goal), nil
}

func (e *Engine) Proposal(proposalID string, now time.Time) (Proposal, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	proposal, ok := e.proposals[proposalID]
	if !ok {
		return Proposal{}, ErrNotFound
	}
	proposal.Status = proposal.StatusAt(canonicalTime(now))
	return cloneProposal(proposal), nil
}

func (e *Engine) VoteRequest(proposalID string, version uint64, praxeonID string) (VoteRequest, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	request, ok := e.voteRequests[voteRequestID(proposalID, version, praxeonID)]
	if !ok {
		return VoteRequest{}, ErrNotFound
	}
	return request, nil
}

func (e *Engine) Ballot(proposalID string, version uint64, praxeonID string) (Ballot, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	ballot, ok := e.ballots[ballotKey(proposalID, version, praxeonID)]
	if !ok {
		return Ballot{}, ErrNotFound
	}
	return ballot, nil
}

func (e *Engine) Decision(proposalID string) (Decision, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	decision, ok := e.decisions[proposalID]
	if !ok {
		return Decision{}, ErrNotFound
	}
	return decision, nil
}

func (e *Engine) Events(agentID string, afterCursor uint64, limit int) ([]Event, uint64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	praxeonID, ok := e.directory.PraxeonForAgent(agentID)
	if !ok {
		return nil, afterCursor, ErrAgentOwnership
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	visible := make([]Event, 0, limit)
	nextCursor := afterCursor
	for _, event := range e.events {
		if event.Cursor <= afterCursor {
			continue
		}
		nextCursor = event.Cursor
		if !eventVisibleToAgent(event, praxeonID, agentID) {
			continue
		}
		visible = append(visible, event)
		if len(visible) == limit {
			break
		}
	}
	return visible, nextCursor, nil
}

func (e *Engine) validAuthorization(authorizationID, praxeonID, agentID string, proposal Proposal, now time.Time) (VoteAuthorization, error) {
	if authorizationID == "" {
		return VoteAuthorization{}, ErrAuthorizationRequired
	}
	authorization, ok := e.authorizations[authorizationID]
	if !ok {
		return VoteAuthorization{}, ErrAuthorizationRequired
	}
	if authorization.Revoked || now.Before(authorization.ValidFrom) || !now.Before(authorization.ExpiresAt) {
		return VoteAuthorization{}, ErrAuthorizationExpired
	}
	if authorization.PraxeonID != praxeonID || authorization.AgentID != agentID {
		return VoteAuthorization{}, ErrAuthorizationInvalid
	}
	if !containsProposalType(authorization.AllowedProposalTypes, proposal.Type) || !contains(authorization.AllowedDomains, proposal.Domain) || proposal.Risk > authorization.MaxRisk {
		return VoteAuthorization{}, ErrAuthorizationInvalid
	}
	if len(authorization.AllowedProposalIDs) > 0 && !contains(authorization.AllowedProposalIDs, proposal.ID) {
		return VoteAuthorization{}, ErrAuthorizationInvalid
	}
	if proposal.Risk >= RiskLevelHigh && authorization.Mode == VoteModePolicyVote {
		if len(authorization.AllowedProposalIDs) != 1 || authorization.AllowedProposalIDs[0] != proposal.ID {
			return VoteAuthorization{}, ErrAuthorizationInvalid
		}
	}
	if authorization.Mode == VoteModePolicyVote && authorization.DecisionPolicyHash == "" {
		return VoteAuthorization{}, ErrAuthorizationInvalid
	}
	return authorization, nil
}

func (e *Engine) validConfirmation(confirmationID, praxeonID, agentID string, proposal Proposal, choice BallotChoice, now time.Time) (BallotConfirmation, error) {
	if confirmationID == "" {
		return BallotConfirmation{}, ErrConfirmationRequired
	}
	confirmation, ok := e.confirmations[confirmationID]
	if !ok {
		return BallotConfirmation{}, ErrConfirmationRequired
	}
	if confirmation.Used || !now.Before(confirmation.ExpiresAt) || confirmation.PraxeonID != praxeonID || confirmation.AgentID != agentID || confirmation.ProposalID != proposal.ID || confirmation.ProposalVersion != proposal.Version || confirmation.Choice != choice {
		return BallotConfirmation{}, ErrConfirmationInvalid
	}
	return confirmation, nil
}

func (e *Engine) hasPolicyAuthorization(praxeonID, agentID string, proposal Proposal, now time.Time) bool {
	for _, authorization := range e.authorizations {
		if authorization.PraxeonID != praxeonID || authorization.AgentID != agentID || authorization.Mode != VoteModePolicyVote {
			continue
		}
		if _, err := e.validAuthorization(authorization.ID, praxeonID, agentID, proposal, now); err == nil {
			return true
		}
	}
	return false
}

func (e *Engine) activePraxeon(praxeonID string) bool {
	for _, route := range e.directory.Routes() {
		if route.Active && route.PraxeonID == praxeonID {
			return true
		}
	}
	return false
}

func (e *Engine) emit(eventType EventType, subjectID, praxeonID, agentID string, visibility EventVisibility, payloadHash string, now time.Time) {
	e.eventCursor++
	e.events = append(e.events, Event{
		Cursor:      e.eventCursor,
		ID:          fmt.Sprintf("event:%d", e.eventCursor),
		Type:        eventType,
		SubjectID:   subjectID,
		PraxeonID:   praxeonID,
		AgentID:     agentID,
		Visibility:  visibility,
		PayloadHash: payloadHash,
		OccurredAt:  canonicalTime(now),
	})
}

func requiredApproval(proposal Proposal) ApprovalRule {
	if proposal.Type == ProposalTypeCharter || proposal.AffectsGovernance || proposal.ReducesRights || proposal.Risk >= RiskLevelHigh || proposal.ProcedureKind == "close_debate" {
		return ApprovalRuleTwoThirds
	}
	if proposal.Type == ProposalTypeRevoke && proposal.TargetApproval != ApprovalRuleUnspecified {
		return proposal.TargetApproval
	}
	return ApprovalRuleSimpleMajority
}

func approved(yes, no uint32, rule ApprovalRule) bool {
	if yes+no == 0 {
		return false
	}
	switch rule {
	case ApprovalRuleSimpleMajority:
		return yes > no
	case ApprovalRuleTwoThirds:
		return yes*3 >= (yes+no)*2
	default:
		return false
	}
}

func validateProposal(proposal Proposal) error {
	if proposal.ID == "" || proposal.Hash == "" || proposal.Title == "" || proposal.BodyHash == "" || proposal.ProposerPraxeonID == "" || proposal.Domain == "" {
		return ErrInvalidInput
	}
	if proposal.Type < ProposalTypeRule || proposal.Type > ProposalTypeProcedure || proposal.Risk < RiskLevelLow || proposal.Risk > RiskLevelCritical {
		return ErrInvalidInput
	}
	if !proposal.Schedule.NoticeOpensAt.Before(proposal.Schedule.BallotOpensAt) || !proposal.Schedule.BallotOpensAt.Before(proposal.Schedule.BallotClosesAt) {
		return ErrInvalidInput
	}
	return nil
}

func validateCivilizationGoal(spec CivilizationGoalSpec) error {
	if spec.ID == "" || spec.Title == "" {
		return ErrInvalidInput
	}
	return validateGoalChecks(spec.SuccessChecks)
}

func validateGoalChecks(checks []GoalCheck) error {
	if len(checks) == 0 {
		return ErrInvalidInput
	}
	var total uint32
	seen := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		if check.ID == "" || check.Description == "" || check.WeightBPS == 0 {
			return ErrInvalidInput
		}
		if _, duplicate := seen[check.ID]; duplicate {
			return ErrInvalidInput
		}
		seen[check.ID] = struct{}{}
		total += check.WeightBPS
	}
	if total != 10_000 {
		return ErrInvalidInput
	}
	return nil
}

func activeRoutes(routes []PraxeonRoute) []PraxeonRoute {
	active := make([]PraxeonRoute, 0, len(routes))
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		if !route.Active || route.PraxeonID == "" {
			continue
		}
		if _, duplicate := seen[route.PraxeonID]; duplicate {
			continue
		}
		seen[route.PraxeonID] = struct{}{}
		active = append(active, route)
	}
	sort.Slice(active, func(i, j int) bool { return active[i].PraxeonID < active[j].PraxeonID })
	return active
}

func goalVisibleToAgent(goal Goal, agentID string) bool {
	return goal.Type == GoalTypeCivilization || contains(goal.AssignedAgentIDs, agentID)
}

func eventVisibleToAgent(event Event, praxeonID, agentID string) bool {
	switch event.Visibility {
	case EventVisibilityPublic:
		return true
	case EventVisibilityPraxeon:
		return event.PraxeonID == praxeonID
	case EventVisibilityAgent:
		return event.AgentID == agentID
	default:
		return false
	}
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func containsProposalType(values []ProposalType, value ProposalType) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func canonicalTime(value time.Time) time.Time {
	return value.UTC().Round(0)
}

func canonicalSchedule(value Schedule) Schedule {
	return Schedule{
		NoticeOpensAt:  canonicalTime(value.NoticeOpensAt),
		BallotOpensAt:  canonicalTime(value.BallotOpensAt),
		BallotClosesAt: canonicalTime(value.BallotClosesAt),
	}
}

func voteRequestID(proposalID string, version uint64, praxeonID string) string {
	return fmt.Sprintf("request:%s:%d:%s", proposalID, version, praxeonID)
}

func eligibilitySnapshotHash(praxeonIDs []string) string {
	digest := sha256.Sum256([]byte(strings.Join(praxeonIDs, "\x00")))
	return fmt.Sprintf("%x", digest[:])
}

func voteAuthorizationHash(authorization VoteAuthorization) string {
	proposalTypes := make([]int, 0, len(authorization.AllowedProposalTypes))
	for _, proposalType := range authorization.AllowedProposalTypes {
		proposalTypes = append(proposalTypes, int(proposalType))
	}
	sort.Ints(proposalTypes)
	domains := append([]string(nil), authorization.AllowedDomains...)
	proposalIDs := append([]string(nil), authorization.AllowedProposalIDs...)
	sort.Strings(domains)
	sort.Strings(proposalIDs)

	parts := []string{
		authorization.ID,
		authorization.PraxeonID,
		authorization.AgentID,
		fmt.Sprintf("mode:%d", authorization.Mode),
		fmt.Sprintf("risk:%d", authorization.MaxRisk),
		fmt.Sprintf("from:%d", canonicalTime(authorization.ValidFrom).UnixNano()),
		fmt.Sprintf("expires:%d", canonicalTime(authorization.ExpiresAt).UnixNano()),
		"policy:" + authorization.DecisionPolicyHash,
		fmt.Sprintf("abstain:%t", authorization.CanAbstain),
		fmt.Sprintf("reason:%t", authorization.RequiresReason),
	}
	for _, proposalType := range proposalTypes {
		parts = append(parts, fmt.Sprintf("type:%d", proposalType))
	}
	for _, domain := range domains {
		parts = append(parts, "domain:"+domain)
	}
	for _, proposalID := range proposalIDs {
		parts = append(parts, "proposal:"+proposalID)
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return fmt.Sprintf("%x", digest[:])
}

func ballotKey(proposalID string, version uint64, praxeonID string) string {
	return fmt.Sprintf("%s\x00%d\x00%s", proposalID, version, praxeonID)
}

func proposalVersionKey(proposalID string, version uint64) string {
	return fmt.Sprintf("%s\x00%d", proposalID, version)
}

func cloneGoal(value Goal) Goal {
	value.AssignedAgentIDs = append([]string(nil), value.AssignedAgentIDs...)
	value.SuccessChecks = cloneGoalChecks(value.SuccessChecks)
	return value
}

func cloneGoalChecks(values []GoalCheck) []GoalCheck {
	result := make([]GoalCheck, len(values))
	for i, value := range values {
		value.SourceResultIDs = append([]string(nil), value.SourceResultIDs...)
		result[i] = value
	}
	return result
}

func cloneProposal(value Proposal) Proposal {
	value.EligiblePraxeonIDs = append([]string(nil), value.EligiblePraxeonIDs...)
	if value.CivilizationGoalSpec != nil {
		copySpec := *value.CivilizationGoalSpec
		copySpec.AssignedAgentIDs = append([]string(nil), copySpec.AssignedAgentIDs...)
		copySpec.SuccessChecks = cloneGoalChecks(copySpec.SuccessChecks)
		value.CivilizationGoalSpec = &copySpec
	}
	return value
}

func cloneAuthorization(value VoteAuthorization) VoteAuthorization {
	value.AllowedProposalTypes = append([]ProposalType(nil), value.AllowedProposalTypes...)
	value.AllowedDomains = append([]string(nil), value.AllowedDomains...)
	value.AllowedProposalIDs = append([]string(nil), value.AllowedProposalIDs...)
	return value
}
