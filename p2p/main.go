package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	net "github.com/libp2p/go-libp2p-net"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	host "github.com/libp2p/go-libp2p/core/host"
	ma "github.com/multiformats/go-multiaddr"
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

var mutex = &sync.Mutex{}

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

// SHA256 hashing
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	//这是第一章的精髓
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock
}

/** 
除了“拨号”和“转发”，
---
关于拨号：
go-libp2p 通过节点发现机制（如 Kademlia DHT 或 mDNS）找到其他节点，类似于“拨号”联系对方。
每个节点有唯一 ID（PeerID），通过多地址（multiaddr，如 /ip4/127.0.0.1/tcp/9000）建立连接。

---
关于转发：
它支持多种传输协议（TCP、QUIC、WebRTC）和中继（Relay），能在 NAT 或防火墙后转发数据，
像一个智能交换机，将消息从一个节点路由到另一个节点。

---
go-libp2p 还提供分布式哈希表（DHT）查找资源、发布/订阅（PubSub）广播消息，
以及 NAT 穿透，功能远超简单的转发中心。

---
NAT穿透是什么？
NAT是一种将私有 IP 地址（如 192.168.x.x）映射到公网 IP 地址的技术，
问题是在 P2P 网络中，两个位于不同 NAT 后的节点（如家庭路由器后的设备）无法直接建立连接，
NAT 穿透是指通过技术手段（如协议或中继）让两个或多个 NAT 后的节点直接通信，绕过 NAT 或防火墙的限制。

*/


// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
// secio 是 go-libp2p 中用于安全通信的一种加密协议，全称是 Secure Communication

func makeBasicHost(listenPort int, secio bool, randseed int64)(host.Host, error){
	
	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	//实际上是为了区分生产和调试环境，randseed == 0, 使用真呢真难过的随机性
	if randseed == 0 {
		r = rand.Reader
	} else {
		//伪随机数生成器的确定性：当使用相同的种子初始化math/rand的随机数生成器时，
		//它总是会产生完全相同的随机数序列。这在测试中非常有用，因为你可以重现完全相同的行为。
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	// GenerateKeyPairWithReader 参数说明:
	// 1. crypto.RSA - 密钥类型，这里使用 RSA 算法
	// 2. 2048 - 密钥长度，2048位是当前推荐的安全强度
	// 3. r - 随机源，前面根据 randseed 决定使用真随机还是伪随机
	// 返回结果:
	// priv - 生成的私钥对象，用于节点身份验证和加密通信
	// pub - 生成的公钥(这里用 _ 忽略)，用于派生节点ID(PeerID)
	// err - 错误对象，如果密钥生成失败会返回非nil值
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil,err 
	}

	// 这段代码创建了一个基本的LibP2P主机配置选项:
	// 1. 监听本地127.0.0.1地址和指定端口
	// 2. 使用前面生成的私钥作为节点身份标识
	// 这些选项将传递给libp2p.New()函数来创建实际的P2P主机
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	if !secio {
		opts = append(opts, libp2p.NoSecurity)
	}

	basicHost, err := libp2p.New(opts...)
	
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().String()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	//en进入 capsule胶囊，合到一起就是概括
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -l %d -d %s -secio\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}

	return basicHost, nil
}

//处理输入数据流
func handleStream(s net.Stream){
	log.Println("Got a new stream!")
	//creating a buffer stream for non blocking read and write
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(rw)
	go writeData(rw)
	// stream 's' will stay open until you close it (or the other side closes it).
}


func readData(rw  *bufio.ReadWriter){
	for{
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if str == ""{
			return
		}
		if str != "\n" {
			// 这里的0表示创建一个空的Block切片，用于后续存储解析后的区块链数据
			// 相当于初始化一个空的区块链容器
			chain := make([]Block, 0)
			//然后往下进行
			// &chain 表示将 chain 变量的内存地址传递给 json.Unmarshal 函数
			// 在 JSON 解码时，解码结果会被直接存入 chain 变量中
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			//这种算是读，也就是reading
			mutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes,err := json.MarshalIndent(Blockchain,"","	")
				if err != nil{
					log.Fatal(err)
				}
				//这行Go代码 fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes)) 的作用是在终端中显示绿色文本
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}
func writeData(rw * bufio.ReadWriter){
	//todo
}


