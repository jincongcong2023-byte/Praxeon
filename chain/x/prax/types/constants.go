package types

// Prax minting parameters — protocol-fixed (Austrian sound money), not governable.
const (
	// InitialMintPerSol is the uprax minted per sol (civilization cycle).
	// 75000 Prax = 75e9 uprax (1 Prax = 10^6 uprax).
	InitialMintPerSol = int64(75_000_000_000)

	// BlocksPerSol is the number of chain blocks in one Martian sol.
	// 1 sol = 24h39m35.244s = 88775.244s / 6s ≈ 14796 blocks (real 1:1).
	BlocksPerSol = int64(14796)

	// HalvingBlocks is the block interval between halvings.
	// One Martian year = 669.6 sol ≈ 670 sol = 670 * BlocksPerSol blocks.
	HalvingBlocks = int64(670 * 14796)

	// TotalSupplyCap is the maximum total Prax supply.
	// 100M Prax = 1e14 uprax.
	// The staking denom is separated (stake), so uprax supply starts at 0
	// and is minted only by this module.
	TotalSupplyCap = int64(100_000_000_000_000)

	// PraxDenom is the base denom of the Prax currency.
	PraxDenom = "uprax"
)
