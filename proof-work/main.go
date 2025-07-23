package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)
const difficulty = 1

type Block struct {
        Index      int
        Timestamp  string
        BPM        int
        Hash       string
        PrevHash   string
        Difficulty int
        Nonce      string
}

var Blockchain []Block

type Message struct {
        BPM int
}

//get the mutex instance
var mutex = &sync.Mutex{}

func run() error{
	//start with a server
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil 
}

func makeMuxRouter() http.Handler{
	// 在Go语言中，方法或变量名以大写字母开头表示该方法是导出的（public），可以被其他包访问
	// 小写字母开头的方法或变量是私有的（private），只能在当前包内使用
	// 这是Go语言的访问控制机制，通过首字母大小写来区分可见性
	// 例如：
	// PublicMethod() - 可以被其他包调用
	// privateMethod() - 只能在当前包内使用
	// 这种设计简化了访问控制，不需要像其他语言那样使用public/private等关键字
	muxRouter :=  mux.NewRouter()
	muxRouter.HandleFunc("/",handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/",handleWriteBlock).Methods("POST")
	return muxRouter;

}


func handleGetBlockchain(w http.ResponseWriter, r *http.Request){
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	//todo
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	//having  another method  is to use w to set header and response
	//改进
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(bytes)
	io.WriteString(w,string(bytes))

}

//need a function to support write handler
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}){
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func handleWriteBlock(w http.ResponseWriter, r * http.Request){
	w.Header().Set("Content-Type","application/json")
	var m Message //Message是传过来的
	decoder := json.NewDecoder(r.Body)
	//报错提前处理
	if err:= decoder.Decode(&m); err != nil{
		//如果报错直接将payload发回去
		respondWithJSON(w,r,http.StatusBadRequest,r.Body)
		return
	}
	defer r.Body.Close()

	//ensure atomicity when creating new block 
	mutex.Lock()

	newBlock := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	mutex.Unlock()

	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		Blockchain = append(Blockchain, newBlock)
		spew.Dump(Blockchain)
	}   
	respondWithJSON(w, r, http.StatusCreated, newBlock)

}	

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

func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}


//distinguish hash valid with block valid
func isHashValid(hash string, difficulty int) bool{
	//difficulty is repeating count
	prefix := strings.Repeat("0",difficulty)
	return strings.HasPrefix(hash, prefix)
}


//perhaps need to change , generate block 
func generateBlock(oldBlock Block, BPM int) Block{
	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index +1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty

	//需要一直算下去
	for i := 0 ; ; i++{
		//将i转化字符串
		hex := fmt.Sprintf("%x", i)
		//通过改变Nonce，再对block算哈希，看是否满足情况
        newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock),newBlock.Difficulty){
			fmt.Println(calculateHash(newBlock), " do more work!")
			time.Sleep(time.Second)
			continue
		}else {
			fmt.Println(calculateHash(newBlock), " work done!")
            newBlock.Hash = calculateHash(newBlock)
            break
		}
	}
	return newBlock
}

func main(){
	//这里Load，在要使用的时候Getenv
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}   
	
	go func(){
		t := time.Now()
		genesisBlock := Block{}
		genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), "", difficulty, ""} 
		spew.Dump(genesisBlock)

		mutex.Lock()
		Blockchain = append(Blockchain, genesisBlock)
		mutex.Unlock()
	}()

	log.Fatal(run())
}
/** 
比特币的工作证明验证机制：

不只是简单检查前导0个数，而是使用目标阈值(target)比较
区块哈希必须小于当前网络目标值(target)才被视为有效
目标值是一个256位数字，通过难度值计算得出
前导0个数只是目标值的一种可视化表现方式
比特币的难度调整算法更精确(BTC每2016个区块调整一次)
实际实现通过比较哈希数值是否小于目标数值，而非仅看前导0
*/


