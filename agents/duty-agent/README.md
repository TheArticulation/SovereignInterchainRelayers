# Duty Agent

The Duty Agent is a sidecar process for SovereignChain validators. It monitors the `x/duty` module for assigned message relay duties, executes a relayer binary for those duties, and sends periodic heartbeats to the chain to signal liveness.

## Configuration

The agent is configured via environment variables:

- `SOVEREIGN_RPC`: The RPC endpoint of a SovereignChain node (e.g., "tcp://localhost:26657").
- `SOVEREIGN_GRPC`: The gRPC endpoint of a SovereignChain node (e.g., "localhost:9090").
- `SOVEREIGN_CHAIN_ID`: The chain ID of the SovereignChain (e.g., "sovereign-1").
- `VAL_ADDR`: The `valoper` address of the validator this agent is running for (e.g., "cosmosvaloper1...").
- `RELAYER_KEY_PATH`: Path to the exported (armored) private key file for the relayer. This key is used to sign heartbeats.
- `RELAYER_BIN`: The path to the relayer binary executable.
- `HEARTBEAT_PERIOD`: The duration between heartbeats (e.g., "30s", "1m").

### Sample Configuration (.env file)

```env
SOVEREIGN_RPC="tcp://localhost:26657"
SOVEREIGN_GRPC="localhost:9090"
SOVEREIGN_CHAIN_ID="sovereign-1"
VAL_ADDR="cosmosvaloper1..."
RELAYER_KEY_PATH="./relayer_key.pem"
RELAYER_BIN="/usr/local/bin/relayer"
HEARTBEAT_PERIOD="30s"
```

## Running the Agent

First, build the agent:

```bash
make duty-agent
```

Then, run it with the environment variables configured:

```bash
# You can source an .env file
export $(cat .env | xargs)

./build/duty-agent
```

### Dry Run

To see the heartbeat transaction that would be sent without actually broadcasting it, use the `--dry-run` flag:

```bash
./build/duty-agent --dry-run
```

This will print the unsigned transaction JSON to the console.
