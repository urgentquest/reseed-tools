
## Example Commands:

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