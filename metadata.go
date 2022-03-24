package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// send request to public IPFS gateway and receive the response
func getMetadata(file *os.File, start int, count int, ch_result chan bool) {
	for i := start; i < start+count; i++ {
		// make request url
		url := "https://ipfs.io/ipfs/"
		url += config.Ipfs_info.CID
		url += "/"
		url += strconv.Itoa(i)
		url += ".json"

		// send request
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
			ch_result <- false
			return
		}

		// process resp
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
			ch_result <- false
			return
		}

		// write metadata
		writeMetadata(file, body)
	}

	ch_result <- true
}

// write the metadata to file
func writeMetadata(file *os.File, data []byte) {
	// parse the data
	var ret interface{}
	json.Unmarshal(data, &ret)

	// writing the result to file
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(ret)
	buf.WriteString(",")
	_, err := file.Write(buf.Bytes())

	if err != nil {
		fmt.Println("Error occured on writing metadata.txt:", err.Error())
	} else {
		fmt.Print(">")
	}
}

func startFetchingMetadata(file *os.File) {
	// initialize the variables
	total_cnt := config.Ipfs_info.Total_Count
	thread_cnt := config.Ipfs_info.Thread_Count
	ch_result := make(chan bool, thread_cnt)

	file.WriteString("[\n")

	// fetch the metadata from IPFS
	fmt.Println("\nStarting the fetching metadata from IPFS now")
	fmt.Println("-------------------- " + time.Now().String() + " --------------------")

	// generate multithread for requesting the metadata from IPFS
	req_cnt := total_cnt / thread_cnt
	for i := 0; i < thread_cnt; i++ {
		if i < thread_cnt-1 {
			go getMetadata(file, config.Ipfs_info.Start_Index+i*10, req_cnt, ch_result)
		} else {
			go getMetadata(file, config.Ipfs_info.Start_Index+i*10, req_cnt+(total_cnt%thread_cnt), ch_result)
		}
	}

	// parse and write the metadata to file
	for i := 0; i < thread_cnt; i++ {
		<-ch_result
	}

	// remove the last comma(",")
	fInfo, err := file.Stat()
	if err != nil {
		fmt.Println("error occured on file.Stat():", err.Error())
		return
	}
	file.Truncate(fInfo.Size() - 1)

	file.WriteString("]")

	fmt.Println("\n-------------------- " + time.Now().String() + " --------------------")
	fmt.Printf("Finished fetching metadata from IPFS\n\n")
}
