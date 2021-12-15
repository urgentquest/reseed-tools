
VERSION=0.2.11
APP=reseed-tools
USER_GH=eyedeekay
CGO_ENABLED=0
export CGO_ENABLED=0
PLUGIN_PORT=7671
export PLUGIN_PORT=7671

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

index:
	@echo "<!DOCTYPE html>" > index.html
	@echo "<html>" >> index.html
	@echo "<head>" >> index.html
	@echo "  <title>Reseed Tools</title>" >> index.html
	@echo "  <link rel=\"stylesheet\" type=\"text/css\" href =\"/style.css\" />" >> index.html
	@echo "</head>" >> index.html
	@echo "<body>" >> index.html
	pandoc README.md >> index.html
	@echo "</body>" >> index.html
	@echo "</html>" >> index.html

build:
	go build $(ARG) -o reseed-tools-$(GOOS)-$(GOARCH)

1.15-build: gofmt
	/usr/lib/go-$(MIN_GO_VERSION)/bin/go build $(ARG) -o reseed-tools-$(GOOS)-$(GOARCH)

clean:
	rm reseed-tools-* tmp -rfv *.deb plugin reseed-tools

tar:
	tar --exclude="./.git" --exclude="./tmp" --exclude=".vscode" --exclude="./*.pem" --exclude="./*.crl" --exclude="./*.crt" -cvf ../reseed-tools.tar.xz .

install:
	install -m755 reseed-tools-$(GOOS)-$(GOARCH) /usr/bin/reseed-tools
	install -m644 etc/default/reseed /etc/default/reseed
	install -m755 etc/init.d/reseed /etc/init.d/reseed
	mkdir -p /etc/systemd/system/reseed.d/
	mkdir -p /var/lib/i2p
	install -g i2psvc -o i2psvc -d /var/lib/i2p/i2p-config/reseed/
	cp -r content /var/lib/i2p/i2p-config/reseed/content
	chown -R i2psvc:i2psvc /var/lib/i2p/i2p-config/reseed/
	install -m644 etc/systemd/system/reseed.d/reseed.conf /etc/systemd/system/reseed.d/reseed.conf
	install -m644 etc/systemd/system/reseed.d/reseed.service /etc/systemd/system/reseed.d/reseed.service

uninstall:
	rm /usr/bin/reseed-tools
	rm /etc/default/reseed
	rm /etc/init.d/reseed
	rm /etc/systemd/system/reseed.d/reseed.conf
	rm /etc/systemd/system/reseed.d/reseed.service
	rm -rf /var/lib/i2p/i2p-config/reseed/

checkinstall: build
	fakeroot checkinstall \
		--default \
		--install=no \
		--fstrans=yes \
		--pkgname=reseed-tools \
		--pkgversion=$(VERSION) \
		--pkggroup=net \
		--pkgrelease=1 \
		--pkgsource="https://i2pgit.org/idk/reseed-tools" \
		--maintainer="$(SIGNER)" \
		--requires="i2p,i2p-router" \
		--suggests="i2p,i2p-router,syndie,tor,tsocks" \
		--nodoc \
		--deldoc=yes \
		--deldesc=yes \
		--backup=no

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
		../reseed-tools-$(GOOS)-$(GOARCH) reseed --signer=you@mail.i2p --netdb=/home/idk/.i2p/netDb --tlsHost=your-domain.tld --onion --p2p --i2p

stop:
	mkdir -p tmp && \
		cd tmp && \
		../reseed-tools-$(GOOS)-$(GOARCH) reseed --signer=you@mail.i2p --netdb=/home/idk/.i2p/netDb --tlsHost=your-domain.tld --onion --p2p --i2p

docker:
	docker build -t eyedeekay/reseed .

docker-push: docker
	docker push --disable-content-trust=false eyedeekay/reseed:$(VERSION)

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

release: version upload checkinstall upload-single-deb plugins upload-su3s upload-bin 

version:
	cat README.md | gothub release -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -; true

delete-version:
	gothub delete -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION)

edit:
	cat README.md | gothub edit -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -d -

upload: tar
	gothub upload -R -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -f ../reseed-tools.tar.xz -n "reseed-tools.tar.xz"

binary:
	GOOS=darwin GOARCH=amd64 make build
	GOOS=darwin GOARCH=arm64 make build
	GOOS=linux GOARCH=386 make build
	GOOS=linux GOARCH=amd64 make build
	GOOS=linux GOARCH=arm make build
	GOOS=linux GOARCH=arm64 make build
	GOOS=openbsd GOARCH=amd64 make build
	GOOS=freebsd GOARCH=386 make build
	GOOS=freebsd GOARCH=amd64 make build
	GOOS=windows GOARCH=amd64 make build
	GOOS=windows GOARCH=386 make build

plugins: binary
	GOOS=darwin GOARCH=amd64 make su3s
	GOOS=darwin GOARCH=arm64 make su3s
	GOOS=linux GOARCH=386 make su3s
	GOOS=linux GOARCH=amd64 make su3s
	GOOS=linux GOARCH=arm make su3s
	GOOS=linux GOARCH=arm64 make su3s
	GOOS=openbsd GOARCH=amd64 make su3s
	GOOS=freebsd GOARCH=386 make su3s
	GOOS=freebsd GOARCH=amd64 make su3s
	GOOS=windows GOARCH=amd64 make su3s
	GOOS=windows GOARCH=386 make su3s

upload-bin:
	GOOS=darwin GOARCH=amd64 make upload-single-bin
	GOOS=darwin GOARCH=arm64 make upload-single-bin
	GOOS=linux GOARCH=386 make upload-single-bin
	GOOS=linux GOARCH=amd64 make upload-single-bin
	GOOS=linux GOARCH=arm make upload-single-bin
	GOOS=linux GOARCH=arm64 make upload-single-bin
	GOOS=openbsd GOARCH=amd64 make upload-single-bin
	GOOS=freebsd GOARCH=386 make upload-single-bin
	GOOS=freebsd GOARCH=amd64 make upload-single-bin
	GOOS=windows GOARCH=amd64 make upload-single-bin
	GOOS=windows GOARCH=386 make upload-single-bin

rm-su3s:
	rm *.su3 -f

download-su3s:
	GOOS=darwin GOARCH=amd64 make download-single-su3
	GOOS=darwin GOARCH=arm64 make download-single-su3
	GOOS=linux GOARCH=386 make download-single-su3
	GOOS=linux GOARCH=amd64 make download-single-su3
	GOOS=linux GOARCH=arm make download-single-su3
	GOOS=linux GOARCH=arm64 make download-single-su3
	GOOS=openbsd GOARCH=amd64 make download-single-su3
	GOOS=freebsd GOARCH=386 make download-single-su3
	GOOS=freebsd GOARCH=amd64 make download-single-su3
	GOOS=windows GOARCH=amd64 make download-single-su3
	GOOS=windows GOARCH=386 make download-single-su3

upload-su3s:
	GOOS=darwin GOARCH=amd64 make upload-single-su3
	GOOS=darwin GOARCH=arm64 make upload-single-su3
	GOOS=linux GOARCH=386 make upload-single-su3
	GOOS=linux GOARCH=amd64 make upload-single-su3
	GOOS=linux GOARCH=arm make upload-single-su3
	GOOS=linux GOARCH=arm64 make upload-single-su3
	GOOS=openbsd GOARCH=amd64 make upload-single-su3
	GOOS=freebsd GOARCH=386 make upload-single-su3
	GOOS=freebsd GOARCH=amd64 make upload-single-su3
	GOOS=windows GOARCH=amd64 make upload-single-su3
	GOOS=windows GOARCH=386 make upload-single-su3

download-single-su3:
	wget -N -c "https://github.com/eyedeekay/reseed-tools/releases/download/v$(VERSION)/reseed-tools-$(GOOS)-$(GOARCH).su3"

upload-single-deb:
	gothub upload -R -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -f reseed-tools_$(VERSION)-1_amd64.deb -l "`sha256sum reseed-tools_$(VERSION)-1_amd64.deb`" -n "reseed-tools_$(VERSION)-1_amd64.deb"

upload-single-bin:
	gothub upload -R -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -f reseed-tools-"$(GOOS)"-"$(GOARCH)" -l "`sha256sum reseed-tools-$(GOOS)-$(GOARCH)`" -n "reseed-tools-$(GOOS)"-"$(GOARCH)"

upload-single-su3:
	gothub upload -R -s $(GITHUB_TOKEN) -u $(USER_GH) -r $(APP) -t v$(VERSION) -f reseed-tools-"$(GOOS)"-"$(GOARCH).su3" -l "`sha256sum reseed-tools-$(GOOS)-$(GOARCH).su3`" -n "reseed-tools-$(GOOS)"-"$(GOARCH).su3"

tmp/content:
	mkdir -p tmp
	cp -rv content tmp/content

tmp/lib:
	mkdir -p tmp/lib
	cp "$(HOME)/Workspace/GIT_WORK/i2p.i2p/build/shellservice.jar" tmp/lib/shellservice.jar

su3s: tmp/content tmp/lib
	i2p.plugin.native -name=reseed-tools-$(GOOS)-$(GOARCH) \
		-signer=hankhill19580@gmail.com \
		-version "$(VERSION)" \
		-author=hankhill19580@gmail.com \
		-autostart=true \
		-clientname=reseed-tools-$(GOOS)-$(GOARCH) \
		-command="reseed-tools-$(GOOS)-$(GOARCH) reseed --yes --signer=\$$PLUGIN/signer --port=$(PLUGIN_PORT)" \
		-consolename="Reseed Tools" \
		-consoleurl="https://127.0.0.1:$(PLUGIN_PORT)" \
		-updateurl="http://idk.i2p/reseed-tools/reseed-tools-$(GOOS)-$(GOARCH).su3" \
		-website="http://idk.i2p/reseed-tools/" \
		-icondata="content/images/reseed-icon.png" \
		-delaystart="1" \
		-desc="`cat description-pak`" \
		-exename=reseed-tools-$(GOOS)-$(GOARCH) \
		-targetos="$(GOOS)" \
		-res=tmp/ \
		-license=MIT
	unzip -o reseed-tools-$(GOOS)-$(GOARCH).zip -d reseed-tools-$(GOOS)-$(GOARCH)-zip

#export sumbblinux=`sha256sum "../reseed-tools-linux.su3"`
#export sumbbwindows=`sha256sum "../reseed-tools-windows.su3"`
