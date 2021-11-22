package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

// global variable that holds all information about the state machine
//  see actor.go for implementation details
var server Actor

// this function implements the follower behavior to AppendEntriesRPC
func handleAppendRPC(w http.ResponseWriter, req *http.Request) {

	// TODO:
	// currently using print for displaying debug information, should
	//  change to log later
	fmt.Println("Trying to handle RPC")

	// extract rpc struct from json payload in request body
	//  assuming the leader sends a post request
	var rpc AppendEntriesRPC
	err := json.NewDecoder(req.Body).Decode(&rpc)

	fmt.Println("Append RPC received:")
	rpc.Print()
	rpc.PrintEntries()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)

	// advance state machine and get response
	resp := server.HandleAppendEntriesRPC(&rpc)

	fmt.Println("Follower log after rpc:")
	server.PrintLogs()
	fmt.Printf("%d, %t\n", resp.Term, resp.Success)

	// return response to leader as json
	json.NewEncoder(w).Encode(resp)
}

// this route is used by leader to get client request in the form of string
//  and append to its own logs. These will be appended to follower afterwards.
//   See AppendRpcTask for more detail
func handleClientReq(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Trying to handle client request")
	var creq ClientRequest

	err := json.NewDecoder(req.Body).Decode(&creq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	server.ReceiveClientRequest(creq.Req)
	w.WriteHeader(http.StatusOK)

	resp_json := make(map[string]string)
	resp_json["message"] = "Success"
	http_resp, err := json.Marshal(resp_json)
	if err != nil {
		fmt.Println("json marshal error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(http_resp)
}

// async task run by the scheduler for every fixed amount of seconds.
//  The interval can be adjusted in main function
// This task implements the process of leader appending logs to its
//  followers
func AppendRpcTask(server *Actor) {

	// lock to avoid dirty data
	server.mu.Lock()
	defer server.mu.Unlock()

	// this is used for sending post requests to followers
	http_client := &http.Client{Timeout: 10 * time.Second}

	// follower and candidate should not use this
	if server.Role != Leader {
		return
	}

	// for each peer (follower), try to append entries
	for i := 0; i < server.NumPeers; i++ {
		peer_id := server.Peers[i]

		if server.NextIndicies[peer_id] > server.LastLogIndicies[peer_id] && server.NextIndicies[peer_id]-1 < len(server.Logs) {
			fmt.Println("trying to append entry to follower...")
			var last_term int
			if server.NextIndicies[peer_id]-1 <= 0 {
				last_term = 0
			} else {
				last_term = server.Logs[server.NextIndicies[peer_id]-2].Term
			}

			// build payload and send post request
			append_rpc := AppendEntriesRPC{
				Term:         server.Logs[server.NextIndicies[peer_id]-1].Term,
				LeaderId:     1,
				PrevLogIndex: server.NextIndicies[peer_id] - 1,
				PrevLogTerm:  last_term,
				Entries:      []Log{server.Logs[server.NextIndicies[peer_id]-1]},
				CommitIndex:  server.CommitIdx}

			// TODO: add error handling
			rpc_json, _ := json.Marshal(append_rpc)
			r, _ := http_client.Post("http://localhost:"+peer_id+"/append-entry-rpc", "application/json", bytes.NewBuffer(rpc_json))

			var follower_resp AppendResp
			err := json.NewDecoder(r.Body).Decode(&follower_resp)

			if err != nil {
				fmt.Println("json parse error for response body")
				return
			}

			if !follower_resp.Success {
				fmt.Println("append failed, decrease next index by one")

				// if append fails, we roll back next index for that peer by one so that
				//  next time the leader will try to append the previous log
				server.NextIndicies[peer_id] -= 1
			} else {
				fmt.Println("append success")

				// if append succeeded, update next index and last log index accordingly
				server.LastLogIndicies[peer_id] = server.NextIndicies[peer_id]
				server.NextIndicies[peer_id] += 1

				// TODO: handle commit index here
			}

			r.Body.Close()
		}
	}
}

func main() {

	id := os.Args[1]
	var rtype RoleType
	if strings.Compare(os.Args[2], "leader") == 0 {
		rtype = Leader
	} else {
		rtype = Follower
	}

	num_peers, _ := strconv.Atoi(os.Args[3])
	peers := os.Args[4:]

	// id := "8090"
	// rtype := Follower
	// num_peers := 1
	// peers := make([]string, 1)
	// peers[0] = "8091"

	server.Init(id, rtype, num_peers, peers)
	http.HandleFunc("/append-entry-rpc", handleAppendRPC)
	http.HandleFunc("/client-api", handleClientReq)
	// port := os.Args[1]

	// use scheduler for sending append rpcs
	scheduler := gocron.NewScheduler(time.UTC)

	// change the number here to modify the interval
	scheduler.Every(3).Seconds().Do(AppendRpcTask, &server)
	scheduler.StartAsync()

	http.ListenAndServe(":"+id, nil)
}
