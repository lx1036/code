package errors

type ErrorResult struct {
    // http code
    Code int `json:"code"`
    // The custom code
    SubCode int    `json:"subCode"`
    Msg     string `json:"msg"`
}
