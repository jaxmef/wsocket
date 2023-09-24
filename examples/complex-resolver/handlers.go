package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jaxmef/wsocket"
)

type sumRequest struct {
	Type string `json:"type"`
	A    int    `json:"a"`
	B    int    `json:"b"`
}

func handleSum(_ context.Context, msg []byte, rw wsocket.ResponseWriter) error {
	jsonMsg := &sumRequest{}
	if err := json.Unmarshal(msg, jsonMsg); err != nil {
		return err
	}

	result := jsonMsg.A + jsonMsg.B
	response := []byte(fmt.Sprintf(`{"type": "sum-response", "result": %d}`, result))
	if err := rw.WriteMessage(wsocket.NewTextMessage(response)); err != nil {
		return err
	}

	return nil
}

type event struct {
	Type string `json:"type"`
	Data struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"data"`
}

func handleInfoEvent(_ context.Context, msg []byte, rw wsocket.ResponseWriter) error {
	jsonMsg := &event{}
	if err := json.Unmarshal(msg, jsonMsg); err != nil {
		return err
	}

	log.Printf("[INFO] %s\n", jsonMsg.Data.Message)

	return nil
}

func handleErrorEvent(_ context.Context, msg []byte, rw wsocket.ResponseWriter) error {
	jsonMsg := &event{}
	if err := json.Unmarshal(msg, jsonMsg); err != nil {
		return err
	}

	log.Printf("[ERROR] %s\n", jsonMsg.Data.Message)

	return nil
}
