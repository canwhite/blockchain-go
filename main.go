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
	//nil as input， this method concatenates record 
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

//verify block
func isBlockValid(newBlock Block, oldBlock Block) bool {
	//index
	//no parenthesis
	if oldBlock.Index +1 != newBlock.Index {
		return false
	}
	//hash
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}
	
	//double check
	if calculateHash(newBlock) != newBlock.Hash{
		return false
	}

	return true
}



func main() {
    fmt.Println("Hello, Go Project!")
}