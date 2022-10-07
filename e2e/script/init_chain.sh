#!/bin/sh

MONIKER=testchainer
CHAIN_ID=testchain
KEYRING=test
HOME_PATH=/data/chain/.gaiad
WALLET_NAME=validator
FAUCET_ADDRESS=cosmos1u9vn33qs6jdr3wwq4u2l9p349n9c95uxz2lew0
RUN_CMD=gaiad

rm -rf $HOME_PATH
mkdir -p $HOME_PATH

# init all three validators
$RUN_CMD init $MONIKER --chain-id=$CHAIN_ID --home=$HOME_PATH

# create keys for all three validators
$RUN_CMD keys add $WALLET_NAME --keyring-backend=$KEYRING --home=$HOME_PATH

# create validator node with tokens to transfer to the three other nodes
$RUN_CMD add-genesis-account $($RUN_CMD keys show $WALLET_NAME -a --keyring-backend=$KEYRING --home=$HOME_PATH) 100000000000uatom,100000000000stake --home=$HOME_PATH
$RUN_CMD add-genesis-account $FAUCET_ADDRESS 100000000000uatom --home=$HOME_PATH
$RUN_CMD gentx $WALLET_NAME 500000000stake --keyring-backend=$KEYRING --home=$HOME_PATH --chain-id=$CHAIN_ID
$RUN_CMD collect-gentxs --home=$HOME_PATH

# validator
# enable rest api server & unsafe cors
sed -i -E 's|enable = false|enable = true|g' $HOME_PATH/config/app.toml
sed -i -E 's|enabled-unsafe-cors = false|enabled-unsafe-cors = true|g' $HOME_PATH/config/app.toml

# allow duplicate ip
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $HOME_PATH/config/config.toml
sed -i -E 's|tcp://127.0.0.1:26657|tcp://0.0.0.0:26657|g' $HOME_PATH/config/config.toml

$RUN_CMD start --home=$HOME_PATH
