package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type AppendEntriesRPC struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Log
	CommitIndex  int
}

func (this *AppendEntriesRPC) PrintEntries() {
	for i := range this.Entries {
		fmt.Println(this.Entries[i].ToStr())
	}
}

type AppendResp struct {
	Term    int
	Success bool
}

type Responses struct {
	Resps []AppendResp
}

type AppendReqs struct {
	Rpcs []AppendEntriesRPC
}

type Log struct {
	Term    int
	Index   int
	Command string
}

type LeaderLog struct{
	Requests AppendReqs
	Resps Responses
}


func (this *Log) ToStr() string {
	return fmt.Sprintf("[Term: %d, Index: %d, command: %s]", this.Term, this.Index, this.Command)
}

func DeepCopyLogs(logs []Log) []Log {
	result := make([]Log, len(logs))
	for i := range logs {
		result[i] = Log{}
		result[i].Index = logs[i].Index
		result[i].Term = logs[i].Term
		result[i].Command = logs[i].Command
	}

	return result
}

// Unmarshal byte array into struct holding all the requests, propagate error
//  if necessary
func ParseAppendReqFromFile(filename string) (AppendReqs, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return AppendReqs{}, err
	}

	requests := AppendReqs{}
	json.Unmarshal(data, &requests)
	return requests, nil
}


//parse missing leader log
func ParseFroMissingLeaderFile(filename string) (LeaderLog, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return LeaderLog{}, err
	}

	payload := LeaderLog{}
	json.Unmarshal(data, &payload)
	return payload, nil
}

func PrintAppendReqs(data *AppendReqs) {
	for i := 0; i < len(data.Rpcs); i++ {
		req := data.Rpcs[i]
		fmt.Printf("term: %d, leaderId: %d, prevLogIndex: %d, prevLogTerm: %d, commitIndex: %d, entries:\n", req.Term, req.LeaderId, req.PrevLogIndex, req.PrevLogTerm, req.CommitIndex)
		// fmt.Print(req.Entries)
		req.PrintEntries()
	}
}

func PrintResps(responses *Responses) {
	for i := range responses.Resps {
		fmt.Printf("{Term: %d, Success: %t}\n", responses.Resps[i].Term, responses.Resps[i].Success)
	}
}
