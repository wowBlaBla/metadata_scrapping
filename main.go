package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Printf("\n@@@ Start Time: %s\n\n", time.Now().String())
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

	// open the metadata.txt file
	file, err := os.OpenFile("./output/metadata.json", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	// start to fetch the metadata from IPFS
	startFetchingMetadata(file)

	// close file
	file.Close()

	fmt.Printf("\n@@@ End Time: %s\n\n", time.Now().String())
}
