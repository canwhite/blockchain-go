package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

//block struct
type Block struct{
	Index int
	Timestamp string
	// BPM : (Beats Per Minute)
	BPM int 
	Hash string //1）to save space， 2） Preserve integrity of the blockchain
	PrevHash string
}

//generate hash
func calculateHash(block Block) string{
	record := fmt.Sprint(block.Index) + block.Timestamp + fmt.Sprint(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	// Sum方法计算并返回SHA-256哈希值，nil参数表示将结果存储在新的字节切片中
	// Sum方法会将当前哈希状态与输入数据（这里为nil）进行最终计算，返回哈希结果
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//generate block 
func generateBlock(oldBlock Block, BPM int) (Block,error){
	var newBlock Block
    t:= time.Now()
	newBlock.Index = oldBlock.Index + 1 //index +1
	newBlock.Timestamp = t.String() 
	newBlock.BPM = BPM 
	newBlock.PrevHash = oldBlock.Hash//current preHash = old block hash
	newBlock.Hash = calculateHash(newBlock) //cal new hash
	return newBlock, nil 
}





func main() {
    fmt.Println("Hello, Go Project!")
}