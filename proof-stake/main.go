package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)


type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
	Validator string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block
//tempBlocks is simply a holding tank of blocks before one of them 
//is picked as the winner to be added to Blockchain
var tempBlocks []Block


//candidateBlocks  handles incoming blocks for validation
var candidateBlocks = make(chan Block)


//announcements broadcasts winning validator to all nodes 
var announcements = make(chan string)

var mutex = &sync.Mutex{}

//validators keeps track of open validators and balances 
var validators = make(map[string]int)


// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp +strconv.Itoa(block.BPM) + block.PrevHash
	return calculateHash(record)
}

// generateBlock creates a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int, address string) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Validator = address //caveat 

	return newBlock, nil
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

/* 
Allow it to enter a token balance (remember for this tutorial, we won’t perform any balance checks since there is no wallet logic)
Receive a broadcast of the latest blockchain
Receive a broadcast of which validator in the network won the latest block
Add itself to the overall list of validators
Enter block data BPM — remember, this is each validator’s pulse rate
Propose a new block
*/

func handleConn(conn net.Conn){

	defer conn.Close()

	//announcement
	go func ()  {
		for {
			msg := <-announcements
			//io写到哪，主要看你第一个参数
			io.WriteString(conn, msg)
		}		
	}()


	//register
	var address string
	io.WriteString(conn,"Enter token balance:");
	scanBalance := bufio.NewScanner(conn)
	
	for scanBalance.Scan(){
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanBalance.Text(), err)
			return
		}
		
		t := time.Now()
		address = calculateHash(t.String())
		validators[address] = balance
		fmt.Println(validators)
		break
	}

	//bpm
	io.WriteString(conn,"\nEnter a new BPM:")


	scanBPM := bufio.NewScanner(conn) 	//buff就是内存临时存储
	//the main task is creating block 
	go func(){
		for {
			for scanBPM.Scan(){
				bpm, err := strconv.Atoi(scanBPM.Text())
				if err != nil {
					log.Printf("%v not a number: %v", scanBPM.Text(), err)
					// 是的，delete(validators, address)会从validators映射中删除该address对应的验证者
					// 这里可以添加一些日志输出以便调试
					log.Printf("删除无效验证者: %s", address)
					delete(validators, address)
					conn.Close()
				}

				// 在处理区块链数据时需要遵循以下加锁原则：
				// 1. 读取共享数据时：如果只是简单读取且后续没有修改操作，可以不加锁
				// 2. 修改共享数据时：必须加锁保护，防止并发修改

				mutex.Lock()
				oldLastIndex := Blockchain[len(Blockchain)-1]
				mutex.Unlock()

				// 在generateBlock时不加锁是因为:
				// 1. generateBlock只是基于旧区块创建新区块，不直接修改区块链状态
				newBlock, err := generateBlock(oldLastIndex, bpm, address)

				if err != nil{
					log.Println(err)
					continue
				}

				if isBlockValid(newBlock, oldLastIndex) {
					candidateBlocks <- newBlock
				}

				io.WriteString(conn, "\nEnter a new BPM:")
			}
		}
	}()

	// simulate receiving broadcast ,watch data
	for {
		mutex.Lock()
		output, err := json.Marshal(Blockchain) //读取和修改
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output)+"\n")
	}


}