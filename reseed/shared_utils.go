package reseed

// SharedUtilities provides common utility functions used across the reseed package.
// Moved from: various files

import (
	"strings"
)

// AllReseeds contains the list of all available reseed servers.
// Moved from: ping.go
var AllReseeds = []string{
	"https://banana.incognet.io/",
	"https://i2p.novg.net/",
	"https://i2pseed.creativecowpat.net:8443/",
	"https://reseed-fr.i2pd.xyz/",
	"https://reseed-pl.i2pd.xyz/",
	"https://reseed.diva.exchange/",
	"https://reseed.i2pgit.org/",
	"https://reseed.memcpy.io/",
	"https://reseed.onion.im/",
	"https://reseed2.i2p.net/",
	"https://www2.mk16.de/",
}

// SignerFilenameFromID creates a filename-safe version of a signer ID.
// Moved from: utils.go
func SignerFilenameFromID(signerID string) string {
	return strings.Replace(signerID, "@", "_at_", 1)
}
