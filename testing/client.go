package main

import (
    "bytes"	
    "fmt"
    "net/http"
    "io/ioutil"
    "encoding/json"
)

type Student struct {    // struct example 
        Name    string `json:"name"`
        Address string `json:"address"`
}

func get(httpurl string) (string) {
     resp, err := http.Get(httpurl)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("Response status:", resp.Status)
    bodyGet, _ := ioutil.ReadAll(resp.Body)
    return string(bodyGet)
}

func post(postData *bytes.Buffer, httpurl string ) (string){
     req, _ := http.NewRequest("POST", httpurl, postData)
     client := &http.Client{}       
     res, _ := client.Do(req)
     defer res.Body.Close()
     fmt.Println("response Status:", res.Status)

     bodyPost, _ := ioutil.ReadAll(res.Body)
     return string(bodyPost)
}



func main() {
     httpurl := "http://67.159.89.3:8090/hello"
     fmt.Println("HTTP JSON URL:", httpurl)

    // get
    getResp := get(httpurl)
    fmt.Println("response Body:", getResp)

    // post
    body := &Student{
                Name:    "abc",
                Address: "xyz",
        }
    buf := new(bytes.Buffer)
    json.NewEncoder(buf).Encode(body)
    postResp := post(buf, httpurl)
    fmt.Println("response Body:", postResp)
}