package vo

import "fmt"

type ResponseList struct {
	Item  interface{} `json:"item"`
	Total int         `json:"total"`
	Extra interface{} `json:"extra"`
}

func BuildList(data interface{}, leng int, extra interface{}) (rsp Response) {
	fmt.Println("data:", data)
	rsp = Response{
		Status: 200,
		Data: ResponseList{
			Item:  data,
			Total: leng,
			Extra: extra,
		},
		Msg: "ok",
	}
	fmt.Println("rsp:", rsp)
	return
}
