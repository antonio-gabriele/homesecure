package com

import (
	"encoding/json"
	mqttc "github.com/eclipse/paho.mqtt.golang"
)

type PairRequest struct {
	PublicKey string `json:"pk,omitempty"`
}

type PairResponse struct {
	PublicKey string `json:"pk,omitempty"`
}

type ResultResponse struct {
	EventName string `json:"eventName,omitempty"`
	Result    any    `json:"result,omitempty"`
}

type CipherRequest struct {
	PublicKey string `json:"pk,omitempty"`
	IV        string `json:"iv,omitempty"`
	Cipher    string `json:"cipher,omitempty"`
}

type COM struct {
	Identifier    string
	MqttClientInt mqttc.Client
}

type PlainRequest struct {
	PublicKey     string          `json:"pk,omitempty"`
	CorrelationId string          `json:"cid,omitempty"`
	Topic         string          `json:"topic,omitempty"`
	ResponseTopic string          `json:"responseTopic,omitempty"`
	Plain         json.RawMessage `json:"plain,omitempty"`
}

type PlainResponse struct {
	PublicKey     string          `json:"pk,omitempty"`
	CorrelationId string          `json:"cid,omitempty"`
	Plain         json.RawMessage `json:"plain,omitempty"`
	ResponseTopic string          `json:"responseTopic,omitempty"`
}

type BroadcastResponse struct {
	Plain json.RawMessage `json:"plain,omitempty"`
}

type CipherResponse struct {
	Cipher string `json:"cipher,omitempty"`
	IV     string `json:"iv,omitempty"`
}

type PingRequest struct {
	Ping string `json:"ping"`
}

type PongResponse struct {
	Pong string `json:"pong"`
}
