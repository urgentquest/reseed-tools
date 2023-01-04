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

You can also use a binary release from: https://github.com/eyedeekay/reseed-tools/releases.


```sh

wget https://github.com/eyedeekay/reseed-tools/releases/download/v0.2.30/reseed-tools_0.2.30-1_amd64.deb
# Obtain the checksum from the release web page
echo "38941246e980dfc0456e066f514fc96a4ba25d25a7ef993abd75130770fa4d4d reseed-tools_0.2.30-1_amd64.deb" > SHA256SUMS
sha256sums -c SHA256SUMS
sudo apt-get install ./reseed-tools_0.2.30-1_amd64.deb
```

That is how you install a binary `.deb` package. Once it's installed, you can start it using `systemd`:

```sh
sudo systemctl start reseed
```

which will run the server in the background.
