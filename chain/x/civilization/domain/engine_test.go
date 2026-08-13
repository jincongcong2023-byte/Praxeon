package domain_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"mars/x/civilization/domain"
)

var baseTime = time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)

type fakeDirectory struct {
	routes []domain.PraxeonRoute
	owners map[string]string
}

func (d fakeDirectory) Routes() []domain.PraxeonRoute {
	return append([]domain.PraxeonRoute(nil), d.routes...)
}

func (d fakeDirectory) PraxeonForAgent(agentID string) (string, bool) {
	owner, ok := d.owners[agentID]
	return owner, ok
}

func directory(size int) fakeDirectory {
	routes := make([]domain.PraxeonRoute, 0, size)
	owners := make(map[string]string, size*2)
	for i := 1; i <= size; i++ {
		praxeonID := fmt.Sprintf("PX-%02d", i)
		primaryID := fmt.Sprintf("agent-%02d-primary", i)
		governanceID := fmt.Sprintf("agent-%02d-governance", i)
		route := domain.PraxeonRoute{
			PraxeonID:      praxeonID,
			PrimaryAgentID: primaryID,
			Active:         true,
		}
		if i%2 == 1 {
			route.GovernanceAgentID = governanceID
			owners[governanceID] = praxeonID
		}
		routes = append(routes, route)
		owners[primaryID] = praxeonID
	}
	return fakeDirectory{routes: routes, owners: owners}
}

func proposal(id string, proposalType domain.ProposalType, risk domain.RiskLevel) domain.Proposal {
	return domain.Proposal{
		ID:                id,
		Hash:              "hash-" + id + "-v1",
		Type:              proposalType,
		Title:             "Proposal " + id,
		Summary:           "Summary " + id,
		BodyHash:          "body-" + id + "-v1",
		ProposerPraxeonID: "PX-01",
		Domain:            "infrastructure",
		Risk:              risk,
		Schedule: domain.Schedule{
			NoticeOpensAt:  baseTime.Add(time.Minute),
			BallotOpensAt:  baseTime.Add(2 * time.Minute),
			BallotClosesAt: baseTime.Add(5 * time.Minute),
		},
	}
}

func goalChecks() []domain.GoalCheck {
	return []domain.GoalCheck{
		{ID: "check-1", Description: "Verified capacity is available", WeightBPS: 6_000},
		{ID: "check-2", Description: "Failure recovery is tested", WeightBPS: 4_000},
	}
}

func fileAndSecond(t *testing.T, engine *domain.Engine, input domain.Proposal) domain.Proposal {
	t.Helper()
	filed, err := engine.FileProposal("PX-01", input, baseTime)
	if err != nil {
		t.Fatalf("file proposal: %v", err)
	}
	noticed, err := engine.SecondProposal("PX-02", filed.ID, filed.Version, baseTime.Add(30*time.Second))
	if err != nil {
		t.Fatalf("second proposal: %v", err)
	}
	return noticed
}

func directBallot(t *testing.T, engine *domain.Engine, proposalID string, version uint64, voter string, choice domain.BallotChoice) domain.Ballot {
	t.Helper()
	ballot, err := engine.SubmitBallot(voter, domain.Ballot{
		ProposalID:      proposalID,
		ProposalVersion: version,
		VoterPraxeonID:  voter,
		Choice:          choice,
		SubmittedBy:     domain.SubmitterTypePraxeon,
	}, 0, baseTime.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("direct ballot for %s: %v", voter, err)
	}
	return ballot
}

func authorization(id, praxeonID, agentID string, mode domain.VoteMode, proposalType domain.ProposalType, risk domain.RiskLevel) domain.VoteAuthorization {
	result := domain.VoteAuthorization{
		ID:                   id,
		PraxeonID:            praxeonID,
		AgentID:              agentID,
		Mode:                 mode,
		AllowedProposalTypes: []domain.ProposalType{proposalType},
		AllowedDomains:       []string{"infrastructure"},
		MaxRisk:              risk,
		ValidFrom:            baseTime,
		ExpiresAt:            baseTime.Add(5 * time.Minute),
		CanAbstain:           true,
	}
	if mode == domain.VoteModePolicyVote {
		result.DecisionPolicyHash = "policy-hash"
	}
	return result
}

func TestSecondProposalAutomaticallyDistributesVoteRequests(t *testing.T) {
	dir := directory(3)
	engine := domain.NewEngine(dir)
	noticed := fileAndSecond(t, engine, proposal("P-001", domain.ProposalTypeRule, domain.RiskLevelMedium))

	if got, want := len(noticed.EligiblePraxeonIDs), 3; got != want {
		t.Fatalf("eligible count: got %d want %d", got, want)
	}
	if noticed.EligibilityHash == "" || noticed.EligibilityFrozenAt.IsZero() {
		t.Fatalf("eligible snapshot was not anchored: %+v", noticed)
	}
	request1, err := engine.VoteRequest(noticed.ID, noticed.Version, "PX-01")
	if err != nil {
		t.Fatalf("request PX-01: %v", err)
	}
	if request1.RecipientAgentID != "agent-01-governance" {
		t.Fatalf("designated governance route not used: %q", request1.RecipientAgentID)
	}
	if request1.DeliveryStatus != domain.VoteRequestDeliveryStatusDelivered || request1.DeliveredAt.IsZero() {
		t.Fatalf("routed request was not marked delivered: %+v", request1)
	}
	request2, err := engine.VoteRequest(noticed.ID, noticed.Version, "PX-02")
	if err != nil {
		t.Fatalf("request PX-02: %v", err)
	}
	if request2.RecipientAgentID != "agent-02-primary" {
		t.Fatalf("primary fallback not used: %q", request2.RecipientAgentID)
	}

	snapshot, err := engine.AgentSnapshot("agent-01-governance", baseTime.Add(90*time.Second))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.VoteRequests) != 1 || snapshot.VoteRequests[0].ID != request1.ID {
		t.Fatalf("snapshot request mismatch: %+v", snapshot.VoteRequests)
	}
	if len(snapshot.RequiredConfirmationRequestIDs) != 1 || snapshot.RequiredConfirmationRequestIDs[0] != request1.ID {
		t.Fatalf("default path must require confirmation: %+v", snapshot.RequiredConfirmationRequestIDs)
	}
}

func TestMissingAgentRouteIsVisibleWithoutChangingEligibility(t *testing.T) {
	dir := directory(3)
	dir.routes[1].PrimaryAgentID = ""
	engine := domain.NewEngine(dir)
	noticed := fileAndSecond(t, engine, proposal("P-NO-ROUTE", domain.ProposalTypeRule, domain.RiskLevelMedium))

	if len(noticed.EligiblePraxeonIDs) != 3 {
		t.Fatalf("delivery failure changed eligible snapshot: %+v", noticed.EligiblePraxeonIDs)
	}
	request, err := engine.VoteRequest(noticed.ID, noticed.Version, "PX-02")
	if err != nil {
		t.Fatalf("failed request record: %v", err)
	}
	if request.DeliveryStatus != domain.VoteRequestDeliveryStatusFailed || request.DeliveryFailure != "no_active_agent_route" || !request.DeliveredAt.IsZero() {
		t.Fatalf("missing route was not represented truthfully: %+v", request)
	}
}

func TestAgentAcknowledgesExactProposalHashIdempotently(t *testing.T) {
	engine := domain.NewEngine(directory(3))
	noticed := fileAndSecond(t, engine, proposal("P-ACK", domain.ProposalTypeRule, domain.RiskLevelMedium))
	request, err := engine.VoteRequest(noticed.ID, noticed.Version, "PX-01")
	if err != nil {
		t.Fatalf("vote request: %v", err)
	}
	if _, err := engine.AcknowledgeVoteRequest(request.RecipientAgentID, request.ID, "wrong-hash", baseTime.Add(time.Minute)); !errors.Is(err, domain.ErrProposalHashMismatch) {
		t.Fatalf("expected exact hash rejection, got %v", err)
	}
	acknowledged, err := engine.AcknowledgeVoteRequest(request.RecipientAgentID, request.ID, noticed.Hash, baseTime.Add(time.Minute))
	if err != nil {
		t.Fatalf("acknowledge request: %v", err)
	}
	if acknowledged.AcknowledgedAt.IsZero() {
		t.Fatal("acknowledgement time was not recorded")
	}
	again, err := engine.AcknowledgeVoteRequest(request.RecipientAgentID, request.ID, noticed.Hash, baseTime.Add(90*time.Second))
	if err != nil {
		t.Fatalf("idempotent acknowledgement: %v", err)
	}
	if !again.AcknowledgedAt.Equal(acknowledged.AcknowledgedAt) {
		t.Fatalf("idempotent acknowledgement changed the first receipt time: first=%v again=%v", acknowledged.AcknowledgedAt, again.AcknowledgedAt)
	}
}

func TestEventCursorVisibilityAndResume(t *testing.T) {
	dir := directory(3)
	engine := domain.NewEngine(dir)
	noticed := fileAndSecond(t, engine, proposal("P-EVENTS", domain.ProposalTypeRule, domain.RiskLevelMedium))

	grant := authorization("AUTH-EVENTS", "PX-01", "agent-01-governance", domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelMedium)
	if _, err := engine.SetVoteAuthorization("PX-01", grant, baseTime); err != nil {
		t.Fatalf("set authorization: %v", err)
	}

	ownerEvents, ownerCursor, err := engine.Events("agent-01-governance", 0, 100)
	if err != nil {
		t.Fatalf("owner events: %v", err)
	}
	if len(ownerEvents) != 3 || ownerEvents[2].Type != domain.EventTypeVoteAuthorizationChanged {
		t.Fatalf("owner must see two public proposal events and its private grant: %+v", ownerEvents)
	}

	otherEvents, otherCursor, err := engine.Events("agent-02-primary", 0, 100)
	if err != nil {
		t.Fatalf("other events: %v", err)
	}
	if len(otherEvents) != 2 {
		t.Fatalf("another Praxeon must not see the private authorization event: %+v", otherEvents)
	}
	if otherCursor != ownerCursor {
		t.Fatalf("filtered stream did not advance over the private event: owner=%d other=%d", ownerCursor, otherCursor)
	}
	for _, event := range otherEvents {
		if event.Type == domain.EventTypeVoteAuthorizationChanged {
			t.Fatal("private authorization leaked into another Praxeon's event stream")
		}
	}

	directBallot(t, engine, noticed.ID, noticed.Version, "PX-02", domain.BallotChoiceYes)
	resumed, resumedCursor, err := engine.Events("agent-01-governance", ownerCursor, 100)
	if err != nil {
		t.Fatalf("resumed events: %v", err)
	}
	if len(resumed) != 1 || resumed[0].Type != domain.EventTypeBallotSubmitted || resumedCursor <= ownerCursor {
		t.Fatalf("cursor did not resume exactly once: events=%+v cursor=%d previous=%d", resumed, resumedCursor, ownerCursor)
	}
}

func TestDefaultConfirmEachUsesExactConfirmationAsAuthorization(t *testing.T) {
	engine := domain.NewEngine(directory(3))
	noticed := fileAndSecond(t, engine, proposal("P-DEFAULT-CONFIRM", domain.ProposalTypeRule, domain.RiskLevelMedium))
	agentID := "agent-01-governance"
	input := domain.Ballot{
		ProposalID: noticed.ID, ProposalVersion: noticed.Version, VoterPraxeonID: "PX-01",
		Choice: domain.BallotChoiceYes, SubmittedBy: domain.SubmitterTypeAgent,
	}
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrConfirmationRequired) {
		t.Fatalf("default Agent path did not require Praxeon confirmation: %v", err)
	}
	confirmation, err := engine.ConfirmBallot("PX-01", domain.BallotConfirmation{
		ID: "CONFIRM-DEFAULT", PraxeonID: "PX-01", AgentID: agentID, ProposalID: noticed.ID,
		ProposalVersion: noticed.Version, Choice: input.Choice, ExpiresAt: baseTime.Add(4 * time.Minute),
	}, baseTime.Add(2*time.Minute+30*time.Second))
	if err != nil {
		t.Fatalf("confirm default ballot: %v", err)
	}
	input.ConfirmationID = confirmation.ID
	ballot, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("default confirmed ballot: %v", err)
	}
	if ballot.AuthorizationID != "" || ballot.ConfirmationID != confirmation.ID {
		t.Fatalf("single confirmation was not used as the exact authorization: %+v", ballot)
	}
}

func TestConfirmEachAndPraxeonOverrideLock(t *testing.T) {
	dir := directory(3)
	engine := domain.NewEngine(dir)
	noticed := fileAndSecond(t, engine, proposal("P-002", domain.ProposalTypeRule, domain.RiskLevelMedium))
	agentID := "agent-01-governance"

	auth, err := engine.SetVoteAuthorization("PX-01", authorization("AUTH-1", "PX-01", agentID, domain.VoteModeConfirmEach, noticed.Type, domain.RiskLevelMedium), baseTime)
	if err != nil {
		t.Fatalf("set authorization: %v", err)
	}
	if auth.Hash == "" {
		t.Fatal("authorization hash was not derived by the state machine")
	}
	agentInput := domain.Ballot{
		ProposalID:      noticed.ID,
		ProposalVersion: noticed.Version,
		VoterPraxeonID:  "PX-01",
		Choice:          domain.BallotChoiceYes,
		SubmittedBy:     domain.SubmitterTypeAgent,
		AuthorizationID: auth.ID,
	}
	if _, err := engine.SubmitBallot(agentID, agentInput, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrConfirmationRequired) {
		t.Fatalf("expected confirmation required, got %v", err)
	}

	confirmation, err := engine.ConfirmBallot("PX-01", domain.BallotConfirmation{
		ID:              "CONFIRM-1",
		PraxeonID:       "PX-01",
		AgentID:         agentID,
		ProposalID:      noticed.ID,
		ProposalVersion: noticed.Version,
		Choice:          domain.BallotChoiceYes,
		ExpiresAt:       baseTime.Add(4 * time.Minute),
	}, baseTime.Add(150*time.Second))
	if err != nil {
		t.Fatalf("confirm ballot: %v", err)
	}
	agentInput.ConfirmationID = confirmation.ID
	agentBallot, err := engine.SubmitBallot(agentID, agentInput, 0, baseTime.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("agent ballot: %v", err)
	}
	if agentBallot.Sequence != 1 || agentBallot.PraxeonLock {
		t.Fatalf("unexpected agent ballot: %+v", agentBallot)
	}
	if agentBallot.ProposalHash != noticed.Hash || agentBallot.AuthorizationHash != auth.Hash {
		t.Fatalf("ballot did not anchor proposal and authorization hashes: %+v", agentBallot)
	}

	direct, err := engine.SubmitBallot("PX-01", domain.Ballot{
		ProposalID:      noticed.ID,
		ProposalVersion: noticed.Version,
		VoterPraxeonID:  "PX-01",
		Choice:          domain.BallotChoiceNo,
		SubmittedBy:     domain.SubmitterTypePraxeon,
	}, 1, baseTime.Add(210*time.Second))
	if err != nil {
		t.Fatalf("direct override: %v", err)
	}
	if direct.Sequence != 2 || !direct.PraxeonLock || direct.Choice != domain.BallotChoiceNo {
		t.Fatalf("direct override did not lock: %+v", direct)
	}

	secondConfirmation, err := engine.ConfirmBallot("PX-01", domain.BallotConfirmation{
		ID:              "CONFIRM-2",
		PraxeonID:       "PX-01",
		AgentID:         agentID,
		ProposalID:      noticed.ID,
		ProposalVersion: noticed.Version,
		Choice:          domain.BallotChoiceYes,
		ExpiresAt:       baseTime.Add(4*time.Minute + 30*time.Second),
	}, baseTime.Add(220*time.Second))
	if err != nil {
		t.Fatalf("second confirmation: %v", err)
	}
	agentInput.ConfirmationID = secondConfirmation.ID
	if _, err := engine.SubmitBallot(agentID, agentInput, 2, baseTime.Add(4*time.Minute)); !errors.Is(err, domain.ErrPraxeonLocked) {
		t.Fatalf("expected praxeon lock, got %v", err)
	}
}

func TestNotifyOnlyAndConfirmationBinding(t *testing.T) {
	engine := domain.NewEngine(directory(3))
	noticed := fileAndSecond(t, engine, proposal("P-AUTH-MODES", domain.ProposalTypeRule, domain.RiskLevelMedium))
	agentID := "agent-01-governance"

	notify, err := engine.SetVoteAuthorization("PX-01", authorization("AUTH-NOTIFY", "PX-01", agentID, domain.VoteModeNotifyOnly, noticed.Type, domain.RiskLevelMedium), baseTime)
	if err != nil {
		t.Fatalf("notify authorization: %v", err)
	}
	input := domain.Ballot{
		ProposalID: noticed.ID, ProposalVersion: noticed.Version, VoterPraxeonID: "PX-01",
		Choice: domain.BallotChoiceYes, SubmittedBy: domain.SubmitterTypeAgent, AuthorizationID: notify.ID,
	}
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrNotifyOnly) {
		t.Fatalf("notify-only Agent submitted a ballot: %v", err)
	}

	confirm, err := engine.SetVoteAuthorization("PX-01", authorization("AUTH-CONFIRM", "PX-01", agentID, domain.VoteModeConfirmEach, noticed.Type, domain.RiskLevelMedium), baseTime)
	if err != nil {
		t.Fatalf("confirm authorization: %v", err)
	}
	approval, err := engine.ConfirmBallot("PX-01", domain.BallotConfirmation{
		ID: "CONFIRM-BOUND", PraxeonID: "PX-01", AgentID: agentID, ProposalID: noticed.ID,
		ProposalVersion: noticed.Version, Choice: domain.BallotChoiceYes, ExpiresAt: baseTime.Add(4 * time.Minute),
	}, baseTime.Add(2*time.Minute+30*time.Second))
	if err != nil {
		t.Fatalf("confirmation: %v", err)
	}
	input.AuthorizationID = confirm.ID
	input.ConfirmationID = approval.ID
	input.Choice = domain.BallotChoiceNo
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrConfirmationInvalid) {
		t.Fatalf("confirmation was not bound to its exact choice: %v", err)
	}
	input.Choice = domain.BallotChoiceYes
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); err != nil {
		t.Fatalf("confirmed ballot: %v", err)
	}
	if _, err := engine.SubmitBallot(agentID, input, 1, baseTime.Add(3*time.Minute+time.Second)); !errors.Is(err, domain.ErrConfirmationInvalid) {
		t.Fatalf("one-time confirmation was reused: %v", err)
	}
}

func TestMultipleAgentsStillProduceOnePraxeonBallot(t *testing.T) {
	dir := directory(3)
	dir.owners["agent-01-secondary"] = "PX-01"
	engine := domain.NewEngine(dir)
	noticed := fileAndSecond(t, engine, proposal("P-003", domain.ProposalTypeRule, domain.RiskLevelMedium))

	auth1, err := engine.SetVoteAuthorization("PX-01", authorization("AUTH-A", "PX-01", "agent-01-governance", domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelMedium), baseTime)
	if err != nil {
		t.Fatalf("auth 1: %v", err)
	}
	auth2, err := engine.SetVoteAuthorization("PX-01", authorization("AUTH-B", "PX-01", "agent-01-secondary", domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelMedium), baseTime)
	if err != nil {
		t.Fatalf("auth 2: %v", err)
	}

	first, err := engine.SubmitBallot("agent-01-governance", domain.Ballot{
		ProposalID: noticed.ID, ProposalVersion: noticed.Version, VoterPraxeonID: "PX-01",
		Choice: domain.BallotChoiceYes, SubmittedBy: domain.SubmitterTypeAgent, AuthorizationID: auth1.ID,
	}, 0, baseTime.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("first agent ballot: %v", err)
	}
	second, err := engine.SubmitBallot("agent-01-secondary", domain.Ballot{
		ProposalID: noticed.ID, ProposalVersion: noticed.Version, VoterPraxeonID: "PX-01",
		Choice: domain.BallotChoiceNo, SubmittedBy: domain.SubmitterTypeAgent, AuthorizationID: auth2.ID,
	}, first.Sequence, baseTime.Add(3*time.Minute+time.Second))
	if err != nil {
		t.Fatalf("second agent update: %v", err)
	}
	if second.Sequence != 2 || second.Choice != domain.BallotChoiceNo {
		t.Fatalf("expected one updated ballot, got %+v", second)
	}
	current, err := engine.Ballot(noticed.ID, noticed.Version, "PX-01")
	if err != nil {
		t.Fatalf("current ballot: %v", err)
	}
	if current.ID != second.ID {
		t.Fatalf("more than one current ballot was exposed: %+v", current)
	}
}

func TestAuthorizationScopeExpiryAndRevocation(t *testing.T) {
	dir := directory(3)
	engine := domain.NewEngine(dir)
	noticed := fileAndSecond(t, engine, proposal("P-004", domain.ProposalTypeRule, domain.RiskLevelMedium))
	agentID := "agent-01-governance"

	expired := authorization("AUTH-EXPIRED", "PX-01", agentID, domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelMedium)
	expired.ExpiresAt = baseTime.Add(150 * time.Second)
	if _, err := engine.SetVoteAuthorization("PX-01", expired, baseTime); err != nil {
		t.Fatalf("set expired authorization: %v", err)
	}
	input := domain.Ballot{
		ProposalID: noticed.ID, ProposalVersion: noticed.Version, VoterPraxeonID: "PX-01",
		Choice: domain.BallotChoiceYes, SubmittedBy: domain.SubmitterTypeAgent, AuthorizationID: expired.ID,
	}
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrAuthorizationExpired) {
		t.Fatalf("expected expired authorization, got %v", err)
	}

	wrongDomain := authorization("AUTH-WRONG-DOMAIN", "PX-01", agentID, domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelMedium)
	wrongDomain.AllowedDomains = []string{"governance"}
	if _, err := engine.SetVoteAuthorization("PX-01", wrongDomain, baseTime); err != nil {
		t.Fatalf("set wrong-domain authorization: %v", err)
	}
	input.AuthorizationID = wrongDomain.ID
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrAuthorizationInvalid) {
		t.Fatalf("expected invalid scope, got %v", err)
	}

	wrongProposal := authorization("AUTH-WRONG-PROPOSAL", "PX-01", agentID, domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelMedium)
	wrongProposal.AllowedProposalIDs = []string{"P-SOMEWHERE-ELSE"}
	if _, err := engine.SetVoteAuthorization("PX-01", wrongProposal, baseTime); err != nil {
		t.Fatalf("set proposal-scoped authorization: %v", err)
	}
	input.AuthorizationID = wrongProposal.ID
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrAuthorizationInvalid) {
		t.Fatalf("expected invalid proposal scope, got %v", err)
	}

	revoked := authorization("AUTH-REVOKED", "PX-01", agentID, domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelMedium)
	if _, err := engine.SetVoteAuthorization("PX-01", revoked, baseTime); err != nil {
		t.Fatalf("set revocable authorization: %v", err)
	}
	if err := engine.RevokeVoteAuthorization("PX-01", revoked.ID, baseTime.Add(time.Minute)); err != nil {
		t.Fatalf("revoke authorization: %v", err)
	}
	input.AuthorizationID = revoked.ID
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrAuthorizationExpired) {
		t.Fatalf("expected revoked authorization, got %v", err)
	}
}

func TestAuthorizationHashCanonicalizesScopeOrdering(t *testing.T) {
	first := authorization("AUTH-CANONICAL", "PX-01", "agent-01-governance", domain.VoteModePolicyVote, domain.ProposalTypeRule, domain.RiskLevelHigh)
	first.AllowedProposalTypes = []domain.ProposalType{domain.ProposalTypeRule, domain.ProposalTypeCharter}
	first.AllowedDomains = []string{"infrastructure", "governance"}
	first.AllowedProposalIDs = []string{"P-2", "P-1"}
	second := first
	second.AllowedProposalTypes = []domain.ProposalType{domain.ProposalTypeCharter, domain.ProposalTypeRule}
	second.AllowedDomains = []string{"governance", "infrastructure"}
	second.AllowedProposalIDs = []string{"P-1", "P-2"}

	storedFirst, err := domain.NewEngine(directory(3)).SetVoteAuthorization("PX-01", first, baseTime)
	if err != nil {
		t.Fatalf("first authorization: %v", err)
	}
	storedSecond, err := domain.NewEngine(directory(3)).SetVoteAuthorization("PX-01", second, baseTime)
	if err != nil {
		t.Fatalf("second authorization: %v", err)
	}
	if storedFirst.Hash == "" || storedFirst.Hash != storedSecond.Hash {
		t.Fatalf("equivalent scope order produced different hashes: first=%s second=%s", storedFirst.Hash, storedSecond.Hash)
	}
}

func TestHighRiskPolicyVoteRequiresProposalSpecificAuthorization(t *testing.T) {
	dir := directory(3)
	engine := domain.NewEngine(dir)
	noticed := fileAndSecond(t, engine, proposal("P-HIGH", domain.ProposalTypeRule, domain.RiskLevelHigh))
	agentID := "agent-01-governance"

	broad := authorization("AUTH-HIGH-BROAD", "PX-01", agentID, domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelHigh)
	if _, err := engine.SetVoteAuthorization("PX-01", broad, baseTime); err != nil {
		t.Fatalf("set broad authorization: %v", err)
	}
	input := domain.Ballot{
		ProposalID: noticed.ID, ProposalVersion: noticed.Version, VoterPraxeonID: "PX-01",
		Choice: domain.BallotChoiceYes, SubmittedBy: domain.SubmitterTypeAgent, AuthorizationID: broad.ID,
	}
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); !errors.Is(err, domain.ErrAuthorizationInvalid) {
		t.Fatalf("expected proposal-specific authorization, got %v", err)
	}

	dedicated := authorization("AUTH-HIGH-DEDICATED", "PX-01", agentID, domain.VoteModePolicyVote, noticed.Type, domain.RiskLevelHigh)
	dedicated.AllowedProposalIDs = []string{noticed.ID}
	if _, err := engine.SetVoteAuthorization("PX-01", dedicated, baseTime); err != nil {
		t.Fatalf("set dedicated authorization: %v", err)
	}
	input.AuthorizationID = dedicated.ID
	if _, err := engine.SubmitBallot(agentID, input, 0, baseTime.Add(3*time.Minute)); err != nil {
		t.Fatalf("dedicated high-risk ballot: %v", err)
	}
}

func TestCountingRulesAndQuorum(t *testing.T) {
	t.Run("simple majority with quorum", func(t *testing.T) {
		engine := domain.NewEngine(directory(5))
		noticed := fileAndSecond(t, engine, proposal("P-SIMPLE", domain.ProposalTypeRule, domain.RiskLevelMedium))
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-01", domain.BallotChoiceYes)
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-02", domain.BallotChoiceYes)
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-03", domain.BallotChoiceNo)

		decision, err := engine.FinalizeProposal(noticed.ID, noticed.Version, baseTime.Add(6*time.Minute))
		if err != nil {
			t.Fatalf("finalize: %v", err)
		}
		if decision.Outcome != domain.ProposalStatusRatified || decision.Yes != 2 || decision.No != 1 || decision.NotVoted != 2 || decision.Delivered != 5 {
			t.Fatalf("unexpected decision: %+v", decision)
		}
	})

	t.Run("no quorum is not rejected", func(t *testing.T) {
		engine := domain.NewEngine(directory(5))
		noticed := fileAndSecond(t, engine, proposal("P-NO-QUORUM", domain.ProposalTypeRule, domain.RiskLevelMedium))
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-01", domain.BallotChoiceYes)
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-02", domain.BallotChoiceYes)

		decision, err := engine.FinalizeProposal(noticed.ID, noticed.Version, baseTime.Add(6*time.Minute))
		if err != nil {
			t.Fatalf("finalize: %v", err)
		}
		if decision.Outcome != domain.ProposalStatusExpiredNoQuorum {
			t.Fatalf("expected no quorum, got %+v", decision)
		}
	})

	t.Run("two thirds uses exact integer boundary", func(t *testing.T) {
		engine := domain.NewEngine(directory(5))
		input := proposal("P-CHARTER", domain.ProposalTypeCharter, domain.RiskLevelMedium)
		noticed := fileAndSecond(t, engine, input)
		if noticed.Approval != domain.ApprovalRuleTwoThirds {
			t.Fatalf("charter must use two thirds: %+v", noticed)
		}
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-01", domain.BallotChoiceYes)
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-02", domain.BallotChoiceYes)
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-03", domain.BallotChoiceNo)

		decision, err := engine.FinalizeProposal(noticed.ID, noticed.Version, baseTime.Add(6*time.Minute))
		if err != nil {
			t.Fatalf("finalize: %v", err)
		}
		if decision.Outcome != domain.ProposalStatusRatified {
			t.Fatalf("2/3 exact boundary must pass: %+v", decision)
		}
	})

	t.Run("abstain counts for quorum but not approval denominator", func(t *testing.T) {
		engine := domain.NewEngine(directory(5))
		noticed := fileAndSecond(t, engine, proposal("P-ABSTAIN", domain.ProposalTypeRule, domain.RiskLevelMedium))
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-01", domain.BallotChoiceYes)
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-02", domain.BallotChoiceAbstain)
		directBallot(t, engine, noticed.ID, noticed.Version, "PX-03", domain.BallotChoiceAbstain)

		decision, err := engine.FinalizeProposal(noticed.ID, noticed.Version, baseTime.Add(6*time.Minute))
		if err != nil {
			t.Fatalf("finalize: %v", err)
		}
		if decision.Outcome != domain.ProposalStatusRatified || decision.Yes != 1 || decision.Abstain != 2 {
			t.Fatalf("abstain was counted in the wrong denominator: %+v", decision)
		}
	})
}

func TestRatifiedCivilizationGoalAppearsInAgentSnapshot(t *testing.T) {
	dir := directory(5)
	engine := domain.NewEngine(dir)
	input := proposal("P-GOAL", domain.ProposalTypeCivilizationGoal, domain.RiskLevelMedium)
	input.CivilizationGoalSpec = &domain.CivilizationGoalSpec{
		ID:              "GOAL-CIV-1",
		Title:           "Increase verified energy reserve",
		DescriptionHash: "goal-description-hash",
		SuccessChecks:   goalChecks(),
		TargetAt:        baseTime.Add(30 * 24 * time.Hour),
	}
	noticed := fileAndSecond(t, engine, input)
	directBallot(t, engine, noticed.ID, noticed.Version, "PX-01", domain.BallotChoiceYes)
	directBallot(t, engine, noticed.ID, noticed.Version, "PX-02", domain.BallotChoiceYes)
	directBallot(t, engine, noticed.ID, noticed.Version, "PX-03", domain.BallotChoiceYes)

	decision, err := engine.FinalizeProposal(noticed.ID, noticed.Version, baseTime.Add(6*time.Minute))
	if err != nil {
		t.Fatalf("finalize goal proposal: %v", err)
	}
	if decision.Outcome != domain.ProposalStatusRatified {
		t.Fatalf("goal proposal not ratified: %+v", decision)
	}
	goal, err := engine.Goal("GOAL-CIV-1")
	if err != nil {
		t.Fatalf("materialized goal: %v", err)
	}
	if goal.Type != domain.GoalTypeCivilization || goal.GovernanceProposalID != noticed.ID || goal.Status != domain.GoalStatusActive {
		t.Fatalf("unexpected civilization goal: %+v", goal)
	}
	snapshot, err := engine.AgentSnapshot("agent-04-primary", baseTime.Add(7*time.Minute))
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.ActiveGoals) != 1 || snapshot.ActiveGoals[0].ID != goal.ID {
		t.Fatalf("civilization goal missing from snapshot: %+v", snapshot.ActiveGoals)
	}
}

func TestMaterialAmendmentInvalidatesAndRedistributes(t *testing.T) {
	dir := directory(3)
	engine := domain.NewEngine(dir)
	current := fileAndSecond(t, engine, proposal("P-AMEND", domain.ProposalTypeRule, domain.RiskLevelMedium))
	directBallot(t, engine, current.ID, current.Version, "PX-01", domain.BallotChoiceYes)

	replacement := proposal("P-AMEND", domain.ProposalTypeRule, domain.RiskLevelMedium)
	replacement.ProposerPraxeonID = current.ProposerPraxeonID
	replacement.Hash = "hash-P-AMEND-v2"
	replacement.BodyHash = "body-P-AMEND-v2"
	replacement.Summary = "Materially revised summary"
	replacement.Schedule = domain.Schedule{
		NoticeOpensAt:  baseTime.Add(7 * time.Minute),
		BallotOpensAt:  baseTime.Add(8 * time.Minute),
		BallotClosesAt: baseTime.Add(10 * time.Minute),
	}
	amended, err := engine.AmendProposal("PX-01", current.Version, replacement, baseTime.Add(4*time.Minute))
	if err != nil {
		t.Fatalf("amend proposal: %v", err)
	}
	if amended.Version != 2 || amended.Status != domain.ProposalStatusFiled || amended.SeconderPraxeonID != "" {
		t.Fatalf("amendment must require a new second: %+v", amended)
	}
	oldRequest, err := engine.VoteRequest(current.ID, current.Version, "PX-01")
	if err != nil {
		t.Fatalf("old request: %v", err)
	}
	if !oldRequest.Invalidated {
		t.Fatal("old request was not invalidated")
	}
	oldBallot, err := engine.Ballot(current.ID, current.Version, "PX-01")
	if err != nil {
		t.Fatalf("old ballot: %v", err)
	}
	if !oldBallot.Invalidated {
		t.Fatal("old ballot was not invalidated")
	}

	reNoticed, err := engine.SecondProposal("PX-03", amended.ID, amended.Version, baseTime.Add(6*time.Minute))
	if err != nil {
		t.Fatalf("second amended proposal: %v", err)
	}
	newRequest, err := engine.VoteRequest(reNoticed.ID, reNoticed.Version, "PX-01")
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if newRequest.ProposalHash != replacement.Hash || newRequest.Invalidated {
		t.Fatalf("new request did not use amended hash: %+v", newRequest)
	}
}

func TestProposalStatusAndAmendmentBoundaries(t *testing.T) {
	engine := domain.NewEngine(directory(3))
	current := fileAndSecond(t, engine, proposal("P-BOUNDARY", domain.ProposalTypeRule, domain.RiskLevelMedium))

	closed, err := engine.Proposal(current.ID, baseTime.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("closed proposal: %v", err)
	}
	if closed.Status != domain.ProposalStatusAwaitingDecision {
		t.Fatalf("closed proposal was still presented as open: %+v", closed)
	}

	replacement := proposal(current.ID, domain.ProposalTypeCivilizationGoal, domain.RiskLevelMedium)
	replacement.CivilizationGoalSpec = &domain.CivilizationGoalSpec{
		ID: "GOAL-TYPE-SWAP", Title: "Invalid type swap", SuccessChecks: goalChecks(),
	}
	if _, err := engine.AmendProposal("PX-01", current.Version, replacement, baseTime.Add(4*time.Minute)); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("proposal type changed across amendment: %v", err)
	}

	replacement = proposal(current.ID, domain.ProposalTypeRule, domain.RiskLevelMedium)
	replacement.Hash = "hash-P-BOUNDARY-v2"
	replacement.BodyHash = "body-P-BOUNDARY-v2"
	if _, err := engine.AmendProposal("PX-01", current.Version, replacement, baseTime.Add(5*time.Minute)); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("closed proposal accepted amendment: %v", err)
	}
}

func TestDirectCivilizationGoalCreationIsRejected(t *testing.T) {
	engine := domain.NewEngine(directory(1))
	_, err := engine.CreateGoal("PX-01", domain.Goal{
		ID:             "GOAL-INVALID",
		Type:           domain.GoalTypeCivilization,
		Title:          "Cannot be self-authorized",
		OwnerPraxeonID: "PX-01",
		SuccessChecks:  goalChecks(),
	}, baseTime)
	if !errors.Is(err, domain.ErrCivilizationGoalAuthority) {
		t.Fatalf("expected governance authority error, got %v", err)
	}
}
