package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Simple blockchain implementation for demonstration
type Block struct {
	Hash         string
	PreviousHash string
	Timestamp    int64
	Transactions []Transaction
	Nonce        int32
	Height       int32
}

type Transaction struct {
	Hash      string
	From      string
	To        string
	Amount    float64
	Timestamp int64
	Signature string
}

// BlockchainServer implements the blockchain service
type BlockchainServer struct {
	UnimplementedBlockchainServiceServer
	blocks     []*Block
	mu         sync.RWMutex
	subscribers []chan *Block
}

func NewBlockchainServer() *BlockchainServer {
	// Initialize with genesis block
	genesis := &Block{
		Hash:         "genesis_hash",
		PreviousHash: "",
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		Nonce:        0,
		Height:       0,
	}
	
	return &BlockchainServer{
		blocks:      []*Block{genesis},
		subscribers: make([]chan *Block, 0),
	}
}

func (s *BlockchainServer) GetBlock(ctx context.Context, req *GetBlockRequest) (*GetBlockResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, block := range s.blocks {
		if block.Hash == req.Hash {
			return &GetBlockResponse{
				Block: &Block{
					Hash:         block.Hash,
					PreviousHash: block.PreviousHash,
					Timestamp:    block.Timestamp,
					Transactions: block.Transactions,
					Nonce:        block.Nonce,
					Height:       block.Height,
				},
				Found: true,
			}, nil
		}
	}

	return &GetBlockResponse{Found: false}, nil
}

func (s *BlockchainServer) GetLatestBlock(ctx context.Context, req *GetLatestBlockRequest) (*GetLatestBlockResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.blocks) == 0 {
		return &GetLatestBlockResponse{}, nil
	}

	latest := s.blocks[len(s.blocks)-1]
	return &GetLatestBlockResponse{
		Block: &Block{
			Hash:         latest.Hash,
			PreviousHash: latest.PreviousHash,
			Timestamp:    latest.Timestamp,
			Transactions: latest.Transactions,
			Nonce:        latest.Nonce,
			Height:       latest.Height,
		},
	}, nil
}

func (s *BlockchainServer) AddTransaction(ctx context.Context, req *AddTransactionRequest) (*AddTransactionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In a real implementation, this would add to mempool and wait for mining
	// For demo purposes, we'll create a new block immediately
	
	latest := s.blocks[len(s.blocks)-1]
	newBlock := &Block{
		Hash:         fmt.Sprintf("block_%d", len(s.blocks)),
		PreviousHash: latest.Hash,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{*req.Transaction},
		Nonce:        1,
		Height:       latest.Height + 1,
	}
	
	s.blocks = append(s.blocks, newBlock)
	
	// Notify subscribers
	for _, sub := range s.subscribers {
		select {
		case sub <- newBlock:
		default:
			// Skip if subscriber is not ready
		}
	}

	return &AddTransactionResponse{
		Success:         true,
		Message:         "Transaction added successfully",
		TransactionHash: req.Transaction.Hash,
	}, nil
}

func (s *BlockchainServer) GetTransaction(ctx context.Context, req *GetTransactionRequest) (*GetTransactionResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, block := range s.blocks {
		for _, tx := range block.Transactions {
			if tx.Hash == req.Hash {
				return &GetTransactionResponse{
					Transaction: &Transaction{
						Hash:      tx.Hash,
						From:      tx.From,
						To:        tx.To,
						Amount:    tx.Amount,
						Timestamp: tx.Timestamp,
						Signature: tx.Signature,
					},
					Found: true,
				}, nil
			}
		}
	}

	return &GetTransactionResponse{Found: false}, nil
}

func (s *BlockchainServer) SubscribeBlocks(req *SubscribeBlocksRequest, stream BlockchainService_SubscribeBlocksServer) error {
	// Create a channel for this subscriber
	blockChan := make(chan *Block, 100)
	
	s.mu.Lock()
	s.subscribers = append(s.subscribers, blockChan)
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		for i, sub := range s.subscribers {
			if sub == blockChan {
				s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
				close(blockChan)
				break
			}
		}
		s.mu.Unlock()
	}()

	// Send existing blocks first
	s.mu.RLock()
	for _, block := range s.blocks {
		if err := stream.Send(block); err != nil {
			s.mu.RUnlock()
			return err
		}
	}
	s.mu.RUnlock()

	// Listen for new blocks
	for {
		select {
		case block := <-blockChan:
			if err := stream.Send(block); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	blockchainServer := NewBlockchainServer()
	RegisterBlockchainServiceServer(s, blockchainServer)

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}