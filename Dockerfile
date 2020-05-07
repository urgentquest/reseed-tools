FROM debian:stable-backports
ARG I2P_GID=1000
ARG I2P_UID=1000
COPY . /var/lib/i2p/go/src/github.com/eyedeekay/i2p-tools-1
WORKDIR /var/lib/i2p/go/src/github.com/eyedeekay/i2p-tools-1
RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y git golang-1.13-go make && \
    mkdir -p /var/lib/i2p/i2p-config/reseed && \
    chown -R $I2P_UID:$I2P_GID /var/lib/i2p && chmod -R o+rwx /var/lib/i2p
RUN /usr/lib/go-1.13/bin/go build -v -tags netgo -ldflags '-w -extldflags "-static"'
USER $I2P_UID
VOLUME /var/lib/i2p/i2p-config/reseed
WORKDIR /var/lib/i2p/i2p-config/reseed
ENTRYPOINT [ "/var/lib/i2p/go/src/github.com/eyedeekay/i2p-tools-1/i2p-tools-1", "reseed", "--yes=true", "--netdb=/var/lib/i2p/i2p-config/netDb" ]