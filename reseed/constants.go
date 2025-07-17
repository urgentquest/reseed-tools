package reseed

// Application version
// Moved from: version.go
const Version = "0.3.3"

// HTTP User Agent constants
// Moved from: server.go
const (
	I2pUserAgent = "Wget/1.11.4"
)

// Random string generation constants
// Moved from: server.go
const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52 possibilities
	letterIdxBits = 6                                                      // 6 bits to represent 64 possibilities / indexes
	letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
)
