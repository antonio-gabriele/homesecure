package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gadu/shared/m/v2/com"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"io"
	"log"

	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	logger := log.New(os.Stderr, "GW: ", log.LstdFlags)
	logger.Println("Starting...")
	NewGW("a1b2c3", logger)
	logger.Println("Started...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logger.Println("Ending...")
}

func NewGW(identifier string, logger *log.Logger) {
	root := GW{}
	uuid := uuid.NewString()
	root.ResponseTopic = fmt.Sprintf("client/%s", uuid)
	root.logger = logger
	root.logger.Println("Starting...")
	root.com = com.NewCOM(identifier)
	root.cache = map[string]Cache{}
	root.Identifier = identifier
	root.initWanMqttClient(uuid)
	root.com.Subscribe(root.ResponseTopic, root.send)
	root.com.Subscribe("clients", root.sendToAll)
}

func (root *GW) initWanMqttClient(uuid string) {
	opts1 := mqtt.NewClientOptions()
	wan := os.Getenv("WAN")
	if wan == "" {
		wan = "wss://user:bitnami@homesecure.dev:15676/ws"
	}
	opts1.AddBroker(wan)
	opts1.SetClientID(uuid)
	opts1.OnConnect = root.connect
	root.mqttClientWan = mqtt.NewClient(opts1)
	root.mqttClientWan.Connect()
}

func (root *GW) connect(client mqtt.Client) {
	client.Subscribe(fmt.Sprintf("%s/edge/pair", root.Identifier), 0, root.pair).Wait()
	client.Subscribe(fmt.Sprintf("%s/edge/recv", root.Identifier), 0, root.recv).Wait()
}

func (root *GW) sendToAll(rawMessage json.RawMessage) {
	plainResponse := PlainResponse{}
	json.Unmarshal(rawMessage, &plainResponse)
	for publicKey := range root.cache {
		plainResponse1 := PlainResponse{PublicKey: publicKey, Plain: plainResponse.Plain}
		fmt.Printf("sending: %s\n", plainResponse1.PublicKey)
		root.innerSend(plainResponse1)
	}
}

func (root *GW) send(rawMessage json.RawMessage) {
	plainResponse := PlainResponse{}
	json.Unmarshal(rawMessage, &plainResponse)
	fmt.Printf("sending: %s\n", plainResponse.PublicKey)
	root.innerSend(plainResponse)
}

func (root *GW) innerSend(plainResponse PlainResponse) {
	_, contains := root.cache[plainResponse.PublicKey]
	if contains == false {
		return
	}
	cipherResponse := CipherResponse{}
	codec, contains := root.cache[plainResponse.PublicKey]
	if contains == false {
		return
	}
	root.codec(&codec)
	root.encrypt(&codec, &plainResponse, &cipherResponse)
	cipher, _ := json.Marshal(cipherResponse)
	publicKey := strings.Replace(plainResponse.PublicKey, "/", "", -1)
	publicKey = strings.Replace(publicKey, "+", "", -1)
	publicKey = strings.Replace(publicKey, "=", "", -1)
	root.mqttClientWan.Publish(fmt.Sprintf("%s/%s/recv", root.Identifier, publicKey), 0, false, cipher)
}

func (root *GW) pair(_ mqtt.Client, message mqtt.Message) {
	request := message.Payload()
	pairRequest := PairRequest{}
	json.Unmarshal(request, &pairRequest)
	root.logger.Printf("pairing: %s", pairRequest.PublicKey)
	privKey1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubKey1 := elliptic.Marshal(elliptic.P256(), privKey1.PublicKey.X, privKey1.PublicKey.Y)
	pubKey2bytes, _ := base64.StdEncoding.DecodeString(pairRequest.PublicKey)
	pubKey2 := new(ecdsa.PublicKey)
	pubKey2.X, pubKey2.Y = elliptic.Unmarshal(elliptic.P256(), pubKey2bytes)
	sharedKey, _ := elliptic.P256().ScalarMult(pubKey2.X, pubKey2.Y, privKey1.D.Bytes())
	bytes := sharedKey.Bytes()
	root.cache[pairRequest.PublicKey] = Cache{
		SharedKey: bytes,
	}
	pairResponse := PairResponse{
		PublicKey: base64.StdEncoding.EncodeToString(pubKey1),
	}
	body, _ := json.Marshal(&pairResponse)
	topic := strings.Replace(pairRequest.PublicKey, "/", "", -1)
	topic = strings.Replace(topic, "+", "", -1)
	topic = strings.Replace(topic, "=", "", -1)
	topic = fmt.Sprintf("%s/%s/pair", root.Identifier, topic)
	root.mqttClientWan.Publish(topic, 0, false, body)
}

func (root *GW) recv(_ mqtt.Client, message mqtt.Message) {
	request := message.Payload()
	cipherRequest := CipherRequest{}
	json.Unmarshal(request, &cipherRequest)
	codec, contains := root.cache[cipherRequest.PublicKey]
	if contains == false {
		publicKey := strings.Replace(cipherRequest.PublicKey, "/", "", -1)
		publicKey = strings.Replace(publicKey, "+", "", -1)
		publicKey = strings.Replace(publicKey, "=", "", -1)
		topic := fmt.Sprintf("%s/%s/repair", root.Identifier, publicKey)
		root.mqttClientWan.Publish(topic, 0, false, []byte{})
		return
	}
	plainRequest := PlainRequest{}
	root.codec(&codec)
	root.decrypt(&codec, &cipherRequest, &plainRequest)
	plainRequest.PublicKey = cipherRequest.PublicKey
	plainRequest.ResponseTopic = root.ResponseTopic
	fmt.Printf("Bus Raw Received For: %s\n", plainRequest.Topic)
	root.com.Publish(plainRequest.Topic, plainRequest)
}

func (root *GW) codec(cache *Cache) {
	if cache.Block == nil {
		cache.Block, _ = aes.NewCipher(cache.SharedKey)
	}
	if cache.GCM == nil {
		cache.GCM, _ = cipher.NewGCM(cache.Block)
	}
}

func (root *GW) decrypt(codec *Cache, cipherRequest *CipherRequest, plainRequest *PlainRequest) {
	cipher, _ := base64.StdEncoding.DecodeString(cipherRequest.Cipher)
	iv, _ := base64.StdEncoding.DecodeString(cipherRequest.IV)
	plain, _ := codec.GCM.Open(nil, iv, cipher, nil)
	gz, _ := gzip.NewReader(bytes.NewReader(plain))
	uncompressed, _ := io.ReadAll(gz)
	json.Unmarshal(uncompressed, &plainRequest)
}

func (root *GW) encrypt(codec *Cache, plainResponse *PlainResponse, cipherResponse *CipherResponse) {
	json, _ := json.Marshal(plainResponse)
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	gz.Write(json)
	gz.Flush()
	gz.Close()
	iv := make([]byte, codec.GCM.NonceSize())
	rand.Read(iv)
	cipher := codec.GCM.Seal(nil, iv, buffer.Bytes(), nil)
	cipherResponse.Cipher = base64.StdEncoding.EncodeToString(cipher)
	cipherResponse.IV = base64.StdEncoding.EncodeToString(iv)
}

type GW struct {
	cache         map[string]Cache
	logger        *log.Logger
	mqttClientWan mqtt.Client
	Identifier    string
	ResponseTopic string
	com           *com.COM
}

type Cache struct {
	SharedKey []byte
	Block     cipher.Block
	GCM       cipher.AEAD
	Local     bool
}

type PairRequest struct {
	PublicKey string `json:"pk"`
}

type PairResponse struct {
	PublicKey string `json:"pk"`
}

type ResultResponse struct {
	EventName string `json:"eventName"`
	Result    any    `json:"result"`
}

type CipherRequest struct {
	PublicKey string `json:"pk"`
	IV        string `json:"iv"`
	Cipher    string `json:"cipher"`
}

type PlainRequest struct {
	PublicKey     string          `json:"pk,omitempty"`
	CorrelationId string          `json:"cid,omitempty"`
	Topic         string          `json:"topic,omitempty"`
	Plain         json.RawMessage `json:"plain,omitempty"`
	ResponseTopic string          `json:"responseTopic,omitempty"`
}

type PlainResponse struct {
	PublicKey     string          `json:"pk"`
	CorrelationId string          `json:"cid"`
	Plain         json.RawMessage `json:"plain"`
}

type BroadcastResponse struct {
	Plain json.RawMessage `json:"plain"`
}

type CipherResponse struct {
	Cipher string `json:"cipher"`
	IV     string `json:"iv"`
}

type PingRequest struct {
	Ping string `json:"ping"`
}

type PongResponse struct {
	Pong string `json:"pong"`
}
