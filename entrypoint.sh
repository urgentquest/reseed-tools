#! /usr/bin/env sh

cd /var/lib/i2p/i2p-config/reseed

cp -r /var/lib/i2p/go/src/i2pgit.org/idk/reseed-tools/content ./content

/var/lib/i2p/go/src/i2pgit.org/idk/reseed-tools/reseed-tools reseed --yes=true --netdb=/var/lib/i2p/i2p-config/netDb $@
