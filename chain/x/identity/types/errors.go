package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/identity module sentinel errors
var (
	ErrInvalidSigner     = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrCardAlreadyIssued = errors.Register(ModuleName, 1101, "card already issued for this address")
	ErrNoCard            = errors.Register(ModuleName, 1102, "address has no identity card")
	ErrAgentAlreadyBound = errors.Register(ModuleName, 1103, "agent address is already bound to a card")
	ErrAgentLimitReached = errors.Register(ModuleName, 1104, "card has reached the active agent limit")
	ErrInvalidAddress    = errors.Register(ModuleName, 1105, "creator or agent is not a valid chain address")
)
