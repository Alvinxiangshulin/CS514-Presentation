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
	Entries      []string
	CommitIndex  int
}

type AppendResp struct {
	Term    int
	Success bool
}

type AppendReqs struct {
	Rpcs []AppendEntriesRPC
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

func PrintAppendReqs(data *AppendReqs) {
	for i := 0; i < len(data.Rpcs); i++ {
		req := data.Rpcs[i]
		fmt.Printf("term: %d, leaderId: %d, prevLogIndex: %d, prevLogTerm: %d, entries: ", req.Term, req.LeaderId, req.PrevLogIndex, req.PrevLogTerm)
		fmt.Print(req.Entries)
		fmt.Printf(", commitIndex: %d\n", req.CommitIndex)
	}
}
