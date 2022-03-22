package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// parse the input data in transaction
func decodeTxParams(txInput string) []string {
	myContractAbi, err := os.ReadFile("./setting/abi.json") // read NFT smart contract abi from abi.json
	if err != nil {
		log.Fatal(err)
	}

	abi, err := abi.JSON(strings.NewReader(string(myContractAbi))) // load contract ABI
	if err != nil {
		log.Fatal(err)
	}

	decodedSig, err := hex.DecodeString(txInput[2:10]) // decode txInput method signature
	if err != nil {
		log.Fatal(err)
	}

	method, err := abi.MethodById(decodedSig) // recover Method from signature and ABI
	if err != nil {
		log.Fatal(err)
	}

	decodedData, err := hex.DecodeString(txInput[10:]) // decode txInput Payload
	if err != nil {
		log.Fatal(err)
	}

	data, err := method.Inputs.Unpack(decodedData) // unpack method inputs
	if err != nil {
		log.Fatal(err)
	}

	// parse the parameters into string
	fmt.Println("\n------------ " + config.Watch_function.Func_ProtoType + " ------------")

	params := make([]string, len(data))
	for i := 0; i < len(data); i++ {
		params[i] = fmt.Sprintf("%v", data[i])
		fmt.Printf("Parameter%d: %s\n", i+1, params[i])
	}
	fmt.Println("")

	return params
}

// Extract the CID from parameters of watch_function
func getCID(txInput string) string {
	params := decodeTxParams(txInput) // decode the params from input data of transaction

	// extract the CID from params
	var param string
	for i := 0; i < len(params); i++ {
		param = params[i]
		if strings.Contains(param, "ipfs") {
			param = strings.Replace(param, "http", "", -1)
			param = strings.Replace(param, "https", "", -1)
			param = strings.Replace(param, "ipfs", "", -1)
			param = strings.Replace(param, ".io", "", -1)
			param = strings.Replace(param, "/", "", -1)
			param = strings.Replace(param, ":", "", -1)
			param = strings.Replace(param, ".", "", -1)
			break
		}
	}

	fmt.Println("CID:", param)
	return param
}
