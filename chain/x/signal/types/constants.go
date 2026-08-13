package types

const (
	// SignalTypeUse is a high-weight signal from real usage/trade.
	SignalTypeUse = uint32(1)
	// SignalTypeLike is a low-weight signal (a mere like).
	SignalTypeLike = uint32(2)

	// UseWeight is the weight of a Use signal.
	UseWeight = uint64(10)
	// LikeWeight is the weight of a Like signal.
	LikeWeight = uint64(1)

	// BlocksPerSol is the number of chain blocks in one Martian sol
	// (1 sol = 24h39m35.244s = 88775.244s / 6s ≈ 14796 blocks).
	BlocksPerSol = int64(14796)

	// LikeQuotaPerSol is the max number of Like signals per address per sol.
	// Use signals are not limited (they burn real Prax).
	LikeQuotaPerSol = uint64(10)

	// SignalHalvingSol is the halving period for signal decay.
	// One Martian year = 670 sol: signal weight halves every 670 sol.
	SignalHalvingSol = int64(670)

	// LikeMinSenderWeight is the minimum sender signal weight for a Like
	// signal to be valid (defense ③: weight follows contribution).
	LikeMinSenderWeight = uint64(10)
)
