package model

// ErrorMsg handles HTML and JSON error messages.
type ErrorMsg struct {
	HTTPCode int    `structs:"httpResponseCode" json:"-"`
	Message  string `structs:"msg" json:"msg"`
}
