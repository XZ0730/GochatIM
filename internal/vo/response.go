package vo

type Response struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Error  string      `json:"err"`
}
