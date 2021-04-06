package doge

const digishieldBlockHeight = 145_000
const baseSubsidy = 50_000_000_000_000 //500k in satoshi
const subsidyDecreaseBlockCount = 100_000

func CalcBlockSubsidy(height int32) int64 {
	if height < digishieldBlockHeight {
		// Up until the Digishield hard fork, subsidy was based on the
		// previous block hash. Rather than actually recalculating that, we
		// simply use the maximum possible here, and let checkpoints enforce
		// that new blocks with different values can't be mined
		return (baseSubsidy << (height / subsidyDecreaseBlockCount)) * 2
	} else if height < 600000 {
		return baseSubsidy << (height / subsidyDecreaseBlockCount)
	} else {
		return baseSubsidy
	}
}