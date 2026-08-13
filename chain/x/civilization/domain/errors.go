package domain

import "errors"

var (
	ErrNotFound                  = errors.New("not found")
	ErrAlreadyExists             = errors.New("already exists")
	ErrInvalidInput              = errors.New("invalid input")
	ErrNotEligible               = errors.New("praxeon is not eligible")
	ErrInvalidState              = errors.New("invalid lifecycle state")
	ErrVersionMismatch           = errors.New("proposal version mismatch")
	ErrProposalHashMismatch      = errors.New("proposal hash mismatch")
	ErrProposalNotOpen           = errors.New("proposal ballot is not open")
	ErrProposalClosed            = errors.New("proposal ballot is closed")
	ErrProposalNotClosed         = errors.New("proposal ballot is not closed")
	ErrSameSeconder              = errors.New("proposer cannot second own proposal")
	ErrAuthorizationRequired     = errors.New("vote authorization required")
	ErrAuthorizationInvalid      = errors.New("vote authorization is invalid")
	ErrAuthorizationExpired      = errors.New("vote authorization is expired or inactive")
	ErrConfirmationRequired      = errors.New("single-ballot confirmation required")
	ErrConfirmationInvalid       = errors.New("single-ballot confirmation is invalid")
	ErrNotifyOnly                = errors.New("notify-only authorization cannot submit ballot")
	ErrPraxeonLocked             = errors.New("praxeon ballot locks out agent updates")
	ErrSequenceMismatch          = errors.New("ballot sequence mismatch")
	ErrAgentOwnership            = errors.New("agent does not belong to voter praxeon")
	ErrFinalized                 = errors.New("proposal already finalized")
	ErrCivilizationGoalAuthority = errors.New("civilization goal requires ratified governance")
)
