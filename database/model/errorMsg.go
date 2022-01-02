package model

// ErrorMsg ...
type ErrorMsg struct {
	HTTPCode int    `structs:"http_response_code" json:"-"`
	Message  string `structs:"msg" json:"msg"`
}
