module i2pgit.org/idk/reseed-tools

go 1.13

require (
	github.com/cretz/bine v0.1.0
	github.com/eyedeekay/checki2cp v0.0.21
	github.com/eyedeekay/go-i2pd v0.0.0-20220213070306-9807541b2dfc
	github.com/eyedeekay/i2pkeys v0.0.0-20220310055120-b97558c06ac8
	github.com/eyedeekay/sam3 v0.33.2
	github.com/go-acme/lego/v4 v4.3.1
	github.com/gorilla/handlers v1.5.1
	github.com/justinas/alice v1.2.0
	github.com/libp2p/go-libp2p v0.13.0
	github.com/libp2p/go-libp2p-core v0.8.0
	github.com/libp2p/go-libp2p-gostream v0.3.1
	github.com/libp2p/go-libp2p-http v0.2.0
	github.com/throttled/throttled/v2 v2.7.1
	github.com/urfave/cli v1.22.5
	github.com/urfave/cli/v3 v3.0.0-alpha
	gitlab.com/golang-commonmark/markdown v0.0.0-20191127184510-91b5b3c99c19
	golang.org/x/text v0.3.7
)

replace github.com/libp2p/go-libp2p => github.com/libp2p/go-libp2p v0.13.0

replace github.com/libp2p/go-libp2p-core => github.com/libp2p/go-libp2p-core v0.8.0

replace github.com/libp2p/go-libp2p-gostream => github.com/libp2p/go-libp2p-gostream v0.3.1

replace github.com/libp2p/go-libp2p-http => github.com/libp2p/go-libp2p-http v0.2.0

//replace github.com/eyedeekay/go-i2pd v0.0.0-20220213070306-9807541b2dfc => ./go-i2pd
