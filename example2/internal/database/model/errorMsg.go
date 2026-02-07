package model

// ErrorMsg holds an error message for HTML and JSON responses.
type ErrorMsg struct {
	HTTPCode int    `structs:"httpResponseCode" json:"-"`
	Message  string `structs:"msg" json:"msg"`
}
