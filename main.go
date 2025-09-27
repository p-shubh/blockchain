package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/joho/godotenv"
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
		// ja := 0
		// For each signature, fetch the transaction
		for _, s := range sigs {
			// if ja > 0 {
			// 	return nil
			// }
			sig := s.Signature
			maxVer := uint64(0) // support legacy + v0 transactions
			tx, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
				Commitment:                     rpc.CommitmentConfirmed,
				MaxSupportedTransactionVersion: &maxVer,
			})
			if err != nil {
				// fmt.Printf("sig=%s slot=%d status=%s\n", sig.String(), 0, "error_fetching_tx")
				// fmt.Printf("Error sig=%s fetching transaction: Details = %s\n", sig.String(), err)

				continue
			}

			handleTx(sig, tx)
			// ja++
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

	client := rpc.New(os.Getenv("SOLANA_RPC_URL"))

	// Example: Solana Program address (replace with your wallet/program)
	// addrStr := "6TH6UY8FHmrm9Qsvzf45qb3mEjE8HVfu7wTuoZ4UHQMr"
	addrStr := "dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN"
	if len(addrStr) > 1 && addrStr[:2] == "0x" {
		log.Fatalf("Provided address looks like Ethereum: %s. Use a Solana base58 address.", addrStr)
	}
	addr := solana.MustPublicKeyFromBase58(addrStr)

	err := fetchAllTransactionsForAddress(ctx, client, addr, func(sig solana.Signature, tx *rpc.GetTransactionResult) {
		// status := "ok"
		// if tx.Meta != nil && tx.Meta.Err != nil {
		// 	status = fmt.Sprintf("err:%v", tx.Meta.Err)
		// }
		slot := uint64(0)
		if tx != nil {
			slot = tx.Slot
		}

		// log.Printf("sig=%s slot=%d status=%s\n", sig.String(), slot, status)

		// print all logs if present
		if tx != nil && tx.Meta != nil && len(tx.Meta.LogMessages) > 0 {
			for _, line := range tx.Meta.LogMessages {
				// log.Printf("log[%d]: %s\n", i, line)
				d := fmt.Sprintf("%s\n", line)
				x := strings.Contains(d, "InitializeMint2")
				if x {
					// print the signature and slot as well with the log
					// log.Println("d:=", d)
					log.Printf("sig=%s slot=%d log: %s\n", sig.String(), slot, line)
				}
			}
		}
	})
	if err != nil {
		log.Fatalf("error fetching txs: %v", err)
	}
}
func callByTransactionId(txSig string) {

	ctx := context.Background()
	client := rpc.New(os.Getenv("SOLANA_RPC_URL"))

	// Example: replace with any valid transaction signature
	// txSig := "5Nx6V7B3oX7VfpXwP2wX6wE3X4BwnVAB7cMWMGBr7HgHmxrD8SuKQtu1LZ7iXWZb1fEzRb6TccE7W5XTxk3e5L8q"
	fetchTransactionBySignatureParsed(ctx, client, txSig)

}
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		return
	} // Load .env file if present

	// Fetch by address (all transactions, paginating)
	fmt.Println(time.Now())
	callByAddress()
	fmt.Println(time.Now())

	// Fetch by transaction ID (single transaction)
	// callByTransactionId("66tTrP3na79SdLEWGRWJRWosQQ5KoKNgzh9qtBEZUQu36RGhPkHgxzPaEe3SxuLZx6BeU3a7YKhrwBhTxrL5J5z8")
	// solanaContract()
}
