module github.com/eyedeekay/i2p-tools-1

go 1.13

require (
	github.com/MDrollette/i2p-tools v0.0.0-20171015191648-e7d4585361c2
	github.com/codegangsta/cli v1.22.1
	github.com/cretz/bine v0.1.0
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/gorilla/handlers v1.4.2
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/justinas/alice v0.0.0-20171023064455-03f45bd4b7da
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/throttled/throttled v2.2.4+incompatible
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf // indirect
	golang.org/x/net v0.0.0-20191101175033-0deb6923b6d9 // indirect
	gopkg.in/throttled/throttled.v2 v2.2.4 // indirect
)

replace github.com/MDrollette/i2p-tools v0.0.0 => ./

replace github.com/codegangsta/cli v1.22.1 => github.com/urfave/cli v1.22.1
