#!/usr/bin/env bash

if [ "$1" == "debug" ]; then
    set -x
fi

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

set -euo pipefail

echo ""
echo -n "Please enter your voter account address: "
# shellcheck disable=SC2034
read -r ACCOUNT <&1

echo ""
echo -n "Please enter your Pinata API key: "
# shellcheck disable=SC2034
read -r PINATA <&1


echo ""
echo -n "Please enter your keystore password (your keys will be imported later): "
# shellcheck disable=SC2034
read -r KEYSTORE_PASSWORD <&1

echo ""
echo -n "Use testnet settings [y/n]: "
# shellcheck disable=SC2034
read -r -n 1 TESTNET <&1

mkdir -p /blockchain/relay

echo "account = \"$ACCOUNT\"
trustNodeDepositAmount     = 1000000  # PLS
eth2EffectiveBalance       = 32000000 # PLS
maxPartialWithdrawalAmount = 8000000  # PLS
gasLimit = \"3000000\"
maxGasPrice = \"1200\"                            #Gwei
batchRequestBlocksNumber = 16
runForEntrustedLsdNetwork = false

[pinata]
apikey = \"$PINATA\"
pinDays = 180
" > /blockchain/relay/config.toml

if [[ $TESTNET =~ ^[Yy]$ ]]
then
echo "[contracts]
lsdTokenAddress = \"0x61135C59A4Eb452b89963188eD6B6a7487049764\"
lsdFactoryAddress = \"0x98f51f52A8FeE5a469d1910ff1F00A3D333bc9A6\"

[[endpoints]]
eth1 = \"https://rpc-testnet-pulsechain.g4mm4.io\"
eth2 = \"https://rpc-testnet-pulsechain.g4mm4.io/beacon-api/\"
" >> /blockchain/relay/config.toml
else
echo "[contracts]
lsdTokenAddress = \"0xLSD_TOKEN_ADDRESS\"
lsdFactoryAddress = \"0xLSD_FACTORY_ADDRESS\"

[[endpoints]]
eth1 = \"https://rpc-pulsechain.g4mm4.io\"
eth2 = \"https://rpc-pulsechain.g4mm4.io/beacon-api/\"
" >> /blockchain/relay/config.toml
fi


echo ""
echo "Created default config.toml"

# Set the keystore to be readable by the relay docker user
chown -R 65532:65532 /blockchain/relay

# Add Docker's official GPG key:
apt-get update
apt-get install ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get -qq update

apt-get -qq install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin unattended-upgrades apt-listchanges

arch=$(uname -i)
if [[ $arch == arm* ]] || [[ $arch == aarch* ]]; then
    apt-get -qq -y install binfmt-support qemu-user-static
fi


echo unattended-upgrades unattended-upgrades/enable_auto_updates boolean true | debconf-set-selections
dpkg-reconfigure -f noninteractive unattended-upgrades
echo 'Unattended-Upgrade::Automatic-Reboot "true";' >> /etc/apt/apt.conf.d/50unattended-upgrades

docker run --detach \
    --name watchtower \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    containrrr/watchtower

docker run --name relay -it --restart always -v /blockchain/relay:/keys ghcr.io/vouchrun/pls-lsd-relay:main import-account --base-path /keys

docker run --name relay -it -e KEYSTORE_PASSWORD --restart always -v /blockchain/relay:/keys ghcr.io/vouchrun/pls-lsd-relay:main start --base-path /keys