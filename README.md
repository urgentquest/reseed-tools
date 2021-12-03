I2P Reseed Tools
==================

This tool provides a secure and efficient reseed server for the I2P network. There are several utility commands to
create, sign, and validate SU3 files. Please note that this requires at least Go version 1.13, and uses Go Modules.

## Dependencies

`go`, `git`, and optionally `make` are required to build the project.
Precompiled binaries for most platforms are available at my github mirror
https://github.com/eyedeekay/i2p-tools-1.

In order to install the build-dependencies on Ubuntu or Debian, you may use:

```sh
sudo apt-get install golang-go git make
```

## Installation(From Source)

```
git clone https://i2pgit.org/idk/reseed-tools
cd reseed-tools
make build
# Optionally, if you want to install to /usr/bin/reseed-tools
sudo make install
```

## Usage

#### Debian/Ubuntu note:

Debian users who are running I2P as a system service must also run the 
`reseed-tools` as the same user. This is so that the reseed-tools can access
the I2P service's netDb directory. On Debian and Ubuntu, that user is `i2psvc`
and the netDb directory is: `/var/lib/i2p/i2p-config/netDb`.

##### Systemd Service

A systemd service is provided which should work with the I2P Debian package
when reseed-tools is installed in `/usr/bin/reseed-tools`. If you install with
`make install` this service is also installed. This service will cause the
bundles to regenerate every 12 hours.

The contact email for your reseed should be added in:
`/etc/systemd/system/reseed.d/reseed.conf`.

Self-signed certificates will be auto-generated for these services. To change
this you should edit the `/etc/systemd/system/reseed.d/reseed.service`.

- To enable starting the reseed service automatically with the system: `sudo systemctl enable reseed.service`
- To run the service manually: `sudo sysctl start reseed.service`  
- To reload the systemd services: `sudo systemctl daemon-reload`
- To view the status/logs: `sudo journalctl -u reseed.service`

##### SysV Service

An initscript is also provided. The initscript, unlike the systemd service,
cannot schedule itself to restart. You should restart the service roughly once
a day to ensure that the information does not expire.

The contact email for your reseed should be added in:
`/etc/init.d/reseed`.

Self-signed certificates will be auto-generated for these services. To change
this you should edit the `/etc/init.d/reseed`.

### Without a webserver, standalone with TLS support

If this is your first time running a reseed server (ie. you don't have any existing keys),
you can simply run the command and follow the prompts to create the appropriate keys, crl and certificates.
Afterwards an HTTPS reseed server will start on the default port and generate 6 files in your current directory
(a TLS key, certificate and crl, and a su3-file signing key, certificate and crl).

```
reseed-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --tlsHost=your-domain.tld
```

## Example Commands:

### Locally behind a webserver (reverse proxy setup), preferred:

If you are using a reverse proxy server it may provide the TLS certificate instead.

```
reseed-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --port=8443 --ip=127.0.0.1 --trustProxy
```

### Without a webserver, standalone, self-supervising(Automatic restarts)

```
./reseed-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --littleboss=start
```

### Without a webserver, standalone, automatic OnionV3 with TLS support

```
./reseed-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --onion --i2p --p2p
```

### Without a webserver, standalone, serve P2P with LibP2P

```
./reseed-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --p2p
```

### Without a webserver, standalone, upload a single signed .su3 to github

* This one isn't working yet, I'll get to it eventually, I've got a cooler idea now.

```
./reseed-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --github --ghrepo=reseed-tools --ghuser=eyedeekay
```

### Without a webserver, standalone, in-network reseed

```
./reseed-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --i2p
```

### Without a webserver, standalone, Regular TLS, OnionV3 with TLS

```
./reseed-tools reseed --tlsHost=your-domain.tld --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --onion
```

### Without a webserver, standalone, Regular TLS, OnionV3 with TLS, and LibP2P

```
./reseed-tools reseed --tlsHost=your-domain.tld --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --onion --p2p
```

### Without a webserver, standalone, Regular TLS, OnionV3 with TLS, I2P In-Network reseed, and LibP2P, self-supervising

```
./reseed-tools reseed --tlsHost=your-domain.tld --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --onion --p2p --littleboss=start
```

### Docker

To make it easier to deploy reseeds, it is possible to run this software as a
Docker image. Because the software requires access to a network database to host
a reseed, you will need to mount the netDb as a volume inside your docker
container to provide access to it, and you will need to run it as the same user
and group inside the container as I2P.

When you run a reseed under Docker in this fashion, it will automatically
generate a self-signed certificate for your reseed server in a Docker volume
mamed reseed-keys. *Back up this directory*, if it is lost it is impossible
to reproduce.

Please note that Docker is not currently compatible with .onion reseeds unless
you pass the --network=host tag.

#### If I2P is running as your user, do this:

        docker run -itd \
            --name reseed \
            --publish 443:8443 \
            --restart always \
            --volume $HOME/.i2p/netDb:$HOME/.i2p/netDb:z \
            --volume reseed-keys:/var/lib/i2p/i2p-config/reseed \
            eyedeekay/reseed \
                --signer $YOUR_EMAIL_HERE

#### If I2P is running as another user, do this:

        docker run -itd \
            --name reseed \
            --user $(I2P_UID) \
            --group-add $(I2P_GID) \
            --publish 443:8443 \
            --restart always \
            --volume /PATH/TO/USER/I2P/HERE/netDb:/var/lib/i2p/i2p-config/netDb:z \
            --volume reseed-keys:/var/lib/i2p/i2p-config/reseed \
            eyedeekay/reseed \
                --signer $YOUR_EMAIL_HERE

#### **Debian/Ubuntu and Docker**

In many cases I2P will be running as the Debian system user ```i2psvc```. This
is the case for all installs where Debian's Advanced Packaging Tool(apt) was
used to peform the task. If you used ```apt-get install``` this command will
work for you. In that case, just copy-and-paste:

        docker run -itd \
            --name reseed \
            --user $(id -u i2psvc) \
            --group-add $(id -g i2psvc) \
            --publish 443:8443 \
            --restart always \
            --volume /var/lib/i2p/i2p-config/netDb:/var/lib/i2p/i2p-config/netDb:z \
            --volume reseed-keys:/var/lib/i2p/i2p-config/reseed \
            eyedeekay/reseed \
                --signer $YOUR_EMAIL_HERE
