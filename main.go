package main

import (
	"fmt"
)

type Block struct{
	Index int
	TimeStamp string
	// BPM : (Beats Per Minute)
	BPM int 
	Hash string //1）to save space， 2） Preserve integrity of the blockchain
	PrevHash string
}

//generate hash
// func calculateHash(block Block) string{}





func main() {
    fmt.Println("Hello, Go Project!")
}