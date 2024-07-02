Using a remote Network Database with SSH
========================================

Beginning in `reseed-tools 2.5.0` it is possible to use reseed-tools to "share" a netDb directory on one host with a reseed server on another host.
This feature is built into the reseed-tools software.
It is also possible to do this manually using `sshfs`, `ssh` combined with `cron`, and most available backup utilities like `borg` and `syncthing`.
This guide only covers `rsync+ssh` and `cron` where I2P is running as a user(not as `i2psvc`).
It requires 2 hosts with exposed SSH ports that can reach eachother.
It also pretty much assumes you're using something based on Debian.

Why?
----

In most setups, a reseed service is using a network database which is kept on the same server as the I2P router where it finds it's netDb.
This is convenient, however if reseed servers are targeted for a RouterInfo spam attack, then the reseed server could potentially be overwhelmed with spammy routerInfos.
That impairs a new user's ability to join the network and slows down network integration.

SSH-Protected Retrieval of NetDB content over I2P
-----------------------------------------------

In this guide, the NetDB is retrieved from a remote router by the reseed server.

### On the Remote Router

Install openssh-server and rsync and enable the service:

```sh
sudo apt install openssh-server rsync
sudo systemctl enable ssh
```

### On the Reseed Server

Set up SSH and generate new keys, without passwords:

```sh
ssh-keygen -f ~/.ssh/netdb_sync_ed25519 -N ""
```

Then, copy the keys to the remote router:

```sh
ssh-copy-id -f ~/.ssh/netdb_sync_ed25519 $(UserRunningI2P)@$(RemoteRouter)
```

After, set up the `cron` job to copy the netDB.

```sh
crontab -e
>>
* 30 * * *  rsync --ignore-existing -raz $(UserRunningI2P)@$(RemoteRouter):$(/Path/To/Remote/NetDB) $(Path/To/My/NetDB)
```

SSH-Protected Sharing of NetDB content over I2P
-----------------------------------------------

In this guide, the NetDB is pushed to a reseed server by a remote router.

### On the Reseed Server

Install openssh-server and rsync and enable the service:

```sh
sudo apt install openssh-server rsync
sudo systemctl enable ssh
```

Next, stop your reseed server.

```sh
killall reseed-tools
```

### On the Remote Router

Start by setting up SSH and generating new keys, without passwords:

```sh
ssh-keygen -f ~/.ssh/netdb_sync_ed25519 -N ""
```

Then, copy the keys to the Reseed Server:

```sh
ssh-copy-id -f ~/.ssh/netdb_sync_ed25519 $(UserRunningReseed)@$(ReseedServer)
```

After, set up the `cron` job to copy the netDB.

```sh
crontab -e
>>
* 30 * * *  rsync --ignore-existing -raz $(/Path/To/My/NetDB) $(UserRunningReseed)@$(ReseedServer):/$(Path/To/Reseed/NetDB)
```
