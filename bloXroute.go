package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var wss *websocket.Conn

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
	methodID := getMethodId(config.Watch_function.Func_ProtoType)

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
