#! /usr/bin/env sh

chown -R $(id -u) /var/lib/i2p/i2p-config/reseed
chmod -R o+rwx /var/lib/i2p/i2p-config/reseed

su -c - $id -u "/var/lib/i2p/go/src/github.com/eyedeekay/i2p-tools-1/i2p-tools-1 reseed --yes=true --netdb=/var/lib/i2p/i2p-config/netDb $@"