package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// fetchTransactionBySignature fetches and prints details for a given tx signature

// Base64 decoding -> strongly typed Go struct
func fetchTransactionBySignatureBase64(ctx context.Context, client *rpc.Client, sigStr string) {
	sig, err := solana.SignatureFromBase58(sigStr)
	if err != nil {
		log.Fatalf("invalid signature: %v", err)
	}

	tx, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding:   solana.EncodingBase64,   // SAFE for Go struct
		Commitment: rpc.CommitmentConfirmed, // or rpc.CommitmentFinalized
	})
	if err != nil {
		log.Fatalf("GetTransaction error: %v", err)
	}
	if tx == nil {
		fmt.Println("Transaction not found (maybe too old or pruned).")
		return
	}

	status := "ok"
	if tx.Meta != nil && tx.Meta.Err != nil {
		status = fmt.Sprintf("err:%v", tx.Meta.Err)
	}
	fmt.Printf("\n=== Transaction %s ===\n", sig.String())
	fmt.Printf("Slot: %d | Status: %s\n", tx.Slot, status)

	if tx.Meta != nil {
		fmt.Printf("Fee: %d lamports\n", tx.Meta.Fee)
	}

	if tx.Meta != nil && len(tx.Meta.LogMessages) > 0 {
		fmt.Println("Logs:")
		for i, line := range tx.Meta.LogMessages {
			fmt.Printf("  log[%d]: %s\n", i, line)
		}
	}
	fmt.Println("=============================")
}

// JSON Parsed decoding -> human readable, not strongly typed
func fetchTransactionBySignatureParsed(ctx context.Context, client *rpc.Client, sigStr string) {
	// Direct RPC call to getTransaction with jsonParsed
	// Call RPC and decode result directly into raw map
	var raw map[string]interface{}
	if err := client.RPCCallForInto(
		ctx,
		&raw,
		"getTransaction",
		[]interface{}{
			sigStr,
			map[string]interface{}{
				"encoding":   "jsonParsed",
				"commitment": "confirmed",
			},
		},
	); err != nil {
		log.Fatalf("getTransaction jsonParsed error: %v", err)
	}

	pretty, _ := json.MarshalIndent(raw, "", "  ")
	fmt.Printf("\n=== Transaction %s (jsonParsed) ===\n", sigStr)
	fmt.Println(string(pretty))
	fmt.Println("=============================")
}
