package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"strconv"
	"github.com/go-co-op/gocron"
)

// global variable that holds all information about the state machine
//  see actor.go for implementation details
var server Actor

const HB_EXPIRED_TIME = 4
const VTERM_EXPIRED_TIME = 4

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

func handleVoteReq(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Trying to handle Vote Request from Follower side")

	var vreq VoteReqRPC

	err := json.NewDecoder(req.Body).Decode(&vreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("Follower received a Vote Request from server: %s\n", vreq.candidateID)

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusCreated)

	resp := server.HandleVoteReq(&vreq)
	fmt.Printf("Vote result: %t\n", resp.voteGranted)
	// return response to leader as json
	json.NewEncoder(w).Encode(resp)

	return
}

func FollowerTask(server *Actor) {
	server.mu.Lock()
	defer server.mu.Unlock()
	// Transition from follower to candidate (if not heartbeat received and no vote term started at current term)

	HBtimeout := time.Since(server.lastHBtime).Seconds()
	VRtimeout := time.Since(server.lastVRtime).Seconds()
	if HBtimeout > HB_EXPIRED_TIME {
		if server.VotedTerm == 0 || VRtimeout > VTERM_EXPIRED_TIME {
			server.Role = Candidate
			server.VotedTerm++
		}
	}
	return
}

func CandidateTask(server *Actor) {
	// lock to avoid dirty data
	server.mu.Lock()
	defer server.mu.Unlock()
	// only candidate confirm if it is selected
	if server.Role != Candidate {
		return
	}
	http_client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < server.NumPeers; i++ {
		peer_id := server.Peers[i]
		req := VoteReqRPC{
			candidateID:  server.ID,
			term:         server.CurrentTerm,
			voteterm:     server.VotedTerm + 1,
			lastLogIndex: server.Logs[len(server.Logs)-1].Index,
			lastLogTerm:  server.Logs[len(server.Logs)-1].Term}

		req_json, er := json.Marshal(req)
		if er != nil {
			fmt.Println("marshal failed")
		}
		r, _ := http_client.Post("http://localhost:"+peer_id+"/request-vote", "application/json", bytes.NewBuffer(req_json))

		var vote_rsp VoteRsp
		err := json.NewDecoder(r.Body).Decode(&vote_rsp)

		if err != nil {
			fmt.Println("json parse error for response body")
			return
		}
		if vote_rsp.voteGranted == true {
			server.counter++
			if server.counter > server.NumPeers/2 {
				break
			}
		} else {
			if vote_rsp.term >= server.VotedTerm { // my voteterm is not up-to-date
				break
			}
		}
		r.Body.Close()
	}

	if server.counter > server.NumPeers/2 {
		for i := 0; i < server.NumPeers; i++ {
			peer_id := server.Peers[i]
			var heartbeat AppendEntriesRPC
			heartbeat_json, er := json.Marshal(heartbeat)
			if er != nil {
				fmt.Println("marshal failed")
			}
			r, _ := http_client.Post("http://localhost:"+peer_id+"/", "application/json", bytes.NewBuffer(heartbeat_json))
			server.Role = Leader
			server.CurrentTerm++
			server.VotedTerm = 0
			r.Body.Close()
		}
		server.counter = 1
	} else {
		server.Role = Follower
	}
}

// async task run by the scheduler for every fixed amount of seconds.
//  The interval can be adjusted in main function
// This task implements the process of leader appending logs to its
//  followers
func LeaderTask(server *Actor) {

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
				LeaderId:     server.ID,
				PrevLogIndex: server.NextIndicies[peer_id] - 1,
				PrevLogTerm:  last_term,
				Entries:      []Log{server.Logs[server.NextIndicies[peer_id]-1]},
				CommitIndex:  server.CommitIdx}

			// TODO: add error handling
			rpc_json, er := json.Marshal(append_rpc)
			if er != nil {
				fmt.Println("marshal failed")
			}
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
			server.PrintLeaderState()
		}
	}

}

func main() {
	// initialization
	// read from input: id string, num_peers int, peers []string
	id := os.Args[1]
	num_peers_string := os.Args[2]
	num_peers, err := strconv.Atoi(num_peers_string)
    if err != nil {
        // handle error
        fmt.Println(err)
        os.Exit(2)
    }
	peers := os.Args[3:]
	if len(peers) != num_peers {
		fmt.Println("Please correct inputs: id; num_peers; peer ports")
		os.Exit(1)
	}
	server.Init(id, num_peers, peers)
	
	// initialize from files
	/*
	config_filename := os.Args[1]
	// config_filename := "./actor_config/follower1.json"
	server.InitFromConfigFile(config_filename)
	*/
	
	if server.Role == Leader {
		server.PrintLeaderState()
	}
	server.counter = 1
	// server.Init(id, rtype, num_peers, peers)
	// port := os.Args[1]

	//client
	//http.HandleFunc("/client-api", handleClientReq)
	//http.HandleFunc("/client-api/disable_server", handleClientReq)
	//leader api
	//http.HandleFunc("/leader/commit", handleClientReq)
	http.HandleFunc("/client-api", handleClientReq)
	//follower api
	http.HandleFunc("/append-entry-rpc", handleAppendRPC)
	http.HandleFunc("/request-vote", handleVoteReq)

	// use scheduler for sending append rpcs
	scheduler := gocron.NewScheduler(time.UTC)

	// change the number here to modify the interval
	scheduler.Every(3).Seconds().Do(LeaderTask, &server)
	scheduler.Every(3).Seconds().Do(FollowerTask, &server)
	scheduler.Every(3).Seconds().Do(CandidateTask, &server)
	scheduler.StartAsync()

	http.ListenAndServe(":"+server.ID, nil)
}
