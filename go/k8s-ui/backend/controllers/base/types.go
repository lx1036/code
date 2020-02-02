package base

type JsonResponse struct {
	Errno  int         `json:"errno"`  // -1,0
	Errmsg string      `json:"errmsg"` // "success" or "failed: xxx"
	Data   interface{} `json:"data"`   // struct{}
}
