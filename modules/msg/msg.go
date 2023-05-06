package msg

type RequestMsg struct {
	Type      int64  `json:"type"`
	ProxyName string `json:"proxy_name"`
	Passwd    string `json:"passwd"`
}

type ResponseMsg struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}
