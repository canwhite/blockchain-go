package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"sync"
	"time"
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