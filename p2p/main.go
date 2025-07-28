package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	golog "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
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
func handleStream(s network.Stream){
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
func writeData(rw *bufio.ReadWriter) {
	// Goroutine for periodic blockchain data sending
	go func() {
		for {
			time.Sleep(5 * time.Second)
			// 读取
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()
			// 输出
			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			// 这行代码的作用是将缓冲区中的数据强制写入底层io.Writer
			// 确保所有缓冲的数据都被发送到网络连接中
			// 如果不调用Flush(), 数据可能只会留在缓冲区而不会实际发送
			rw.Flush()
			mutex.Unlock()
		}
	}()

	// Separate goroutine for user input handling
	go func() {
		stdReader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("> ")
			sendData, err := stdReader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			sendData = strings.Replace(sendData, "\n", "", -1)
			bpm, err := strconv.Atoi(sendData)
			if err != nil {
				log.Fatal(err)
			}
			newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)

			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
				mutex.Lock()
				Blockchain = append(Blockchain, newBlock)
				mutex.Unlock()
			}

			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}

			spew.Dump(Blockchain)

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
		}
	}()
}


func main(){
	//创世区块儿
	t := time.Now()
	genesisBlock := Block{}
	/* 
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
	*/
	genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), ""}
	Blockchain = append(Blockchain, genesisBlock)

	golog.SetAllLoggers(golog.LevelInfo) // Change to DEBUG for extra info

	//parse options，这里返回的都是地址
	//-l 指定监听端口
	listenF := flag.Int("l", 0, "wait for incoming connections")
	//-d 指定要连接的目标节点地址
	target := flag.String("d", "", "target peer to dial")
	//-secio 启用secio加密通道
	secio := flag.Bool("secio", false, "enable secio")
	//-seed：设置随机种子用于节点ID生成
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()


	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	//make a host that listens on the given multi address
	ha, err := makeBasicHost(*listenF,*secio,*seed)
	if err != nil{
		log.Fatal(err)
	}

	//监听模式
	if *target == "" {
		log.Println("listening for connections")
		// Set a stream handler on host A. /p2p/1.0.0 is
		// a user-defined protocol name.
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)
		select {} // hang forever。程序不退出，整个程序保持活跃状态

	}else{ //拨打模式，主动向peer拨打
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)
		// The following code extracts target's peer ID from the
		// given multiaddress
		// /ip4/127.0.0.1/tcp/10000/p2p/QmPeerID
		// /ip4/127.0.0.1 - IPv4地址
		// /tcp/10000 - TCP端口
		// /p2p/QmPeerID - 节点的PeerID（可选）
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil{
			log.Fatalln(err)	
		}

		// 解析目标地址中的PeerID
		// ipipfsaddr.ValueForProtocol(ma.P_IPFS) 会从多地址中提取出 /p2p/ 后面的部分
		// 例如对于地址 /ip4/127.0.0.1/tcp/10000/p2p/QmPeerID
		// 这里提取出来的pid就是 "QmPeerID" 这个字符串

		pid, err := ipfsaddr.ValueForProtocol(ma.P_P2P)
		if err != nil {
			log.Fatalln(err)
		}
		// 这里需要解码是因为从多地址中提取的pid是Base58编码的字符串形式
		// peer.IDB58Decode的作用是将Base58编码的PeerID字符串转换为二进制格式的PeerID对象
		// 这是必要的因为:
		// 1. 网络传输和加密操作需要使用二进制格式的PeerID
		// 2. Base58编码只是用于人类可读和URL安全的表示
		// 3. 解码后可以验证PeerID的格式是否正确
		peerid, err := peer.Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}
		
		// Decapsulate操作会从多地址中移除指定的协议部分
		// 例如原地址是 /ip4/127.0.0.1/tcp/10000/p2p/QmPeerID
		// 这里会先创建 /p2p/QmPeerID 部分的多地址
		// 然后从原地址中移除这部分，得到 /ip4/127.0.0.1/tcp/10000
		// 这样我们就分离出了基础网络地址和peer ID两部分
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peerid.String()))
		//so target address 就是/ip4/127.0.0.1/tcp/10000
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		ha.Peerstore().AddAddr(peerid,targetAddr, peerstore.PermanentAddrTTL)
		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		// Create a thread to read and write data.
		go writeData(rw)
		go readData(rw)
		
		select{}
	}

}