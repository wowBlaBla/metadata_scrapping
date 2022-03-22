package main

import (
	"fmt"
	"os"
	"time"
)

var file *os.File

func main() {
	loadConfig(&config)

	// get the new CID from pending tx, if config.UpdateCID is true
	if config.Update_CID {
		connect_wss()
		startMonitoring()

		txInput := handle_wss()
		if txInput == "failed" {
			fmt.Println("Error occured")
			return
		}

		config.Ipfs_info.CID = getCID(txInput)
		updateConfigFile()
	}

	// initialize the variables
	total_cnt := config.Ipfs_info.Total_Count
	thread_cnt := config.Ipfs_info.Thread_Count
	ch_result := make(chan bool, thread_cnt)

	// open the metadata.txt file
	var err error
	file, err = os.OpenFile("./output/metadata.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	// write the current time to metadata.txt file
	str_time := time.Now().String()
	file.WriteString("\n-------------------- " + str_time + " --------------------\n")

	// generate multithread for requesting the metadata from IPFS
	req_cnt := total_cnt / thread_cnt
	for i := 0; i < thread_cnt; i++ {
		if i < thread_cnt-1 {
			go getMetadata(config.Ipfs_info.Start_Index+i*10, req_cnt, ch_result)
		} else {
			go getMetadata(config.Ipfs_info.Start_Index+i*10, req_cnt+(total_cnt%thread_cnt), ch_result)
		}
	}

	// parse and write the metadata to file
	for i := 0; i < thread_cnt; i++ {
		<-ch_result
	}

	// write the current time to metadata.txt file
	str_time = time.Now().String()
	file.WriteString("-------------------- " + str_time + " --------------------\n")

	// close file
	file.Close()

	fmt.Println(time.Now().String())
}
