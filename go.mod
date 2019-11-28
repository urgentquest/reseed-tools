module github.com/eyedeekay/i2p-tools-1

go 1.13

require (
	crawshaw.io/littleboss v0.0.0-20190317185602-8957d0aedcce
	github.com/MDrollette/i2p-tools v0.0.0
	github.com/codegangsta/cli v1.22.1
	github.com/cretz/bine v0.1.0
	github.com/gorilla/handlers v1.4.2
	github.com/justinas/alice v0.0.0-20171023064455-03f45bd4b7da
	github.com/libp2p/go-libp2p v0.4.2
	github.com/shurcooL/go v0.0.0-20190704215121-7189cc372560 // indirect
	github.com/shurcooL/go-goon v0.0.0-20170922171312-37c2f522c041 // indirect
	github.com/throttled/throttled v2.2.4+incompatible
)

replace github.com/MDrollette/i2p-tools v0.0.0 => ./

replace github.com/codegangsta/cli v1.22.1 => github.com/urfave/cli v1.22.1
