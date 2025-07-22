package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
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
	



}

