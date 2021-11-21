package main

import (
    "fmt"
    "net/http"
    "encoding/json"
)

type myData struct {     // example struct 
	Name string
	Address  string
}

func hello(w http.ResponseWriter, r *http.Request) {

    switch r.Method {
	case "GET":
		fmt.Fprintf(w, "hello--------\n")	
	case "POST":
		decoder := json.NewDecoder(r.Body)

		var data myData
		err := decoder.Decode(&data)
		if err != nil {
		   panic(err)
		}	   

		address := data.Address
		name := data.Name
		fmt.Println("name: ", name)
		fmt.Println("address: ", address)
		fmt.Fprintf(w, "Post request found with name=%s, address=%s", name, address)

	default:
		fmt.Fprintf(w, "Request method %s is not supported", r.Method)
	}
}

func main() {

    http.HandleFunc("/hello", hello)
    http.ListenAndServe("67.159.89.3:8090", nil)
}

