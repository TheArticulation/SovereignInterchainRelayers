package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"

	// Assuming the duty module types are available in this path
	dutymoduletypes "sovereign/x/duty/types"

	"sovereign/app" // Assuming app is at this path for encoding config
)

func main() {
	// --- Configuration ---
	dryRun := flag.Bool("dry-run", false, "Output JSON of transaction without sending")
	flag.Parse()

	// Read environment variables
	config := readEnvConfig()
	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	valAddr, err := sdk.ValAddressFromBech32(config.ValAddr)
	if err != nil {
		log.Fatalf("Invalid validator address: %v", err)
	}

	// --- Main Loop ---
	log.Printf("Starting duty agent for validator %s", config.ValAddr)
	log.Printf("Heartbeat period: %s", config.HeartbeatPeriod)

	ticker := time.NewTicker(config.HeartbeatPeriod)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Ticker event: checking for duties and sending heartbeat...")

		// --- Query for Assignments (Placeholder) ---
		// In a real implementation, you would query the chain for assigned duties.
		// For this example, we'll simulate finding a duty.
		log.Println("Querying for duty assignments...")
		// assignments, err := queryDuties(config, valAddr)
		// if err != nil {
		// 	log.Printf("Error querying duties: %v", err)
		// 	continue
		// }
		// if len(assignments) == 0 {
		// 	log.Println("No new duties assigned.")
		// }

		// --- Execute Relayer for each assignment (Placeholder) ---
		// for _, duty := range assignments {
		// 	log.Printf("Executing relayer for duty: MsgID %d on Route %s->%s", duty.MsgID, duty.Route.Origin, duty.Route.Destination)
		// 	cmd := exec.Command(config.RelayerBin, "--route", duty.Route.Origin+"-"+duty.Route.Destination, "--msg-id", fmt.Sprintf("%d", duty.MsgID))
		// 	output, err := cmd.CombinedOutput()
		// 	if err != nil {
		// 		log.Printf("Relayer execution failed for MsgID %d: %v. Output: %s", duty.MsgID, err, string(output))
		// 	} else {
		// 		log.Printf("Relayer executed successfully for MsgID %d. Output: %s", duty.MsgID, string(output))
		// 	}
		// }

		// --- Send Heartbeat ---
		if err := sendHeartbeat(config, valAddr, *dryRun); err != nil {
			log.Printf("Error sending heartbeat: %v", err)
		}
	}
}

type Config struct {
	SovereignRPC      string
	SovereignGRPC     string
	SovereignChainID  string
	ValAddr           string
	RelayerKeyPath    string
	RelayerBin        string
	HeartbeatPeriod   time.Duration
}

func readEnvConfig() Config {
	periodStr := os.Getenv("HEARTBEAT_PERIOD")
	period, err := time.ParseDuration(periodStr)
	if err != nil {
		log.Printf("Invalid HEARTBEAT_PERIOD '%s', defaulting to 30s. Error: %v", periodStr, err)
		period = 30 * time.Second
	}

	return Config{
		SovereignRPC:      os.Getenv("SOVEREIGN_RPC"),
		SovereignGRPC:     os.Getenv("SOVEREIGN_GRPC"),
		SovereignChainID:  os.Getenv("SOVEREIGN_CHAIN_ID"),
		ValAddr:           os.Getenv("VAL_ADDR"),
		RelayerKeyPath:    os.Getenv("RELAYER_KEY_PATH"),
		RelayerBin:        os.Getenv("RELAYER_BIN"),
		HeartbeatPeriod:   period,
	}
}

func validateConfig(c Config) error {
	if c.SovereignRPC == "" { return fmt.Errorf("SOVEREIGN_RPC must be set") }
	if c.SovereignGRPC == "" { return fmt.Errorf("SOVEREIGN_GRPC must be set") }
	if c.SovereignChainID == "" { return fmt.Errorf("SOVEREIGN_CHAIN_ID must be set") }
	if c.ValAddr == "" { return fmt.Errorf("VAL_ADDR must be set") }
	if c.RelayerKeyPath == "" { return fmt.Errorf("RELAYER_KEY_PATH must be set") }
	if c.RelayerBin == "" { return fmt.Errorf("RELAYER_BIN must be set") }
	return nil
}

// queryDuties would connect to gRPC and query the x/duty module.
// This is a placeholder as it requires the actual proto definitions to be compiled.
func queryDuties(config Config, valAddr sdk.ValAddress) ([]dutymoduletypes.Duty, error) {
	// Setup gRPC connection
	grpcConn, err := grpc.Dial(
		config.SovereignGRPC,
		grpc.WithInsecure(), // Use secure options in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC: %w", err)
	}
	defer grpcConn.Close()

	queryClient := dutymoduletypes.NewQueryClient(grpcConn)
	res, err := queryClient.Duties(context.Background(), &dutymoduletypes.QueryDutiesRequest{
		ValidatorAddress: valAddr.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("duty query failed: %w", err)
	}
	return res.Duties, nil
}

func sendHeartbeat(config Config, valAddr sdk.ValAddress, dryRun bool) error {
	log.Println("Building heartbeat transaction...")
	// In a real implementation, these heights would be queried from respective chains.
	originHeights := map[string]uint64{
		"light-1": 123,
		"other-2": 456,
	}
	originHeightsJSON, err := json.Marshal(originHeights)
	if err != nil {
		return fmt.Errorf("failed to marshal origin heights: %w", err)
	}

	// For signing, we'll create a temporary in-memory keyring and import the key.
	// This avoids needing a home directory or complex config.
	kr, err := keyring.New("duty-agent", keyring.BackendMemory, "", nil)
	if err != nil {
		return fmt.Errorf("failed to create keyring: %w", err)
	}
	keyBytes, err := os.ReadFile(config.RelayerKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read relayer key: %w", err)
	}
	// Assuming the key is armored (like `gaiad keys export`)
	if err := kr.ImportPrivKey("relayer", string(keyBytes), "password"); err != nil {
		return fmt.Errorf("failed to import private key: %w", err)
	}

	relayerKey, err := kr.Key("relayer")
	if err != nil {
		return fmt.Errorf("failed to get key from keyring: %w", err)
	}

	// This is a simplified signature. A real implementation should sign a canonical representation.
	sig, _, err := kr.Sign("relayer", []byte(originHeightsJSON), signing.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return fmt.Errorf("failed to sign heartbeat data: %w", err)
	}
	
	msg := dutymoduletypes.NewMsgHeartbeat(valAddr, string(originHeightsJSON), sig)

	// --- Transaction Building & Broadcasting ---
	encodingConfig := app.MakeEncodingConfig() // Using the app's encoding config
	clientCtx := client.Context{}.
		WithClient(nil). // We don't need a client for this part
		WithChainID(config.SovereignChainID).
		WithTxConfig(encodingConfig.TxConfig).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithKeyring(kr).
		WithFromAddress(relayerKey.GetAddress()).
		WithFromName("relayer")

	txf := tx.NewFactoryCLI(clientCtx, nil).
		WithGas(flags.DefaultGasLimit).
		WithFees("200usov") // Example fee

	// Build the transaction
	txb, err := tx.BuildUnsignedTx(txf, msg)
	if err != nil {
		return fmt.Errorf("failed to build unsigned tx: %w", err)
	}

	// Sign the transaction
	err = tx.Sign(txf, "relayer", txb, true)
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

	txJSON, err := clientCtx.TxConfig.TxJSONEncoder()(txb.GetTx())
	if err != nil {
		return fmt.Errorf("failed to encode tx to JSON: %w", err)
	}

	if dryRun {
		log.Println("Dry run enabled. Transaction JSON:")
		fmt.Println(string(txJSON))
		return nil
	}
	
	// Create a real client context for broadcasting
	rpcClient, err := client.NewClientFromNode(config.SovereignRPC)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	clientCtx = clientCtx.WithClient(rpcClient)

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return fmt.Errorf("failed to encode tx to bytes: %w", err)
	}
	
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return fmt.Errorf("broadcast failed: %w", err)
	}

	log.Printf("Heartbeat sent successfully! TxHash: %s", res.TxHash)
	return nil
}
