package main

import "fmt"

func main() {
	data, err := ParseAppendReqFromFile("test_input/test_follower_same.json")
	if err == nil {
		PrintAppendReqs(&data)
	} else {
		fmt.Println("err")
	}

}
