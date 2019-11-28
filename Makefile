
VERSION=0.0.2
APP=i2p-tools-1
USER_GH=eyedeekay

GOOS?=$(shell uname -s | tr A-Z a-z)
GOARCH?="amd64"

#ARG=-v -tags netgo -ldflags '-w -extldflags "-static"'

echo:
	@echo "type make version to do release $(APP) $(VERSION) $(GOOS) $(GOARCH) "

version:
	cat README.md | gothub release -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -

edit:
	cat README.md | gothub edit -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -

upload: binary tar
	gothub upload -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -f ../i2p-tools.tar.xz -n "i2p-tools.tar.xz"

build: gofmt
	go build $(ARG) -o i2p-tools-$(GOOS)-$(GOARCH)

clean:
	rm i2p-tools-* *.key *.i2pKeys *.crt *.crl *.pem tmp -rf

binary:
	GOOS=darwin GOARCH=amd64 make build
	GOOS=linux GOARCH=386 make build
	GOOS=linux GOARCH=amd64 make build
	GOOS=linux GOARCH=arm make build
	GOOS=linux GOARCH=arm64 make build
	GOOS=openbsd GOARCH=amd64 make build
	GOOS=freebsd GOARCH=386 make build
	GOOS=freebsd GOARCH=amd64 make build

tar:
	tar --exclude="./.git" --exclude="./tmp"  -cvf ../i2p-tools.tar.xz .

install:
	install -m755 i2p-tools-$(GOOS)-$(GOARCH) /usr/local/bin/i2p-tools
	install -m755 etc/init.d/reseed /etc/init.d/reseed

### You shouldn't need to use these now that the go mod require rule is fixed,
## but I'm leaving them in here because it made it easier to test that both
## versions behaved the same way. -idk

build-fork:
	go build -o i2p-tools-idk

build-unfork:
	go build -o i2p-tools-md

fork:
	sed -i 's|MDrollette/i2p-tools|eyedeekay/i2p-tools-1|g' main.go cmd/*.go reseed/*.go su3/*.go
	make gofmt build-fork

unfork:
	sed -i 's|eyedeekay/i2p-tools-1|MDrollette/i2p-tools|g' main.go cmd/*.go reseed/*.go su3/*.go
	sed -i 's|RTradeLtd/i2p-tools-1|MDrollette/i2p-tools|g' main.go cmd/*.go reseed/*.go su3/*.go
	make gofmt build-unfork

gofmt:
	gofmt -w main.go cmd/*.go reseed/*.go su3/*.go

try:
	mkdir -p tmp && \
		cd tmp && \
		../i2p-tools-$(GOOS)-$(GOARCH) reseed --signer=you@mail.i2p --netdb=/home/idk/.i2p/netDb --tlsHost=your-domain.tld --onion --p2p --i2p --littleboss=start

stop:
	mkdir -p tmp && \
		cd tmp && \
		../i2p-tools-$(GOOS)-$(GOARCH) reseed --signer=you@mail.i2p --netdb=/home/idk/.i2p/netDb --tlsHost=your-domain.tld --onion --p2p --i2p --littleboss=stop