module github.com/eyedeekay/i2p-tools-1

go 1.13

replace github.com/MDrollette/i2p-tools v0.0.0-20171015191648-e7d4585361c2 => ./

require (
	github.com/MDrollette/i2p-tools v0.0.0-20171015191648-e7d4585361c2
	github.com/codegangsta/cli v1.22.0
	github.com/cretz/bine v0.1.0
	github.com/gorilla/handlers v1.4.2
	github.com/justinas/alice v0.0.0-20171023064455-03f45bd4b7da
	github.com/throttled/throttled v2.2.4+incompatible
	github.com/urfave/cli v1.22.1
)
