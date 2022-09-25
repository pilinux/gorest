package model

// ErrorMsg - to handle HTML + JSON
type ErrorMsg struct {
	HTTPCode int    `structs:"httpResponseCode" json:"-"`
	Message  string `structs:"msg" json:"msg"`
}
