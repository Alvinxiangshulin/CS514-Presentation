package main

// import (
// 	"fmt"
// 	"sync"
// )

// type Follower struct {
// 	mu          sync.Mutex
// 	CurrentTerm int
// 	Logs        []Log
// 	Timeout     int
// }

// func (this *Follower) Init() {
// 	this.mu.Lock()
// 	defer this.mu.Unlock()
// 	this.CurrentTerm = 4
// 	this.Logs = make([]Log, 0)
// 	this.Logs = append(this.Logs, Log{1, 1, "add"})
// 	this.Logs = append(this.Logs, Log{1, 2, "add"})
// 	this.Logs = append(this.Logs, Log{1, 3, "add"})
// 	this.Logs = append(this.Logs, Log{4, 4, "add"})
// 	this.Logs = append(this.Logs, Log{4, 5, "add"})
// 	this.Logs = append(this.Logs, Log{5, 6, "add"})
// 	this.Logs = append(this.Logs, Log{5, 7, "add"})
// 	this.Logs = append(this.Logs, Log{6, 8, "add"})
// 	this.Logs = append(this.Logs, Log{6, 9, "add"})
// 	this.Logs = append(this.Logs, Log{6, 10, "add"})
// 	this.Logs = append(this.Logs, Log{7, 11, "add"})
// 	this.Logs = append(this.Logs, Log{7, 12, "add"})

// 	this.Timeout = 100
// }

// func (this *Follower) CheckPrev(index, term int) bool {
// 	if index > len(this.Logs) {
// 		return false
// 	} else if index == 0 {
// 		return len(this.Logs) == 0
// 	} else if len(this.Logs) == 0 {
// 		return true
// 	}

// 	log := this.Logs[index-1]
// 	return log.Term == term

// }

// func (this *Follower) PrintLogs() {
// 	for i := 0; i < len(this.Logs); i++ {
// 		fmt.Println(this.Logs[i].ToStr())
// 	}
// }

// func (this *Follower) HandleAppendEntriesRPC(rpc *AppendEntriesRPC) AppendResp {
// 	this.mu.Lock()
// 	defer this.mu.Unlock()
// 	// validity check ?

// 	if rpc.Term < this.CurrentTerm {
// 		//resp.Resps = append(responses.Resps, AppendResp{this.CurrentTerm, false})
// 		//continue
// 		return AppendResp{this.CurrentTerm, false}
// 	}

// 	if len(rpc.Entries) == 0 {
// 		// TODO: reset timer
// 		//continue
// 		return AppendResp{-1, false}
// 	}

// 	// return failure if log does not contain an entry at prevLogIndex whose
// 	//  term matches prevLogTerm
// 	if !this.CheckPrev(rpc.PrevLogIndex, rpc.PrevLogTerm) {
// 		//responses.Resps = append(responses.Resps, AppendResp{this.CurrentTerm, false})
// 		//continue
// 		return AppendResp{this.CurrentTerm, false}
// 	}

// 	if rpc.Term > this.CurrentTerm {
// 		this.CurrentTerm = rpc.Term
// 	}

// 	if rpc.PrevLogIndex < len(this.Logs)-1 {
// 		this.Logs = this.Logs[:rpc.PrevLogIndex]
// 	}
// 	this.Logs = append(this.Logs, DeepCopyLogs(rpc.Entries)...)
// 	return AppendResp{this.CurrentTerm, true}

// }

// /*
// func main() {
// 	//data, err := ParseAppendReqFromFile("test_input/test_follower_same.json")
// 	data, err := ParseRPCAndRespFromFile("test_input/test_extra.json")
// 	reqs := data.Rpcs
// 	//resp := data.Resps
// 	if err == nil {
// 		//PrintAppendReqs(&reqs)
// 	} else {
// 		fmt.Println("err")
// 	}

// 	responses := Responses{}
// 	follower := Follower{}
// 	follower.Init()

// 	req_arr := AppendReqs{}
// 	req_arr.Rpcs = reqs
// 	follower.HandleAppendEntriesRPC(&req_arr, &responses)

// 	// open an output file
// 	out_file, out_file_err := os.OpenFile("follower_out_extra.json", os.O_CREATE|os.O_WRONLY, 0777)
// 	if out_file_err != nil {
// 		panic(out_file_err)
// 	}

// 	PrintResps(&responses)
// 	follower.PrintLogs()
// 	//PrintResps(&resp)

// 	writter := bufio.NewWriter(out_file)
// 	barr, json_err := json.MarshalIndent(responses, "", "    ")
// 	// fmt.Print(barr)
// 	// fmt.Println(len(responses.Resps))
// 	if json_err != nil {
// 		panic(json_err)
// 	}
// 	if _, write_err := writter.Write(barr); err != nil {
// 		panic(write_err)
// 	}
// 	writter.Flush()
// }
// */
