I2P Reseed Tools
==================

This tool provides a secure and efficient reseed server for the I2P network. There are several utility commands to create, sign, and validate SU3 files.

## Installation

If you have go installed you can download, build, and install this tool with `go get`

```
go get github.com/MDrollette/i2p-tools
i2p-tools -h
```

## Usage

### Locally behind a webserver (reverse proxy setup), preferred:

```
i2p-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --port=8443 --ip=127.0.0.1 --trustProxy
```

### Without a webserver, standalone with TLS support

```
i2p-tools reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --tlsHost=your-domain.tld
```

If this is your first time running a reseed server (ie. you don't have any existing keys),
you can simply run the command and follow the prompts to create the appropriate keys, crl and certificates.
Afterwards an HTTPS reseed server will start on the default port and generate 6 files in your current directory
(a TLS key, certificate and crl, and a su3-file signing key, certificate and crl).

Get the source code here on github or a pre-build binary anonymously on

http://reseed.i2p/
http://j7xszhsjy7orrnbdys7yykrssv5imkn4eid7n5ikcnxuhpaaw6cq.b32.i2p/

also a short guide and complete tech info.

## Experimental, currently only available from eyedeekay/i2p-tools-1 fork

Requires ```go mod``` and at least go 1.13. To build the eyedeekay/i2p-tools-1
fork, from anywhere:

        git clone https://github.com/eyedeekay/i2p-tools-1
        cd i2p-tools-1
        make build

### Without a webserver, standalone, automatic OnionV3 with TLS support

```
./i2p-tools-1 reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --onion
```

### Without a webserver, standalone, serve P2P with LibP2P

```
./i2p-tools-1 reseed --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --p2p
```

### Without a webserver, standalone, Regular TLS, OnionV3 with TLS

```
./i2p-tools-1 reseed --tlsHost=your-domain.tld --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --onion
```

### Without a webserver, standalone, Regular TLS, OnionV3 with TLS, and LibP2P

```
./i2p-tools-1 reseed --tlsHost=your-domain.tld --signer=you@mail.i2p --netdb=/home/i2p/.i2p/netDb --onion --p2p
```
