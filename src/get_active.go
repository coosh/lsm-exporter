package main

import (
	"encoding/json"
	"log"
)

type RunningModel struct {
	Model string `json:"model"`
	Proxy string `json:"proxy"`
	State string `json:"state"`
}

type RunningResponse struct {
	Running []RunningModel `json:"running"`
}

func getActiveModels() []RunningModel {
	body, err := fetch(llamaSwapURL + "/running")
	if err != nil {
		log.Printf("failed to fetch llama-swap /running: %v", err)
		return nil
	}

	var resp RunningResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		log.Printf("failed to parse /running response: %v", err)
		return nil
	}

	return resp.Running
}
