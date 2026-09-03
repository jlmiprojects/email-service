package models

import "encoding/json"

type Result struct {
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

func MakeResult(code int, message string, err error) Result {

	if err != nil {
		return Result{
			StatusCode: code, Message: message, Error: err.Error(),
		}
	} else {
		return Result{
			StatusCode: code, Message: message, Error: "",
		}
	}
}

func (u *Result) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}
