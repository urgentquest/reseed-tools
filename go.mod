module github.com/eyedeekay/i2p-tools-1

go 1.13

require (
	crawshaw.io/littleboss v0.0.0-20190317185602-8957d0aedcce // indirect
	github.com/MDrollette/i2p-tools v0.0.0
	github.com/codegangsta/cli v1.22.1
	github.com/cretz/bine v0.1.0
	github.com/eyedeekay/sam3 v0.32.2
	github.com/gomodule/redigo v1.8.0 // indirect
	github.com/gorilla/handlers v1.4.2
	github.com/justinas/alice v0.0.0-20171023064455-03f45bd4b7da
	github.com/libp2p/go-libp2p v0.6.0
	github.com/libp2p/go-libp2p-core v0.5.0
	github.com/libp2p/go-libp2p-gostream v0.2.1
	github.com/libp2p/go-libp2p-http v0.1.5
	github.com/shurcooL/go v0.0.0-20190704215121-7189cc372560 // indirect
	github.com/shurcooL/go-goon v0.0.0-20170922171312-37c2f522c041 // indirect
	github.com/throttled/throttled v2.2.4+incompatible
)

replace github.com/MDrollette/i2p-tools v0.0.0 => ./

replace github.com/codegangsta/cli v1.22.1 => github.com/urfave/cli v1.22.1
