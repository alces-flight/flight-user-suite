package process

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"
)

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

func (response *Response) WriteToParentFD() error {
	responseJson, _ := response.ToJSON()

	// The additional file handle passed from parent process will be ID 3
	responseFile, err := os.OpenFile("/proc/self/fd/3", os.O_WRONLY, 0644)
	if err != nil {
		if !strings.Contains(err.Error(), "no such device or address") {
			return fmt.Errorf("failed to return response to calling process: %w", err)
		}
	} else {
		fmt.Fprintln(responseFile, responseJson)
		responseFile.Close()
	}
	return nil
}

func IsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
