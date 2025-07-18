package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block

// SHA256 hashing
func calculateHash(block Block) string {
	record := fmt.Sprint(block.Index) + block.Timestamp + fmt.Sprint(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// make sure the chain we're checking is longer than the current blockchain
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

// bcServer handles incoming concurrent Blocks
var bcServer chan []Block

func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bcServer = make(chan []Block)

	//创建创世模块
	t := time.Now()
	genesisBlock := Block{0, t.String(), 0, "", ""} 
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	//start a tcp server
	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Println("TCP Server listening on :", os.Getenv("ADDR"))
	
	// Handle incoming connections
	/**
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
	*/
}


/*
func handleConnection(conn net.Conn) {
	defer conn.Close()
	
	// Read incoming data
	decoder := json.NewDecoder(conn)
	var msg []Block
	if err := decoder.Decode(&msg); err != nil {
		log.Println("Decode error:", err)
		return
	}

	// Validate and potentially replace blockchain
	if len(msg) > 0 && isBlockValid(msg[0], Blockchain[len(Blockchain)-1]) {
		bcServer <- msg
	}
	
	// Send our blockchain to peer
	if err := json.NewEncoder(conn).Encode(Blockchain); err != nil {
		log.Println("Encode error:", err)
	}
}
*/