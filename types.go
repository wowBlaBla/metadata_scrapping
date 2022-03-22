package main

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
