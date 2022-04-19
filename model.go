package main

import "encoding/json"

type VolCallbackEvent struct {
	RequestID string          `json:"RequestId"`
	Version   string          `json:"Version"`
	EventType string          `json:"EventType"`
	EventTime string          `json:"EventTime"`
	Data      json.RawMessage `json:"Data"`
}
