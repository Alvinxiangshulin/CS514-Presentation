package main

import (
	"os"
	"encoding/json"
	"fmt"
	"net/http"
)

var server Follower

func handleAppendRPC(w http.ResponseWriter, req *http.Request) {
	var rpc AppendEntriesRPC
	err := json.NewDecoder(req.Body).Decode(&rpc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)
	resp := server.HandleAppendEntriesRPC(&rpc)
	fmt.Printf("%d, %t\n", resp.Term, resp.Success)
	json.NewEncoder(w).Encode(resp)
}

func hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "hello\n")
}

// func headers(w http.ResponseWriter, req *http.Request) {

// }
func main() {
	server.Init()
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/append-entry-rpc", handleAppendRPC)
	port := os.Args[1]
	http.ListenAndServe(":" + port, nil)
}
