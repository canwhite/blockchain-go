package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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

// Blockchain is a series of validated Blocks
var Blockchain []Block

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

// about fork attacking
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain){
		Blockchain = newBlocks
	}
}


func run() error{
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	//所有方法传入的都是引用
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//if的承接写法
	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter,r *http.Request){
	//get firstly
	bytes,err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// io.WriteString(w, string(bytes))
	//改进
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

type Message struct{
	BPM int
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request){
	var m Message;
	//get
	// queryParams := r.URL.Query()
	// bpm := queryParams.Get("BPM")
	// m.BPM, _ = strconv.Atoi(bpm)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	// 使用defer来确保在函数返回前关闭请求体，防止资源泄漏
	defer r.Body.Close()

	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}

	//continue
	if isBlockValid(newBlock,Blockchain[len(Blockchain)-1]){
		newBlockChain := append(Blockchain,newBlock)
		replaceChain(newBlockChain)
		spew.Dump(newBlock)
	}
	respondWithJSON(w,r,http.StatusCreated,newBlock)
}

//writing need a function , named as respond with json
func respondWithJSON(w http.ResponseWriter,r *http.Request, code int, payload interface{}){
	//转化为json
	response, err  := json.MarshalIndent(payload, "", "  ")
	
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

//POST:
//curl -X POST http://127.0.0.1:8080 -H "Content-Type: application/json" -d '{"BPM":60}'
func main() {
	//先获取.env中的内容
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{0, t.String(), 0, "", ""}
		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()	
	// 使用log.Fatal(run())运行的好处：
	// 1. 如果run()函数返回错误，log.Fatal会自动记录错误并终止程序
	// 2. 相比直接调用run()，这种方式可以确保程序在遇到错误时不会继续执行
	// 3. log.Fatal会自动将错误信息输出到标准错误流，方便调试
	// 4. 这种方式更符合Go语言的错误处理习惯，使代码更健壮
	log.Fatal(run())
}