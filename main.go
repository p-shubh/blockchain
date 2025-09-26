package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// fetchAllTransactionsForAddress fetches all signatures for `addr` (paginating)
// then fetches each transaction and calls `handleTx` with the signature and tx info.
func fetchAllTransactionsForAddress(
	ctx context.Context,
	client *rpc.Client,
	addr solana.PublicKey,
	handleTx func(sig solana.Signature, tx *rpc.GetTransactionResult),
) error {
	var before *solana.Signature // pagination cursor (use Signature pointer as required by RPC)
	limit := 1000

	for {
		// Get signatures batch (newest first)
		opts := &rpc.GetSignaturesForAddressOpts{
			Limit:      &limit,
			Commitment: rpc.CommitmentConfirmed, // try Confirmed for more results
		}
		if before != nil {
			opts.Before = *before
		}
		sigs, err := client.GetSignaturesForAddressWithOpts(ctx, addr, opts)
		if err != nil {
			return fmt.Errorf("GetSignaturesForAddressWithOpts: %w", err)
		}
		if len(sigs) == 0 {
			fmt.Println("No more signatures found.")
			return nil
		}
		ja := 0
		// For each signature, fetch the transaction
		for _, s := range sigs {
			if ja > 0 {
				return nil
			}
			sig := s.Signature
			tx, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
				Commitment: rpc.CommitmentConfirmed,
			})
			if err != nil {
				fmt.Printf("sig=%s slot=%d status=%s\n", sig.String(), 0, "error_fetching_tx")

				continue
			}

			handleTx(sig, tx)
			ja++
			time.Sleep(10 * time.Millisecond) // be kind to RPC
		}

		// pagination: set `before` to the last signature
		last := sigs[len(sigs)-1].Signature
		before = &last

		// If we got fewer than limit entries, we’re done
		if len(sigs) < int(limit) {
			return nil
		}
	}
}

func callByAddress() {
	ctx := context.Background()

	// Use the right cluster for your address
	// Mainnet:
	// client := rpc.New("https://api.mainnet-beta.solana.com")

	client := rpc.New("")

	// Example: Solana Program address (replace with your wallet/program)
	addrStr := ""
	if len(addrStr) > 1 && addrStr[:2] == "0x" {
		log.Fatalf("Provided address looks like Ethereum: %s. Use a Solana base58 address.", addrStr)
	}
	addr := solana.MustPublicKeyFromBase58(addrStr)

	err := fetchAllTransactionsForAddress(ctx, client, addr, func(sig solana.Signature, tx *rpc.GetTransactionResult) {
		status := "ok"
		if tx.Meta != nil && tx.Meta.Err != nil {
			status = fmt.Sprintf("err:%v", tx.Meta.Err)
		}
		slot := uint64(0)
		if tx != nil {
			slot = tx.Slot
		}
		fmt.Printf("sig=%s slot=%d status=%s\n", sig.String(), slot, status)

		// print all logs if present
		if tx != nil && tx.Meta != nil && len(tx.Meta.LogMessages) > 0 {
			for i, line := range tx.Meta.LogMessages {
				fmt.Printf("  log[%d]: %s\n", i, line)
			}
		}
	})
	if err != nil {
		log.Fatalf("error fetching txs: %v", err)
	}
}
func callByTransactionId(txSig string) {

	ctx := context.Background()
	client := rpc.New("")

	// Example: replace with any valid transaction signature
	// txSig := ""
	fetchTransactionBySignatureParsed(ctx, client, txSig)

}
func main() {
	// Fetch by address (all transactions, paginating)
	// callByAddress()

	// Fetch by transaction ID (single transaction)
	callByTransactionId("")
}
