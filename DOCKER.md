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