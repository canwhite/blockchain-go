package main

/**
Proof of Stake 执行流程
├── 初始化阶段
│   ├── 加载环境变量
│   ├── 创建创世区块
│   ├── 启动TCP服务器 (端口9000)
│   ├── 启动后台服务
│   │   ├── 候选区块处理器 (从candidateBlocks通道读取)
│   │   └── 获胜者选择器 (每30秒执行一次)
│   └── 等待客户端连接
│
├── 客户端连接处理 (handleConn)
│   ├── 启动公告广播器
│   ├── 验证者注册
│   │   ├── 要求输入代币余额
│   │   └── 生成验证者地址并存储到validators映射
│   ├── BPM数据处理
│   │   ├── 要求输入BPM
│   │   ├── 生成新区块
│   │   ├── 验证区块有效性
│   │   ├── 发送到候选区块通道
│   │   └── 继续等待下一个BPM输入
│   └── 区块链状态广播
│       └── 每30秒向客户端发送当前区块链状态
│
├── 权益证明核心逻辑 (pickWinner)
│   ├── 获取临时区块池
│   ├── 构建彩票池
│   │   ├── 遍历临时区块
│   │   ├── 检查验证者是否已在池中
│   │   └── 根据代币余额加权添加到彩票池
│   ├── 随机选择获胜者
│   │   ├── 使用时间戳作为随机种子
│   │   └── 从彩票池中随机选择
│   ├── 添加获胜区块到主链
│   └── 广播获胜信息
│
└── 核心算法
    ├── 区块哈希计算 (calculateBlockHash)
    ├── 区块生成 (generateBlock)
    └── 区块验证 (isBlockValid)
        ├── 索引连续性检查
        ├── 前哈希匹配检查
        └── 当前哈希正确性检查

// 构建彩票池的核心逻辑
for _, block := range temp {
    // 避免重复添加同一验证者
    for _, node := range lotteryPool {
        if block.Validator == node {
            continue OUTER
        }
    }

    // 根据代币余额加权
    k, ok := setValidators[block.Validator]
    if ok {
        // 代币越多，被添加到彩票池的次数越多，中奖概率越高
        for i := 0; i < k; i++ {
            lotteryPool = append(lotteryPool, block.Validator)
        }
    }
}
客户端输入BPM → 生成区块 → 候选区块通道 → 临时区块池 → 权益证明选择 → 主区块链 → 状态广播
数据输入阶段：handleConn处理客户端输入，输入广播
验证阶段：isBlockValid验证区块有效性
缓冲阶段：通过candidateBlocks通道和tempBlocks缓冲
选择阶段：pickWinner进行权益证明选择
提交阶段：将获胜区块添加到主链
通知阶段：广播最新状态
*/

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)


type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
	Validator string
}

//candidateBlocks  handles incoming blocks for validation
var candidateBlocks = make(chan Block)

//tempBlocks is simply a holding tank of blocks before one of them 
//is picked as the winner to be added to Blockchain
var tempBlocks []Block


// Blockchain is a series of validated Blocks
var Blockchain []Block

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
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash
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
		
		fmt.Println(balance)
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
		time.Sleep(30 * time.Second) // 添加适当的延迟，避免无限快速广播
		mutex.Lock()
		output, err := json.Marshal(Blockchain) //读取和修改
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output)+"\n")
	}
}


// pickWinner creates a lottery pool of validators and chooses the validator who gets to forge a block to the blockchain
// by random selecting from the pool, weighted by amount of tokens staked
// pick winner 实际上是从TempBlocks往BlockChain里加东西
func pickWinner(){
	time.Sleep(30 * time.Second)

	mutex.Lock()
	temp := tempBlocks //because of reading
	mutex.Unlock()

	lotteryPool := []string{}


	if len(temp) > 0 {


	//continue and break jump label
	OUTER:
		// slightly modified traditional proof of stake algorithm
		// from all validators who submitted a block, weight them by the number of staked tokens
		// in traditional proof of stake, validators can participate without sub
		for _, block := range temp {
			// if already in lottery pool, skip
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			// lock list of validators to prevent data race
			mutex.Lock()
			setValidators := validators
			mutex.Unlock()
			//"comma ok"模式。所以即使map中存储的是int值，你仍然可以获取到两个返回值：值本身和一个表示键是否存在的布尔值。
			k, ok := setValidators[block.Validator]
			if ok {
				//如果该验证者存在于映射中（ok为true），则根据权重值将该验证者多次添加到lotteryPool中
				//根据权重往池子里加是个好思路
				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}
		}

		// randomly pick winner from lottery pool
		//使用当前时间的Unix时间戳作为种子创建一个新的随机数源，可以想一下为什么用时间戳做种子
		s := rand.NewSource(time.Now().Unix())
		//基于上边的seed，创建随机数生成实例
		r := rand.New(s)
		//r.Intn(len(lotteryPool)) 生成一个0到lotteryPool长度减1之间的随机整数
		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]
		// add block of winner to blockchain and let all the other nodes know

		for _, block := range temp {
			if block.Validator == lotteryWinner {
				mutex.Lock()
				Blockchain = append(Blockchain, block)
				mutex.Unlock()
				for range validators {
					announcements <- "\nwinning validator: " + lotteryWinner + "\n"
				}
				break
			}
		}
	}

	// 清理临时区块池
	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}

func main(){
	
	err := godotenv.Load() //if we need it,
	if err != nil {
		log.Fatal(err)
	}
	
	// create genesis block 
	t := time.Now()

	// this is convenient for reducer declare
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, calculateBlockHash(genesisBlock), "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain,genesisBlock)

	//start TCP and serve TCP server
	//启动时tcp:port.如果godotenv已经load，using it just need to os.getEnv
	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP Server Listening on port :", os.Getenv("ADDR"))
	defer server.Close()


	go func(){
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock() 
		}
	}()

	go func() {
		for {
			pickWinner()
		}
	}()

	//handleConn是建立连接之后往candidateBlock中加
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}

}