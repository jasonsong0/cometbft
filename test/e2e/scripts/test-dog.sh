#!/usr/bin/env bash

set -e

MANIFEST=${1:-networks/200-nodes-dog.toml}
# MANIFEST=networks/7-nodes.toml

LOAD=200 # total load is $LOAD * $CONN
CONN=2
TARGET_REDUNDANCIES=(0.1 0.5 1)
ADJUST_INTERVALS=("500ms" "1000ms" "2000ms")
TEST_DURATION=1800 # 30 min
# TEST_DURATION=1200 # 20 min
# TEST_DURATION=600 # 10 min

# run once
make runner
./build/runner -f $MANIFEST -t DO infra create --yes

# to be able to run `runner load` from CC
TESTNET_DIR=${MANIFEST%.toml}
CC_ADDR=$(cat $TESTNET_DIR/.cc-ip)
ssh_cc() {
    ssh -o LogLevel=ERROR -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null root@$CC_ADDR -t "$@"
}
./scripts/upload-runner.sh $MANIFEST

run_instance() {
    # Wait a minimum of 90 seconds between runs; in the meantime, setup and start next testnet.
    sleep 90 & # sleep in background
    ./build/runner -f $MANIFEST -t DO setup --clean --le=false --keep-address-book
    ./build/runner -f $MANIFEST -t DO start
    wait # until sleeping has finished
    
    # Keep laptop awake while loading (only MacOS)
    caffeinate -u -t $TEST_DURATION &

    # Perturb nodes in parallel.
    sleep 30 && gtimeout $TEST_DURATION ./build/runner -f $MANIFEST -t DO perturb &
    
    # load txs from CC
    ssh_cc ./build/runner -f $MANIFEST -t DO load -r $LOAD -c $CONN -T $TEST_DURATION --internal-ip
    
    ./build/runner -f $MANIFEST -t DO stop
}

# Baseline (DOG disabled)
sed -i.bak -e "s/.*enable_dog_protocol.*/\t\"mempool.enable_dog_protocol = false\",/g" $MANIFEST 
run_instance
sleep 60

# DOG enabled
sed -i.bak -e "s/.*enable_dog_protocol.*/\t\"mempool.enable_dog_protocol = true\",/g" $MANIFEST 
for TARGET in ${TARGET_REDUNDANCIES[@]}; do
    # change target manifest
    sed -i.bak -e "s/.*target_redundancy .*/\t\"mempool.target_redundancy = $TARGET\",/g" $MANIFEST 
    for INTERVAL in ${ADJUST_INTERVALS[@]}; do
        sed -i.bak -e "s/.*adjust_redundancy_interval .*/\t\"mempool.adjust_redundancy_interval = $INTERVAL\",/g" $MANIFEST 
        echo 🟢 TARGET: $TARGET, INTERVAL: $INTERVAL
        run_instance
    done
    sleep 60
done

# TODO: download prometheus data (it may be huge)

# Remove CC from terraform's state, so it does not get destroyed later.
cd terraform
terraform state rm digitalocean_droplet.cc
cd ..

## Be careful! This will destroy CC with the prometheus data!
./build/runner -f $MANIFEST -t DO clean
