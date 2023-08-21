package com

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"os"
)

func NewCOM(identifier string) *COM {
	com := COM{}
	com.Identifier = identifier
	com.initBusMqttClient()
	return &com
}

func (root *COM) Reply(plainRequest *PlainRequest, content any) {
	plainResponse := createResponse(plainRequest, content)
	payload, _ := json.Marshal(plainResponse)
	if plainRequest.PublicKey == "" {
		root.MqttClientInt.Publish("clients", 0, false, payload).Wait()
	} else {
		root.MqttClientInt.Publish(plainRequest.ResponseTopic, 0, false, payload).Wait()
	}

}

func createResponse(plainRequest *PlainRequest, content any) PlainResponse {
	plain, _ := json.Marshal(content)
	plainResponse := PlainResponse{
		PublicKey:     plainRequest.PublicKey,
		CorrelationId: plainRequest.CorrelationId,
		Plain:         plain,
		ResponseTopic: plainRequest.ResponseTopic,
	}
	return plainResponse
}

func (root *COM) initBusMqttClient() {
	opts1 := mqtt.NewClientOptions()
	uuid, _ := uuid.NewUUID()
	opts1.SetClientID(uuid.String())
	bus := os.Getenv("BUS")
	if bus == "" {
		bus = "mqtt://localhost:1883"
	}
	opts1.AddBroker(bus)
	root.MqttClientInt = mqtt.NewClient(opts1)
	root.MqttClientInt.Connect().Wait()
}

func (root *COM) Publish(eventName string, plainResponse any) {
	fmt.Printf("Bus Publish Message To: %s\n", eventName)
	payload, _ := json.Marshal(plainResponse)
	root.MqttClientInt.Publish(eventName, 0, false, payload).Wait()
}

func (root *COM) Subscribe(eventName string, callback func(rawMessage json.RawMessage)) {
	fmt.Printf("Bus Subscribe: %s\n", eventName)
	root.MqttClientInt.Subscribe(eventName, 0, func(client mqtt.Client, message mqtt.Message) {
		fmt.Printf("Bus Received Message For: %s\n", eventName)
		rawMessage := json.RawMessage{}
		json.Unmarshal(message.Payload(), &rawMessage)
		callback(rawMessage)
	}).Wait()
}
