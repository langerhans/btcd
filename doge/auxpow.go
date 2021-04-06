package doge

const blockMinVersionAuxpow = 0x00620002
const blockVersionFlagAuxpow = 0x00000100

const versionAuxPow = 1 << 8

func IsAuxPoWBlockVersion(version int32) bool {
	return version >= blockMinVersionAuxpow && (version & blockVersionFlagAuxpow) > 0
}

func GetBaseVersion(version int32) int32 {
	return version % versionAuxPow
}