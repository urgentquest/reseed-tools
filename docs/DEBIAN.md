# Debian and Ubuntu Packages

It's possible to generate a package which is compatible with Debian and Ubuntu,
using the command:

```sh

make checkinstall
sudo apt-get install ./reseed-tools_0.2.30-1_amd64.deb
```

This requires you to have `fakeroot` and `checkinstall` installed. Use the command

```sh

sudo apt-get install fakeroot checkinstall
```

to install them.
