package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// send request to public IPFS gateway and receive the response
func getMetadata(start int, count int, ch_result chan bool) {
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
		writeMetadata(body)
	}

	ch_result <- true
}

// write the metadata to file
func writeMetadata(data []byte) {
	// parse the data
	var ret interface{}
	json.Unmarshal(data, &ret)

	// writing the result to file
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(ret)
	_, err := file.Write(buf.Bytes())

	if err == nil {
		// fmt.Printf("%+v\n", ret)
	}
}
