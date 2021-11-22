package main

import (
	"errors"
	"fmt"
	"sync"
)

type RoleType int

const (
	Leader RoleType = iota
	Follower
	Candidate
)

type Actor struct {
	// tracks the index of the latest appended log from current leader to each follower
	LastLogIndicies map[string]int

	// tracks the index of the next should-be-appended log for each follower
	NextIndicies map[string]int

	// number of all other peers in system (excluding itself)
	NumPeers int

	// counts the number of times a log has been successfully appended to a follower
	//  used for deciding when to commit at where
	AppendCounter []int

	// track commit status
	CommitIdx int

	// port numbers in string form for each peer in system
	Peers []string

	mu          sync.RWMutex
	CurrentTerm int
	Logs        []Log

	// timeout for receving heartbeat messages and election
	Timeout int
	Role    RoleType

	// port number of the actor, assuming this is unique in system
	ID       string
	VotedFor string
}

func (this *Actor) Init(id string, r RoleType, num_peers int, peers []string) {
	// TODO: init peers, logs, appendcounter
	this.CurrentTerm = 1
	this.Role = r
	this.NumPeers = num_peers
	this.NextIndicies = make(map[string]int)
	this.LastLogIndicies = make(map[string]int)
	this.AppendCounter = make([]int, 0)

	this.Logs = make([]Log, 0)
	this.VotedFor = ""

	for i := 0; i < this.NumPeers; i++ {
		this.NextIndicies[peers[i]] = 1
		this.LastLogIndicies[peers[i]] = 0
	}

	this.CommitIdx = 0
	this.Timeout = 100

	this.Peers = make([]string, len(peers))
	for i := 0; i < len(peers); i++ {
		this.Peers[i] = peers[i]
	}
}

func (this *Actor) CheckPrev(index, term int) bool {

	if this.Role != Follower {
		panic(errors.New("trying to check prev as non-follower"))
	}

	if index > len(this.Logs) {
		return false
	} else if index == 0 {
		return len(this.Logs) == 0
	} else if len(this.Logs) == 0 {
		return true
	}

	log := this.Logs[index-1]
	return log.Term == term

}

func (this *Actor) PrintLogs() {
	this.mu.Lock()
	defer this.mu.Unlock()
	for i := 0; i < len(this.Logs); i++ {
		fmt.Println(this.Logs[i].ToStr())
	}
}

func (this *Actor) HandleAppendEntriesRPC(rpc *AppendEntriesRPC) AppendResp {
	this.mu.Lock()
	defer this.mu.Unlock()
	// validity check ?

	if this.Role != Follower {
		panic(errors.New("trying to append entry as non-follower"))
	}

	if rpc.Term < this.CurrentTerm {
		fmt.Println("rpc term less than current term, refuse")
		return AppendResp{this.CurrentTerm, false}
	}

	if len(rpc.Entries) == 0 {
		// TODO: reset heartbeat timer

		fmt.Println("heartbeat received")
		return AppendResp{-1, false}
	}

	// return failure if log does not contain an entry at prevLogIndex whose
	//  term matches prevLogTerm
	if !this.CheckPrev(rpc.PrevLogIndex, rpc.PrevLogTerm) {

		fmt.Println("prev index does not match, refuse")
		return AppendResp{this.CurrentTerm, false}
	}

	if rpc.Term > this.CurrentTerm {
		this.CurrentTerm = rpc.Term
	}

	if rpc.PrevLogIndex < len(this.Logs)-1 {
		this.Logs = this.Logs[:rpc.PrevLogIndex]
	}
	this.Logs = append(this.Logs, DeepCopyLogs(rpc.Entries)...)

	fmt.Println("append rpc success")
	return AppendResp{this.CurrentTerm, true}

}

// leader reaction to follower responses for append rpcs
// is this even actually used ...? can't remember :(
func (this *Actor) HandleResp(newPrevIdx int, resp *RespPayload) bool {

	this.mu.Lock()
	defer this.mu.Unlock()

	if this.Role != Leader {
		panic(errors.New("Trying to handle resp as non-leader"))
	}

	// special value -1 for responses to heartbeat messages
	if resp.Body.Term == -1 {
		return true
	}

	if resp.Body.Success == false {
		this.NextIndicies[resp.PeerID] -= 1
		return false
	}

	// TODO

	return true
}

// receive client command string and append to own log
func (this *Actor) ReceiveClientRequest(req string) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.Role != Leader {
		panic(errors.New("trying to process client command as non-leader"))
	}

	new_log := Log{this.CurrentTerm, len(this.Logs) + 1, req}
	this.Logs = append(this.Logs, new_log)
}
