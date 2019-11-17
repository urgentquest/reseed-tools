
VERSION=0.0.1
APP=i2p-tools-1
USER_GH=eyedeekay

echo:
	@echo "type make version to do release $(APP) $(VERSION)"

version:
	cat README.md | gothub release -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -

edit:
	cat README.md | gothub edit -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -

build:
	go build -v -tags netgo \
		-ldflags '-w -extldflags "-static"' -o i2p-tools

install:
	install -m755 i2p-tools /usr/local/bin/i2p-tools
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