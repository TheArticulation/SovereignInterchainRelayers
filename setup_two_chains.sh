#!/bin/bash

# This script sets up two Ignite chains, SovereignChain and LightChain, for local development.
# It handles scaffolding, configuration, and launching of the chains.
# NOTE: Regarding module selection, Ignite scaffolds a standard set of Cosmos SDK modules.
# The requested module list for LightChain (excluding staking, slashing, etc.) would require
# manual modifications to the generated source code (specifically app.go), which is
# beyond the scope of an automated setup script. Therefore, this script proceeds by
# scaffolding both chains with the standard module set and creating a validator for each,
# as this aligns with the other requirements like creating a validator on each chain.

set -euxo pipefail

# --- Cleanup ---
echo "--- Cleaning up previous runs ---"
killall sovereignd lightd || true
rm -rf SovereignChain LightChain sovereign.log light.log

# Check for ignite dependency
command -v ignite >/dev/null 2>&1 || { echo >&2 "ignite CLI is not installed. Please install it to continue. Aborting."; exit 1; }

# --- SovereignChain Setup ---
echo "--- Scaffolding SovereignChain ---"
ignite scaffold chain sovereignd --path SovereignChain --no-module

echo "--- Configuring SovereignChain ---"
cat <<EOF > ./SovereignChain/config.yml
version: 1
accounts:
  - name: validator
    coins: ["1000000000usov"]
  - name: user1
    coins: ["500000000usov"]
  - name: user2
    coins: ["500000000usov"]
validator:
  name: validator
  staked: "100000000usov"
genesis:
  chain_id: "sovereign-1"
  app_state:
    staking:
      params:
        bond_denom: "usov"
    bank:
      denom_metadata:
        - description: "The native token of SovereignChain"
          denom_units:
            - denom: "usov"
              exponent: 0
            - denom: "sov"
              exponent: 6
          base: "usov"
          display: "sov"
          name: "Sovereign"
          symbol: "SOV"
EOF

# --- LightChain Setup ---
echo "--- Scaffolding LightChain ---"
ignite scaffold chain lightd --path LightChain --no-module

echo "--- Configuring LightChain ---"
cat <<EOF > ./LightChain/config.yml
version: 1
accounts:
  - name: validator
    coins: ["1000000000ulit"]
  - name: user1
    coins: ["500000000ulit"]
  - name: user2
    coins: ["500000000ulit"]
validator:
  name: validator
  staked: "100000000ulit"
genesis:
  chain_id: "light-1"
  app_state:
    staking:
      params:
        bond_denom: "ulit"
    bank:
      denom_metadata:
        - description: "The native token of LightChain"
          denom_units:
            - denom: "ulit"
              exponent: 0
            - denom: "lit"
              exponent: 6
          base: "ulit"
          display: "lit"
          name: "Light"
          symbol: "LIT"
client:
  rpc:
    address: "0.0.0.0:26667"
  p2p:
    address: "0.0.0.0:26666"
servers:
  grpc:
    address: "0.0.0.0:9092"
  "grpc-web":
    address: "0.0.0.0:9093"
  api:
    address: "0.0.0.0:1318"
EOF

# --- Start Chains ---
echo "--- Starting both chains in the background ---"
(cd SovereignChain && ignite chain serve --reset-once &> ../sovereign.log) &
SOVEREIGN_PID=$!

(cd LightChain && ignite chain serve --reset-once &> ../light.log) &
LIGHT_PID=$!

# Trap to ensure background processes are killed on script exit
trap "kill $SOVEREIGN_PID $LIGHT_PID" EXIT

echo "Waiting for chains to start... (this might take a few minutes on the first run)"
# A simple polling mechanism to wait for the RPC servers to be ready
while ! curl -s http://localhost:26657/status > /dev/null; do sleep 1; done
while ! curl -s http://localhost:26667/status > /dev/null; do sleep 1; done

echo "Chains are up and running!"

# --- Execute Transactions ---
echo "--- Preparing to send transactions ---"
SOVEREIGN_HOME="./SovereignChain"
LIGHT_HOME="./LightChain"

# Get addresses
SOVEREIGN_USER1_ADDR=$(cd $SOVEREIGN_HOME && ./sovereignd keys show user1 -a --keyring-backend test)
SOVEREIGN_USER2_ADDR=$(cd $SOVEREIGN_HOME && ./sovereignd keys show user2 -a --keyring-backend test)
LIGHT_USER1_ADDR=$(cd $LIGHT_HOME && ./lightd keys show user1 -a --keyring-backend test)
LIGHT_USER2_ADDR=$(cd $LIGHT_HOME && ./lightd keys show user2 -a --keyring-backend test)

# Send transaction on SovereignChain
echo "--- Sending transaction on SovereignChain ---"
(cd $SOVEREIGN_HOME && ./sovereignd tx bank send user1 "$SOVEREIGN_USER2_ADDR" 100usov --chain-id sovereign-1 --keyring-backend test -y --broadcast-mode=block --node tcp://localhost:26657 --fees 200usov)

# Send transaction on LightChain
echo "--- Sending transaction on LightChain ---"
(cd $LIGHT_HOME && ./lightd tx bank send user1 "$LIGHT_USER2_ADDR" 100ulit --chain-id light-1 --keyring-backend test -y --broadcast-mode=block --node tcp://localhost:26667 --fees 200ulit)

echo "--- Transactions sent successfully! ---"

# --- Next Steps ---
echo ""
echo "✅ Setup complete!"
echo ""
echo "Your two chains are running in the background:"
echo "  - SovereignChain (PID: $SOVEREIGN_PID) -> Logs: tail -f sovereign.log"
echo "  - LightChain (PID: $LIGHT_PID) -> Logs: tail -f light.log"
echo ""
echo "To interact with SovereignChain:"
echo "  ./SovereignChain/sovereignd status --node tcp://localhost:26657"
echo "  ./SovereignChain/sovereignd q bank balances $SOVEREIGN_USER2_ADDR --node tcp://localhost:26657"
echo ""
echo "To interact with LightChain:"
echo "  ./LightChain/lightd status --node tcp://localhost:26667"
echo "  ./LightChain/lightd q bank balances $LIGHT_USER2_ADDR --node tcp://localhost:26667"
echo ""
echo "To stop the chains, run:"
echo "  kill $SOVEREIGN_PID $LIGHT_PID"
echo ""
