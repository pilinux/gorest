package model

// ErrorMsg ...
type ErrorMsg struct {
	HTTPCode int    `structs:"httpResponseCode" json:"-"`
	Message  string `structs:"msg" json:"msg"`
}
