package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

// TransferSPLToken transfers SPL tokens using Token Program
func TransferSPLToken(
	ctx context.Context,
	client *rpc.Client,
	from solana.PrivateKey,
	to solana.PublicKey,
	mint solana.PublicKey,
	amount uint64,
) (solana.Signature, error) {

	// Derive Associated Token Accounts (ATA)
	fromATA, _, _ := solana.FindAssociatedTokenAddress(from.PublicKey(), mint)
	toATA, _, _ := solana.FindAssociatedTokenAddress(to, mint)

	// Build transfer instruction
	ix := token.NewTransferInstruction(
		amount,           // Amount in smallest unit
		fromATA,          // From ATA
		toATA,            // To ATA
		from.PublicKey(), // Owner
		nil,
	).Build()

	// Get recent blockhash
	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("blockhash error: %w", err)
	}

	// Build transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{ix},
		recent.Value.Blockhash,
		solana.TransactionPayer(from.PublicKey()),
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("tx build error: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if from.PublicKey().Equals(key) {
				return &from
			}
			return nil
		},
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("sign error: %w", err)
	}

	// Send transaction
	sig, err := client.SendTransactionWithOpts(
		ctx,
		tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		if rpcErr, ok := err.(*jsonrpc.RPCError); ok {
			return solana.Signature{}, fmt.Errorf("rpc error: %s", rpcErr)
		}
		return solana.Signature{}, fmt.Errorf("send error: %w", err)
	}

	return sig, nil
}

func solanaContract() {
	ctx := context.Background()

	// Load vars from env
	rpcURL := os.Getenv("SOLANA_RPC_URL")           // e.g. https://api.devnet.solana.com
	fromKey := os.Getenv("SOLANA_FROM_PRIVATE_KEY") // base58 private key
	toKey := os.Getenv("SOLANA_TO_PUBLIC_KEY")      // destination wallet
	mintAddr := os.Getenv("SPL_TOKEN_MINT")         // token mint address
	amount := uint64(1_000_000)                     // default: 1 token (if mint has 6 decimals)

	if rpcURL == "" || fromKey == "" || toKey == "" || mintAddr == "" {
		log.Fatal("❌ Missing one or more required env vars: SOLANA_RPC_URL, SOLANA_FROM_PRIVATE_KEY, SOLANA_TO_PUBLIC_KEY, SPL_TOKEN_MINT")
	}

	client := rpc.New(rpcURL)

	from, err := solana.PrivateKeyFromBase58(fromKey)
	if err != nil {
		log.Fatalf("invalid from private key: %v", err)
	}
	to := solana.MustPublicKeyFromBase58(toKey)
	mint := solana.MustPublicKeyFromBase58(mintAddr)

	// Call transfer
	sig, err := TransferSPLToken(ctx, client, from, to, mint, amount)
	if err != nil {
		log.Fatalf("transfer failed: %v", err)
	}

	fmt.Println("✅ Token transfer success! Tx Signature:", sig.String())
}
