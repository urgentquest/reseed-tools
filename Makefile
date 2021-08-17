
VERSION=0.0.8
APP=reseed-tools
USER_GH=eyedeekay

GOOS?=$(shell uname -s | tr A-Z a-z)
GOARCH?="amd64"

ARG=-v -tags netgo -ldflags '-w -extldflags "-static"'

#MIN_GO_VERSION=`ls /usr/lib/go-1.14 2>/dev/null >/dev/null && echo 1.14`
MIN_GO_VERSION?=1.15

I2P_UID=$(shell id -u i2psvc)
I2P_GID=$(shell id -g i2psvc)

WHOAMI=$(shell whoami)

echo:
	@echo "type make version to do release $(APP) $(VERSION) $(GOOS) $(GOARCH) $(MIN_GO_VERSION) $(I2P_UID) $(I2P_GID)"

version:
	cat README.md | gothub release -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -

edit:
	cat README.md | gothub edit -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -

upload: binary tar
	gothub upload -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -f ../reseed-tools.tar.xz -n "reseed-tools.tar.xz"

build: gofmt
	/usr/lib/go-$(MIN_GO_VERSION)/bin/go build $(ARG) -o reseed-tools-$(GOOS)-$(GOARCH)

clean:
	rm reseed-tools-* *.key *.i2pKeys *.crt *.crl *.pem tmp -rfv

binary:
	GOOS=darwin GOARCH=amd64 make build
	GOOS=linux GOARCH=386 make build
	GOOS=linux GOARCH=amd64 make build
	GOOS=linux GOARCH=arm make build
	GOOS=linux GOARCH=arm64 make build
	GOOS=openbsd GOARCH=amd64 make build
	GOOS=freebsd GOARCH=386 make build
	GOOS=freebsd GOARCH=amd64 make build
	GOOS=windows GOARCH=amd64 make build

tar:
	tar --exclude="./.git" --exclude="./tmp"  -cvf ../reseed-tools.tar.xz .

install:
	install -m755 reseed-tools-$(GOOS)-$(GOARCH) /usr/local/bin/reseed-tools
	install -m755 etc/init.d/reseed /etc/init.d/reseed

### You shouldn't need to use these now that the go mod require rule is fixed,
## but I'm leaving them in here because it made it easier to test that both
## versions behaved the same way. -idk

build-fork:
	/usr/lib/go-$(MIN_GO_VERSION)/bin/go build -o reseed-tools-idk

build-unfork:
	/usr/lib/go-$(MIN_GO_VERSION)/bin/go build -o reseed-tools-md

fork:
	sed -i 's|idk/reseed-tools|idk/reseed-tools|g' main.go cmd/*.go reseed/*.go su3/*.go
	make gofmt build-fork

unfork:
	sed -i 's|idk/reseed-tools|idk/reseed-tools|g' main.go cmd/*.go reseed/*.go su3/*.go
	sed -i 's|RTradeLtd/reseed-tools|idk/reseed-tools|g' main.go cmd/*.go reseed/*.go su3/*.go
	make gofmt build-unfork

gofmt:
	gofmt -w main.go cmd/*.go reseed/*.go su3/*.go

try:
	mkdir -p tmp && \
		cd tmp && \
		../reseed-tools-$(GOOS)-$(GOARCH) reseed --signer=you@mail.i2p --netdb=/home/idk/.i2p/netDb --tlsHost=your-domain.tld --onion --p2p --i2p --littleboss=start

stop:
	mkdir -p tmp && \
		cd tmp && \
		../reseed-tools-$(GOOS)-$(GOARCH) reseed --signer=you@mail.i2p --netdb=/home/idk/.i2p/netDb --tlsHost=your-domain.tld --onion --p2p --i2p --littleboss=stop

docker:
	docker build -t eyedeekay/reseed .

docker-push: docker
	docker push --disable-content-trust false eyedeekay/reseed:$(VERSION)

users:
	docker run --rm eyedeekay/reseed cat /etc/passwd

docker-ls:
		docker run --rm \
		--user $(I2P_UID) \
		--group-add $(I2P_GID) \
		--name reseed \
		--publish 8443:8443 \
		--volume /var/lib/i2p/i2p-config/netDb:/var/lib/i2p/i2p-config/netDb \
		eyedeekay/reseed ls /var/lib/i2p/i2p-config -lah

docker-server:
	docker run -itd \
		--name reseed \
		--user $(I2P_UID) \
		--group-add $(I2P_GID) \
		--publish 8443:8443 \
		--restart=always \
		--volume /var/lib/i2p/i2p-config/netDb:/var/lib/i2p/i2p-config/netDb:z \
		--volume reseed-keys:/var/lib/i2p/i2p-config/reseed \
		eyedeekay/reseed \
			--signer=hankhill19580@gmail.com
	docker logs -f reseed

docker-run:
	docker run -itd \
		--name reseed \
		--user $(I2P_UID) \
		--group-add $(I2P_GID) \
		--publish 8443:8443 \
		--volume /var/lib/i2p/i2p-config/netDb:/var/lib/i2p/i2p-config/netDb:z \
		--volume reseed-keys:/var/lib/i2p/i2p-config/reseed \
		eyedeekay/reseed \
			--signer=hankhill19580@gmail.com

docker-homerun:
	docker run -itd \
		--name reseed \
		--user 1000 \
		--group-add 1000 \
		--publish 8443:8443 \
		--volume $(HOME)/i2p/netDb:/var/lib/i2p/i2p-config/netDb:z \
		--volume reseed-keys:/var/lib/i2p/i2p-config/reseed:z \
		eyedeekay/reseed \
			--signer=hankhill19580@gmail.com

export JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64/jre/
export CGO_CFLAGS=-I/usr/lib/jvm/java-8-openjdk-amd64/include/ -I/usr/lib/jvm/java-8-openjdk-amd64/include/linux/

gojava:
	go get -u -v github.com/sridharv/gojava
	cp -v ~/go/bin/gojava ./gojava

jar: gojava
	echo $(JAVA_HOME)
	./gojava -v -o reseed.jar -s . build ./reseed

plugins: binary
	GOOS=darwin GOARCH=amd64 make su3s
	GOOS=linux GOARCH=386 make su3s
	GOOS=linux GOARCH=amd64 make su3s
	GOOS=linux GOARCH=arm make su3s
	GOOS=linux GOARCH=arm64 make su3s
	GOOS=openbsd GOARCH=amd64 make su3s
	GOOS=freebsd GOARCH=386 make su3s
	GOOS=freebsd GOARCH=amd64 make su3s
	GOOS=windows GOARCH=amd64 make su3s

su3s:
	i2p.plugin.native -name=reseed-tools-$(GOOS)-$(GOARCH) \
		-signer=hankhill19580@gmail.com \
		-version "$(VERSION)" \
		-author=hankhill19580@gmail.com \
		-autostart=true \
		-clientname=reseed-tools-$(GOOS)-$(GOARCH) \
		-command="\$$PLUGIN/lib/reseed-tools-$(GOOS)-$(GOARCH)s -dir=\$$PLUGIN/lib reseed --signer=you@mail.i2p --netdb=\$$CONFIG/netDb --onion --i2p" \
		-consolename="Reseed Tools" \
		-delaystart="200" \
		-desc="Reseed Tools Plugin" \
		-exename=reseed-tools-$(GOOS)-$(GOARCH) \
		-targetos="$(GOOS)" \
		-license=MIT
	unzip -o reseed-tools-$(GOOS)-$(GOARCH).zip -d reseed-tools-$(GOOS)-$(GOARCH)-zip

#export sumbblinux=`sha256sum "../reseed-tools-linux.su3"`
#export sumbbwindows=`sha256sum "../reseed-tools-windows.su3"`