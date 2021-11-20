package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func LeaderUpdate(state int) {
	return
}

func handleConnection(c net.Conn, chn chan int) { //AEInput and AEOutput are handled in this func
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		var Resp AppendResp
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}

		result := strconv.Itoa(rand.Int()) + "\n"
		c.Write([]byte(string(result)))
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

	var cn [3]net.Conn
	charr := make([]chan int, 10) //channel is used to transfer log info from a client,
	// which will help leader's state machine update.
	// And send back message to client

	for chn := 0; chn < 3; chn++ {
		charr[chn] = make(chan int)
	}

	for i := 0; i < 3; i++ { //connect to three clients
		var err error
		cn[i], err = l.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(cn[i], charr[i])
	}

	for {
		select { //update state machine of leader and decide what to send to each clients
		case r := <-charr[0]:
			{
				//update state machine
				if r == 1 {

				} else {

				}
				charr[0] <- 1
			}
		case r := <-charr[1]:
			{
				//update state machine
				if r == 1 {

				} else {

				}
				charr[1] <- 1
			}
		case r := <-charr[2]:
			{
				//update state machine
				if r == 1 {

				} else {

				}
				charr[2] <- 1
			}

		default:
			{
				//heartbeat broadcast
				time.Sleep(10)
				for chn := 0; chn < 3; chn++ {
					charr[chn] <- 0
				}
			}

		}
	}

}
