package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/signal module sentinel errors
var (
	ErrInvalidSigner      = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidSignalType  = errors.Register(ModuleName, 1101, "invalid signal type (must be 1=Use or 2=Like)")
	ErrLikeQuotaExceeded  = errors.Register(ModuleName, 1102, "like quota exceeded for this sol")
	ErrSenderWeightTooLow = errors.Register(ModuleName, 1103, "sender signal weight too low to send a like")
	ErrSelfSignal         = errors.Register(ModuleName, 1104, "cannot signal an address owned by the same card")
	ErrDuplicateLike      = errors.Register(ModuleName, 1105, "sender already liked this target in the current sol")
	ErrUseRequiresReceipt = errors.Register(ModuleName, 1106, "use signal requires a verified adoption receipt")
	ErrInvalidAddress     = errors.Register(ModuleName, 1107, "sender or target is not a valid chain address")
	ErrSignalOverflow     = errors.Register(ModuleName, 1108, "signal weight exceeds the supported range")
)
