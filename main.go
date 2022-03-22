package main

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/gorilla/websocket"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

type METADATA struct {
	Name         string      `json:"name"`
	Image        string      `json:"image"`
	External_Url string      `json:"external_url"`
	Description  string      `json:"description"`
	Attributes   []ATTRIBUTE `json:"attributes"`
}

type ATTRIBUTE struct {
	Type  string `json:"trait_type"`
	Value string `json:"value"`
}

type RPC struct {
	Https string `json:"https"`
	Wss   string `json:"wss"`
}

type WATCH_FUNCTION struct {
	Func_ProtoType string `json:"func_prototype"`
	From           string `json:"from"`
	To             string `json:"to"`
}

type IPFS_INFO struct {
	CID          string `json:"cid"`
	Start_Index  int    `json:"start_index"`
	Total_Count  int    `json:"total_count"`
	Thread_Count int    `json:"thread_count"`
}

type Config struct {
	Update_CID          bool           `json:"update_cid"`
	BloxRouteAuthHeader string         `json:"bloxRouteAuthHeader"`
	Rpc                 RPC            `json:"rpc"`
	Watch_function      WATCH_FUNCTION `json:"watch_function"`
	Ipfs_info           IPFS_INFO      `json:"ipfs_info"`
}

var config Config
var file *os.File
var wss *websocket.Conn

// load config from 'config.json' file
func loadConfig(config *Config) {
	_config, _ := os.ReadFile("./config.json")
	err := json.Unmarshal([]byte(_config), config)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Configuration has been read successfully!")
	}
}

// update config file with new config
func updateConfigFile() {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(config)

	err := os.WriteFile("./config.json", buf.Bytes(), os.ModeType)
	if err == nil {
		fmt.Println("Configuration has been updated successfully!")
	}
}

// establish the connection with websocket and bloXroute
func connect_wss() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialer_wss := websocket.DefaultDialer
	dialer_wss.TLSClientConfig = tlsConfig

	var err error
	wss, _, err = dialer_wss.Dial(config.Rpc.Wss, http.Header{"Authorization": []string{config.BloxRouteAuthHeader}})
	if err != nil {
		fmt.Println("WebSocket_wss connection failed:", err.Error())
	}
}

// start monitoring the pending txs from bloXroute
func startMonitoring() {
	// calculate the methodID from function proto type string
	_tmp := solsha3.SoliditySHA3(config.Watch_function.Func_ProtoType)
	_hash_str := hex.EncodeToString(_tmp)
	methodID := "0x" + _hash_str[0:8]

	// make the subscribe request
	subRequest := `{"jsonrpc": "2.0", "id": 1, "method": "subscribe", "params": ["newTxs", {"include": [], "filters": "{method_id} == '` + methodID + `'`

	if config.Watch_function.From != "" {
		subRequest = subRequest + ` AND ({from} == '` + config.Watch_function.From + `')`
	}

	if config.Watch_function.To != "" {
		subRequest = subRequest + ` AND ({to} == '` + config.Watch_function.To + `')`
	}

	subRequest = subRequest + `"}]}`

	// send subRequest to bloXroute
	err := wss.WriteMessage(websocket.TextMessage, []byte(subRequest))
	if err != nil {
		fmt.Println("WebSocket_bloWS sent failed:", err.Error())
	}
}

// receive the message from bloXroute and process it
func handle_wss() string {
	fmt.Println("Listening to bloXroute")

	for {
		// read message from bloXroute
		_, resp, err := wss.ReadMessage()
		if err != nil {
			fmt.Println("handle_wss:", err.Error())
			return "failed"
		}

		// parsing received message
		var data map[string]interface{}
		json.Unmarshal(resp, &data)

		data_params := data["params"].(map[string]interface{})
		data_params_result := data_params["result"].(map[string]interface{})
		data_params_result_txContents := data_params_result["txContents"].(map[string]interface{})
		_input := data_params_result_txContents["input"].(string)

		return _input
	}
}

// parse the input data in transaction
func decodeTxParams(txInput string) []string {
	// read NFT smart contract abi from abi.json
	myContractAbi, err := os.ReadFile("abi.json")
	if err != nil {
		log.Fatal(err)
	}

	// load contract ABI
	abi, err := abi.JSON(strings.NewReader(string(myContractAbi)))
	if err != nil {
		log.Fatal(err)
	}

	// decode txInput method signature
	decodedSig, err := hex.DecodeString(txInput[2:10])
	if err != nil {
		log.Fatal(err)
	}

	// recover Method from signature and ABI
	method, err := abi.MethodById(decodedSig)
	if err != nil {
		log.Fatal(err)
	}

	// decode txInput Payload
	decodedData, err := hex.DecodeString(txInput[10:])
	if err != nil {
		log.Fatal(err)
	}

	// unpack method inputs
	data, err := method.Inputs.Unpack(decodedData)
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
	// var ret METADATA
	var ret interface{}
	json.Unmarshal(data, &ret)

	// writing the result to file
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(ret)
	_, err := file.Write(buf.Bytes())

	if err == nil {
		fmt.Printf("%+v\n", ret)
	}
}

// main go routine
func main() {
	// load config
	loadConfig(&config)

	// get the new CID from pending tx, if config.UpdateCID is true
	if config.Update_CID {
		// connect websocket
		connect_wss()

		// start monitoring the pending transactions
		startMonitoring()

		// receive message from bloXroute process the message
		txInput := handle_wss()
		if txInput == "failed" {
			fmt.Println("Error occured")
			return
		}

		// decode the params from input data of transaction
		params := decodeTxParams(txInput)

		// extract the CID from params
		for i := 0; i < len(params); i++ {
			param := params[i]
			if strings.Contains(param, "ipfs") {
				param = strings.Replace(param, "http", "", -1)
				param = strings.Replace(param, "https", "", -1)
				param = strings.Replace(param, "ipfs", "", -1)
				param = strings.Replace(param, ".io", "", -1)
				param = strings.Replace(param, "/", "", -1)
				param = strings.Replace(param, ":", "", -1)
				param = strings.Replace(param, ".", "", -1)

				fmt.Println("CID:", param)
				config.Ipfs_info.CID = param
			}
		}

		// update the config file
		updateConfigFile()
	}

	// initialize the variables
	total_cnt := config.Ipfs_info.Total_Count
	thread_cnt := config.Ipfs_info.Thread_Count
	ch_result := make(chan bool, thread_cnt)

	// open the metadata.txt file
	var err error
	file, err = os.OpenFile("metadata.txt", os.O_APPEND|os.O_WRONLY, 0644)
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
