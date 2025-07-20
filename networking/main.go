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
	"os"
	"strconv"
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
	//fmt.Sprint是拼接字符串的方法
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
	
	// Handle incoming connections，this is a infinite loop
	for {
		//accept
		conn, err := server.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConn(conn)
	}

}

func handleConn(conn net.Conn){
	//defer的好处是，可以在代码的开头处理两头边界，然后后续专注逻辑，这是一个很好的开始
	defer conn.Close()
	// io 包提供了与终端、文件等输入输出设备交互的功能
	// 在这里，conn 是一个网络连接，我们可以通过 io.WriteString 向连接写入数据
	// 这类似于向终端输出，只是输出目标从终端变成了网络连接
	io.WriteString(conn,"Enter a new BPM:");

	scanner := bufio.NewScanner(conn)

	//take in BPM from stdin and add it to blockchain after conducting necessary validation
	// go must after a func call
	// go是携程的意思，后接一个自执行函数
	go func ()  {
		for scanner.Scan(){
			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil{
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}
			newBlock,err := generateBlock(Blockchain[len(Blockchain)-1],bpm)
			if err != nil {
				log.Println(err)
				continue
			}
			//进行校验
			if isBlockValid(newBlock,Blockchain[len(Blockchain)-1]){
				newBlockChain := append(Blockchain,newBlock)
				replaceChain(newBlockChain)
			}
			bcServer <- Blockchain
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()

	// simulate receiving broadcast	
	go func ()  {
		//类似while(true)
		for{	
			time.Sleep(30 * time.Second)
			//注意这里是转json，实际上是一个byte[]
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			// string() 是 Go 语言中的类型转换，将其他类型的数据转换为字符串、
			io.WriteString(conn, string(output))
		} 
	}()
	
	// 这里是一个无限循环，用于持续监听bcServer通道的消息
	// bcServer是一个通道，当有新的区块链数据时会被发送到这个通道
	// 每次接收到新的区块链数据时，使用spew.Dump打印出当前的区块链状态
	// 这种设计模式常用于实时更新和监控区块链的状态变化
	for _ = range bcServer {
		spew.Dump(Blockchain)
	}

	
}


