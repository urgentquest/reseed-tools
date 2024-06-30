Upgrading from an older version of reseed-tools
===============================================

This reseed server sometimes gains helpful features that reseed operators may wish to use.
Additionally, it is possible that at some point we'll need to release a security update.
This document provides a path to upgrade the various binary distributions of reseed-tools.

Debian and Ubuntu Users
-----------------------

1. Shut down the existing `reseed-tools` service.
 If you are using `sysvinit` or something like it, you should be able to run: `sudo service reseed stop`.
 If you are using `systemd` you should be able to run `sudo systemctl stop reseed`.
 If those commands don't work, use `killall reseed-tools`
2. Download the `.deb` package from the Github Releases page.
 Make sure you get the right package for your ARCH/OS pair.
 Most will need the `_amd64.deb` package.
3. Install the package using: `sudo dpkg -i ./reseed-tools*.deb`

Docker Users
------------

1. Build the container locally: `docker build -t eyedeekay/reseed .`
2. Stop the container: `docker stop reseed`
3. Start the container: `docker start reseed`

Freestanding `tar.gz` Users, People who built from source
---------------------------------------------------------

1. Shut down the existing `reseed-tools` service.
 If you are using `sysvinit` or something like it, you should be able to run: `sudo service reseed stop`.
 If you are using `systemd` you should be able to run `sudo systemctl stop reseed`.
 If those commands don't work, use `killall reseed-tools`
2. Extract the tar file: `tar xzf reseed-tools.tgz`
3. Copy the `reseed-tools` binary to the correct location if you're on `amd64` or compile it if you are not.
 `cp reseed-tools reseed-tools-linux-amd64`
 OR
 `make build`
4. Install the new software and service management files:
 `sudo make install`
