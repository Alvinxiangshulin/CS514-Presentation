package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var EntryList [13]Log {{0,0,""},{1, 1, "add"},{1, 2, "add"},{1, 3, "add"},{4, 4, "add"},{4, 5, "add"},{5, 6, "add"},{5, 7, "add"},{6, 8, "add"},{6, 9, "add"},{6, 10, "add"},{7, 11, "add"},{7, 12, "add"}}
const FollowerNum = 3



func handleConnection(c net.Conn, chn1 chan AppendEntriesRPC, chn2 chan AppendResp) { //AEInput and AEOutput are handled in this func
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {

		w := bufio.NewWriter(c)
		defer c.Close()

		var RPC AppendEntriesRPC
		RPC =<- chn1				// a bunch of data?
		enc := gob.NewEncoder(w)		
		enc.Encode(&RPC)

		r := bufio.NewReader(c)    //timeout? from followes?
		defer c.Close()
		dec := gob.NewDecoder(r)
		var Resp AppendResp
		err := dec.Decode(&Resp)
		chn2 <- Resp
		if err != nil{
			fmt.Println("Unparsable gob information!")
		}
		fmt.Printf("Received : %+v", Resp)

		//c.Write([]byte(string(result)))
	}
	c.Close()
	return
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	var cn [FollowerNum]net.Conn
	charr1 := make([]chan AppendEntriesRPC, FollowerNum) 
	charr2 := make([]chan AppendResp, FollowerNum)

	for chn := 0; chn < FollowerNum; chn++ {
		charr1[chn] = make(chan AppendEntriesRPC, 10)
		charr2[chn] = make(chan AppendResp, 10)
	}

	for i := 0; i < FollowerNum; i++ { //connect to three clients
		var err error
		cn[i], err = l.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(cn[i], charr1[i], charr2[i])
	}

	term := 1
	leaderId := 1
	prevLogIndex := 0
	preLogTerm := 0
	commitIndex := 0
	var nextIndex [FollowerNum]int
	var prevIndex [FollowerNum]int 
	for i := 0; i < FollowerNum; i ++{
		temp := make([]commitlog,10)
		commitList = append(commitList, temp)
	}
	var RPC [FollowerNum+1]AppendEntriesRPC
	var SendRPC AppendEntriesRPC

	for {
		select { //update state machine of leader and decide what to send to each client

		case r := <-charr2[0]:
			{
				if r.Success{
					nextIndex[0] = prevIndex[0] + 1
					commitIndex := prevIndex[0]
					for _, v := range preIndex {
						if v < commitIndex
							commitIndex = v
					}					
					RPC[0].CommitIndex = commitIndex
					RPC[0].Entries = EntryList[nextIndex[0]+1]
					RPC[0].LeaderId = leaderId
					RPC[0].PrevLogIndex = preIndex[0]
					RPC[0].PrevLogTerm = term - 1
					RPC[0].Term = term
					SendRPC = RPC[0]
					prevIndex[0] += 1
				}
				else{
					prevIndex[0] -= 1
					SendRPC = RPC[0]
				}
				charr1[0] <- SendRPC
			}
		case r := <-charr2[1]:
			{
				
				if r.Success{
					nextIndex[1] = prevIndex[1] + 1
					commitIndex := preIndex[1]
					for _, v := range prevIndex {
						if v < commitIndex
							commitIndex = v
					}
					
					RPC[1].CommitIndex = commitIndex
					RPC[1].Entries = EntryList[nextIndex[1]+1]
					RPC[1].LeaderId = leaderId
					RPC[1].PrevLogIndex = preIndex[1]
					RPC[1].PrevLogTerm = term - 1
					RPC[1].Term = term
					SendRPC = RPC[1]
					prevIndex[1] += 1
				}
				else{
					prevIndex[1] -= 1
					SendRPC = RPC[1]
				}
				charr1[1] <- SendRPC
			}
		case r := <-charr2[2]:
			{
				
				if r.Success{
					nextIndex[2] = prevIndex[2] + 1
					commitIndex := preIndex[2]
					for _, v := range prevIndex {
						if v < commitIndex
							commitIndex = v
					}
					
					RPC[2].CommitIndex = commitIndex
					RPC[2].Entries = EntryList[nextIndex[2]+1]
					RPC[2].LeaderId = leaderId
					RPC[2].PrevLogIndex = preIndex[2]
					RPC[2].PrevLogTerm = term - 1
					RPC[2].Term = term
					SendRPC = RPC[2]
					prevIndex[2] += 1
				}
				else{
					prevIndex[2] -= 1
					SendRPC = RPC[2]
				}
				charr1[2] <- SendRPC
			}

		default:
			{
				time.Sleep(10)					//heartbeat
				RPC[3].CommitIndex = 0
				RPC[3].Entries = EntryList[0]
				RPC[3].LeaderId = leaderId
				RPC[3].PreLogIndex = 0
				RPC[3].PrevLogTerm = term - 1
				RPC[3].Term = term
				SendRPC = RPC[3]
				for chn := 0; chn < 3; chn++ {
					charr1[chn] <- SendRPC 
				}
			}

		}
	}

}
