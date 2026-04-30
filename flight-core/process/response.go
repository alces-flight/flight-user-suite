package process

import "encoding/json"

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (response *Response) ToJSON() (string, error) {
	json, err := json.Marshal(response)
	if err != nil {
		return "", err
	}
	return string(json), nil
}
