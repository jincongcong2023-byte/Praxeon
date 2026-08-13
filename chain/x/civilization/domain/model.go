package domain

import "time"

type GoalType uint8

const (
	GoalTypeUnspecified GoalType = iota
	GoalTypePraxeon
	GoalTypeAgent
	GoalTypeCivilization
)

type GoalStatus uint8

const (
	GoalStatusUnspecified GoalStatus = iota
	GoalStatusDraft
	GoalStatusActive
	GoalStatusBlocked
	GoalStatusPaused
	GoalStatusAchieved
	GoalStatusAbandoned
	GoalStatusExpired
)

type ProposalType uint8

const (
	ProposalTypeUnspecified ProposalType = iota
	ProposalTypeRule
	ProposalTypeCharter
	ProposalTypeCivilizationGoal
	ProposalTypeRevoke
	ProposalTypeProcedure
)

type ProposalStatus uint8

const (
	ProposalStatusUnspecified ProposalStatus = iota
	ProposalStatusFiled
	ProposalStatusNoticed
	ProposalStatusDeliberating
	ProposalStatusBallotOpen
	ProposalStatusRatified
	ProposalStatusRejected
	ProposalStatusExpiredNoQuorum
	ProposalStatusInvalid
	ProposalStatusAwaitingDecision
)

func (s ProposalStatus) Terminal() bool {
	return s == ProposalStatusRatified || s == ProposalStatusRejected || s == ProposalStatusExpiredNoQuorum || s == ProposalStatusInvalid
}

type RiskLevel uint8

const (
	RiskLevelUnspecified RiskLevel = iota
	RiskLevelLow
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

type ApprovalRule uint8

const (
	ApprovalRuleUnspecified ApprovalRule = iota
	ApprovalRuleSimpleMajority
	ApprovalRuleTwoThirds
)

type VoteMode uint8

const (
	VoteModeUnspecified VoteMode = iota
	VoteModeNotifyOnly
	VoteModeConfirmEach
	VoteModePolicyVote
)

type BallotChoice uint8

const (
	BallotChoiceUnspecified BallotChoice = iota
	BallotChoiceYes
	BallotChoiceNo
	BallotChoiceAbstain
)

func (c BallotChoice) Valid() bool {
	return c == BallotChoiceYes || c == BallotChoiceNo || c == BallotChoiceAbstain
}

type SubmitterType uint8

const (
	SubmitterTypeUnspecified SubmitterType = iota
	SubmitterTypePraxeon
	SubmitterTypeAgent
)

type EventType uint8

const (
	EventTypeUnspecified EventType = iota
	EventTypeGoalCreated
	EventTypeProposalFiled
	EventTypeProposalNoticed
	EventTypeProposalAmended
	EventTypeVoteAuthorizationChanged
	EventTypeBallotConfirmed
	EventTypeBallotSubmitted
	EventTypeDecisionFinalized
	EventTypeCivilizationGoalActivated
	EventTypeVoteRequestAcknowledged
)

type EventVisibility uint8

const (
	EventVisibilityUnspecified EventVisibility = iota
	EventVisibilityPublic
	EventVisibilityPraxeon
	EventVisibilityAgent
)

type Event struct {
	Cursor      uint64
	ID          string
	Type        EventType
	SubjectID   string
	PraxeonID   string
	AgentID     string
	Visibility  EventVisibility
	PayloadHash string
	OccurredAt  time.Time
}

type GoalCheck struct {
	ID              string
	Description     string
	WeightBPS       uint32
	Satisfied       bool
	SourceResultIDs []string
}

type Goal struct {
	ID                   string
	Type                 GoalType
	Title                string
	DescriptionHash      string
	OwnerPraxeonID       string
	GovernanceProposalID string
	AssignedAgentIDs     []string
	Status               GoalStatus
	SuccessChecks        []GoalCheck
	CreatedAt            time.Time
	TargetAt             time.Time
}

type CivilizationGoalSpec struct {
	ID               string
	Title            string
	DescriptionHash  string
	SuccessChecks    []GoalCheck
	AssignedAgentIDs []string
	TargetAt         time.Time
}

type Schedule struct {
	NoticeOpensAt  time.Time
	BallotOpensAt  time.Time
	BallotClosesAt time.Time
}

type Proposal struct {
	ID                   string
	Version              uint64
	Hash                 string
	Type                 ProposalType
	Title                string
	Summary              string
	BodyHash             string
	ProposerPraxeonID    string
	SeconderPraxeonID    string
	Domain               string
	Risk                 RiskLevel
	Approval             ApprovalRule
	AffectsGovernance    bool
	ReducesRights        bool
	ProcedureKind        string
	TargetApproval       ApprovalRule
	Schedule             Schedule
	EligiblePraxeonIDs   []string
	EligibilityHash      string
	EligibilityFrozenAt  time.Time
	Status               ProposalStatus
	CivilizationGoalSpec *CivilizationGoalSpec
}

func (p Proposal) StatusAt(now time.Time) ProposalStatus {
	if p.Status.Terminal() || p.Status == ProposalStatusFiled {
		return p.Status
	}
	if now.Before(p.Schedule.NoticeOpensAt) {
		return ProposalStatusNoticed
	}
	if now.Before(p.Schedule.BallotOpensAt) {
		return ProposalStatusDeliberating
	}
	if now.Before(p.Schedule.BallotClosesAt) {
		return ProposalStatusBallotOpen
	}
	return ProposalStatusAwaitingDecision
}

type VoteRequestDeliveryStatus uint8

const (
	VoteRequestDeliveryStatusUnspecified VoteRequestDeliveryStatus = iota
	VoteRequestDeliveryStatusDelivered
	VoteRequestDeliveryStatusFailed
)

type PraxeonRoute struct {
	PraxeonID         string
	GovernanceAgentID string
	PrimaryAgentID    string
	Active            bool
}

type VoteRequest struct {
	ID                string
	ProposalID        string
	ProposalVersion   uint64
	ProposalHash      string
	EligiblePraxeonID string
	RecipientAgentID  string
	DeliveryStatus    VoteRequestDeliveryStatus
	DeliveryFailure   string
	DeliveredAt       time.Time
	AcknowledgedAt    time.Time
	BallotDeadline    time.Time
	Invalidated       bool
}

type VoteAuthorization struct {
	ID                   string
	Hash                 string
	PraxeonID            string
	AgentID              string
	Mode                 VoteMode
	AllowedProposalTypes []ProposalType
	AllowedDomains       []string
	AllowedProposalIDs   []string
	MaxRisk              RiskLevel
	ValidFrom            time.Time
	ExpiresAt            time.Time
	DecisionPolicyHash   string
	CanAbstain           bool
	RequiresReason       bool
	Revoked              bool
}

type BallotConfirmation struct {
	ID              string
	PraxeonID       string
	AgentID         string
	ProposalID      string
	ProposalVersion uint64
	Choice          BallotChoice
	ExpiresAt       time.Time
	Used            bool
}

type Ballot struct {
	ID                string
	ProposalID        string
	ProposalVersion   uint64
	ProposalHash      string
	VoterPraxeonID    string
	Choice            BallotChoice
	SubmittedBy       SubmitterType
	SignerID          string
	AuthorizationID   string
	AuthorizationHash string
	ConfirmationID    string
	ReasonHash        string
	Sequence          uint64
	PraxeonLock       bool
	Invalidated       bool
	SubmittedAt       time.Time
}

type Decision struct {
	ID              string
	ProposalID      string
	ProposalVersion uint64
	Outcome         ProposalStatus
	Eligible        uint32
	Delivered       uint32
	Acknowledged    uint32
	Yes             uint32
	No              uint32
	Abstain         uint32
	NotVoted        uint32
	Approval        ApprovalRule
	FinalizedAt     time.Time
}

type AgentSnapshot struct {
	ID                             string
	AgentID                        string
	PraxeonID                      string
	ActiveGoals                    []Goal
	VoteRequests                   []VoteRequest
	RequiredConfirmationRequestIDs []string
	EventCursor                    uint64
	GeneratedAt                    time.Time
}

type IdentityDirectory interface {
	Routes() []PraxeonRoute
	PraxeonForAgent(agentID string) (string, bool)
}
