#!/bin/bash
set -euo pipefail

# This script demonstrates the enforcement mechanism of the duty module.
# It simulates a validator failing its duty and getting slashed, then recovery.

echo "--- 1. Dispatching Hyperlane Message Sovereign->Light ---"
# This assumes a command exists to dispatch a message. The specifics would
# depend on the Hyperlane module's implementation. We'll simulate a dispatch
# that creates a duty for message ID 1.
MSG_ID=1 
sovereignd tx dispatch "Sovereign" "Light" "Hello from Sovereign" --from user1 --chain-id sovereign-1 --node tcp://localhost:26657 -y --broadcast-mode=block
echo "Message dispatched. A new duty should be created."
sleep 5 # Wait for the next block

echo "--- 2. Show Duty Assignment ---"
# Get the validator address (assuming only one validator)
VAL_ADDR=$(sovereignd keys show validator -a --keyring-backend test)
ASSIGNMENT=$(sovereignd q duty assignments --val "$VAL_ADDR" --node tcp://localhost:26657 -o json)
echo "Current assignment for $VAL_ADDR:"
echo "$ASSIGNMENT" | jq

DEADLINE_HEIGHT=$(echo "$ASSIGNMENT" | jq -r '.deadline_height')
if [ -z "$DEADLINE_HEIGHT" ] || [ "$DEADLINE_HEIGHT" == "null" ]; then
    echo "Error: Could not find duty assignment for validator $VAL_ADDR."
    exit 1
fi
echo "Duty deadline is at block height: $DEADLINE_HEIGHT"

echo "--- 3. Kill Duty-Agent for Validator ---"
# This assumes the duty-agent's PID is stored in a file.
AGENT_PID_FILE="/tmp/duty-agent-${VAL_ADDR}.pid"
if [ ! -f "$AGENT_PID_FILE" ]; then
    echo "Error: PID file for duty agent not found at $AGENT_PID_FILE."
    echo "Please ensure the agent is running and its PID is stored correctly."
    exit 1
fi
AGENT_PID=$(cat "$AGENT_PID_FILE")
echo "Killing duty-agent with PID $AGENT_PID..."
kill "$AGENT_PID"
rm "$AGENT_PID_FILE"
echo "Agent killed. The validator will now miss its duty."

echo "--- 4. Wait Until Deadline Has Passed ---"
TARGET_HEIGHT=$((DEADLINE_HEIGHT + 2))
echo "Waiting to reach block height $TARGET_HEIGHT..."
while true; do
    CURRENT_HEIGHT=$(sovereignd status --node tcp://localhost:26657 | jq -r '.SyncInfo.latest_block_height')
    echo "Current height: $CURRENT_HEIGHT"
    if [ "$CURRENT_HEIGHT" -ge "$TARGET_HEIGHT" ]; then
        echo "Deadline passed."
        break
    fi
    sleep 2
done

echo "--- 5. Submit Missed Duty Report ---"
# The proofs are placeholders here. A real implementation would require complex
# cryptographic proofs of message inclusion on the origin and non-inclusion on the destination.
REPORTER_ADDR=$(sovereignd keys show user2 -a --keyring-backend test)
sovereignd tx duty report-missed \
    "Sovereign" "Light" "$MSG_ID" "$VAL_ADDR" \
    "origin_proof_placeholder" "dest_non_inclusion_proof_placeholder" \
    --from "$REPORTER_ADDR" --chain-id sovereign-1 --node tcp://localhost:26657 -y --broadcast-mode=block
echo "Missed duty report submitted."
sleep 5 # Wait for the next block

echo "--- 6. Show Slashing on SovereignChain ---"
VAL_CONS_ADDR=$(sovereignd q staking validator "$VAL_ADDR" -o json | jq -r '.consensus_pubkey.key')
echo "Querying slashing info for validator consensus address: $VAL_CONS_ADDR"
sovereignd q slashing signing-info "$VAL_CONS_ADDR" --node tcp://localhost:26657
echo "Note the 'jailed_until' and 'missed_blocks_counter' fields."

echo "--- 7. Restart Duty-Agent and Confirm Recovery ---"
echo "Restarting duty-agent in the background..."
# This is a placeholder for the actual command to start the agent
./agents/duty-agent/duty-agent --val-addr "$VAL_ADDR" &
echo $! > "$AGENT_PID_FILE"
echo "Agent restarted with new PID $(cat $AGENT_PID_FILE)."

# Unjail the validator
sovereignd tx slashing unjail --from validator --chain-id sovereign-1 --node tcp://localhost:26657 -y --broadcast-mode=block
sleep 5

echo "Dispatching a new message to confirm the system is working..."
sovereignd tx dispatch "Sovereign" "Light" "Hello again from Sovereign" --from user1 --chain-id sovereign-1 --node tcp://localhost:26657 -y --broadcast-mode=block
sleep 10 # Give the agent time to process

echo "Checking assignments (should be empty or new)..."
sovereignd q duty assignments --val "$VAL_ADDR" --node tcp://localhost:26657

echo "--- 8. Query LightChain to Show Sovereign Set Enforcement ---"
echo "Querying the current validator set root on LightChain..."
lightd q sovereignclient current-set --node tcp://localhost:26667
echo "This confirms LightChain is aware of the SovereignChain validator set."

echo "✅ Demo complete!"
