package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewBlockchainServiceClient(conn)

	// Example 1: Get latest block
	fmt.Println("\n=== Getting Latest Block ===")
	latestBlock, err := client.GetLatestBlock(context.Background(), &GetLatestBlockRequest{})
	if err != nil {
		log.Printf("Error getting latest block: %v", err)
	} else {
		fmt.Printf("Latest block: Hash=%s, Height=%d, Transactions=%d\n",
			latestBlock.Block.Hash, latestBlock.Block.Height, len(latestBlock.Block.Transactions))
	}

	// Example 2: Add a transaction
	fmt.Println("\n=== Adding Transaction ===")
	transaction := &Transaction{
		Hash:      "tx_12345",
		From:      "Alice",
		To:        "Bob",
		Amount:    10.5,
		Timestamp: time.Now().Unix(),
		Signature: "signature_xyz",
	}

	addResp, err := client.AddTransaction(context.Background(), &AddTransactionRequest{
		Transaction: transaction,
	})
	if err != nil {
		log.Printf("Error adding transaction: %v", err)
	} else {
		fmt.Printf("Transaction added: %s\n", addResp.Message)
	}

	// Example 3: Get specific transaction
	fmt.Println("\n=== Getting Transaction ===")
	txResp, err := client.GetTransaction(context.Background(), &GetTransactionRequest{
		Hash: "tx_12345",
	})
	if err != nil {
		log.Printf("Error getting transaction: %v", err)
	} else if txResp.Found {
		fmt.Printf("Transaction found: From=%s, To=%s, Amount=%.2f\n",
			txResp.Transaction.From, txResp.Transaction.To, txResp.Transaction.Amount)
	} else {
		fmt.Println("Transaction not found")
	}

	// Example 4: Subscribe to blocks (streaming)
	fmt.Println("\n=== Subscribing to Blocks ===")
	subscribeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.SubscribeBlocks(subscribeCtx, &SubscribeBlocksRequest{})
	if err != nil {
		log.Printf("Error subscribing to blocks: %v", err)
		return
	}

	fmt.Println("Subscribed to blocks. Listening for new blocks...")
	
	// Receive blocks from stream
	for {
		block, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("No more blocks to receive")
			break
		}
		if err != nil {
			log.Printf("Error receiving block: %v", err)
			break
		}

		fmt.Printf("New block received: Hash=%s, Height=%d, Transactions=%d\n",
			block.Hash, block.Height, len(block.Transactions))
		
		// For demo purposes, only receive a few blocks
		break
	}

	// Example 5: Interactive client
	fmt.Println("\n=== Interactive Client ===")
	interactiveClient(client)
}

func interactiveClient(client BlockchainServiceClient) {
	fmt.Println("\nInteractive gRPC Client")
	fmt.Println("1. Get latest block")
	fmt.Println("2. Add transaction")
	fmt.Println("3. Get transaction")
	fmt.Println("4. Get block by hash")
	fmt.Println("5. Subscribe to blocks")
	fmt.Println("6. Exit")

	var choice int
	var hash string
	
	for {
		fmt.Print("\nEnter choice (1-6): ")
		fmt.Scan(&choice)

		switch choice {
		case 1:
			resp, err := client.GetLatestBlock(context.Background(), &GetLatestBlockRequest{})
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			fmt.Printf("Latest block: Hash=%s, Height=%d\n", resp.Block.Hash, resp.Block.Height)

		case 2:
			var from, to string
			var amount float64
			fmt.Print("Enter from address: ")
			fmt.Scan(&from)
			fmt.Print("Enter to address: ")
			fmt.Scan(&to)
			fmt.Print("Enter amount: ")
			fmt.Scan(&amount)

			transaction := &Transaction{
				Hash:      fmt.Sprintf("tx_%d", time.Now().Unix()),
				From:      from,
				To:        to,
				Amount:    amount,
				Timestamp: time.Now().Unix(),
				Signature: "manual_signature",
			}

			resp, err := client.AddTransaction(context.Background(), &AddTransactionRequest{
				Transaction: transaction,
			})
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			fmt.Printf("Result: %s\n", resp.Message)

		case 3:
			fmt.Print("Enter transaction hash: ")
			fmt.Scan(&hash)
			resp, err := client.GetTransaction(context.Background(), &GetTransactionRequest{Hash: hash})
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			if resp.Found {
				fmt.Printf("Transaction: From=%s, To=%s, Amount=%.2f\n",
					resp.Transaction.From, resp.Transaction.To, resp.Transaction.Amount)
			} else {
				fmt.Println("Transaction not found")
			}

		case 4:
			fmt.Print("Enter block hash: ")
			fmt.Scan(&hash)
			resp, err := client.GetBlock(context.Background(), &GetBlockRequest{Hash: hash})
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			if resp.Found {
				fmt.Printf("Block: Hash=%s, Height=%d, Transactions=%d\n",
					resp.Block.Hash, resp.Block.Height, len(resp.Block.Transactions))
			} else {
				fmt.Println("Block not found")
			}

		case 5:
			fmt.Println("Subscribing to blocks...")
			stream, err := client.SubscribeBlocks(context.Background(), &SubscribeBlocksRequest{})
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}

			go func() {
				for {
					block, err := stream.Recv()
					if err != nil {
						log.Printf("Subscription error: %v", err)
						return
					}
					fmt.Printf("New block: %s (Height: %d)\n", block.Hash, block.Height)
				}
			}()
			fmt.Println("Subscription started in background")

		case 6:
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Invalid choice")
		}
	}
}

// Simple CLI client for testing
func cliClient() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewBlockchainServiceClient(conn)

	// Run simple commands
	ctx := context.Background()
	
	// Get latest block
	latest, err := client.GetLatestBlock(ctx, &GetLatestBlockRequest{})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Printf("Latest block: %s (Height: %d)\n", latest.Block.Hash, latest.Block.Height)
}