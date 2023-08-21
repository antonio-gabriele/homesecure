package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"gadu/bee/m/v2/zstack"
	"gadu/shared/m/v2/com"
	"gadu/shared/m/v2/infra"
	"github.com/glebarez/sqlite"
	"github.com/shimmeringbee/bytecodec"
	"github.com/shimmeringbee/bytecodec/bitbuffer"
	"github.com/shimmeringbee/zcl/commands/local/onoff"

	//"github.com/shimmeringbee/zcl/commands/local/onoff"
	"github.com/shimmeringbee/zcl/communicator"
	"go.bug.st/serial"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"

	"github.com/shimmeringbee/zcl"
	"github.com/shimmeringbee/zcl/commands/global"
	"github.com/shimmeringbee/zigbee"
)

type BeeZclCommand struct {
	Channel string `json:"channel"`
	Cmd     uint8
	Payload []uint8
}
type ZigbeeTiCc25xx struct {
	logger    *log.Logger
	com       *com.COM
	database  *gorm.DB
	nodeTable *zstack.NodeTable
	zStack    *zstack.ZStack
	*communicator.Communicator
	*zcl.CommandRegistry
}

const MyEndpointId = 1

var removed = map[string]bool{}
var TransactionId = uint8(0)

func NewTiZigbee(identifier string, logger *log.Logger) {
	root := ZigbeeTiCc25xx{}
	root.database, _ = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	root.com = com.NewCOM(identifier)
	//infra.NewChannels(root.com, root.database, logger)
	root.logger = logger
	root.logger.Println("Starting...")
	migrate(root.database)
	root.com.Subscribe("bee/readBindingTable", root.readBindingTable)
	root.com.Subscribe("bee/explore", root.explore)
	root.com.Subscribe("bee/permitJoin", root.permitJoin)
	root.com.Subscribe("channels/channels", root.channels)
	root.com.Subscribe("channels/channels/items", root.channelsItems)
	root.com.Subscribe("channels/channels/wr", root.channelsSave)
	root.com.Subscribe("channels/behaviours/wr", root.behavioursSave)
	root.com.Subscribe("channels/commands/execute", root.commandExecute)
	root.init()
}

func (root *ZigbeeTiCc25xx) behavioursSave(rawMessage json.RawMessage) {
	plainRequest := com.PlainRequest{}
	json.Unmarshal(rawMessage, &plainRequest)
	root.logger.Println("Saving...")
	var beeNodeEndpointCluster BeeNodeEndpointCluster
	json.Unmarshal(plainRequest.Plain, &beeNodeEndpointCluster)
	root.database.Where(&BeeNodeEndpointCluster{IEEEAddress: beeNodeEndpointCluster.IEEEAddress, EndpointId: beeNodeEndpointCluster.EndpointId, ClusterId: beeNodeEndpointCluster.ClusterId}).
		Select("enabled").
		Updates(&beeNodeEndpointCluster)
	root.com.Reply(&plainRequest, true)
	//root.runtime.Bus.Publish("behaviour/configuration/execute", logicBehaviours)
}

func (root *ZigbeeTiCc25xx) channelsSave(rawMessage json.RawMessage) {
	plainRequest := com.PlainRequest{}
	json.Unmarshal(rawMessage, &plainRequest)
	root.logger.Println("Saving...")
	var beeNodeEndpoint BeeNodeEndpoint
	json.Unmarshal(plainRequest.Plain, &beeNodeEndpoint)
	root.database.Where(&BeeNodeEndpoint{IEEEAddress: beeNodeEndpoint.IEEEAddress, EndpointId: beeNodeEndpoint.EndpointId}).
		Select("enabled", "room", "name").
		Updates(&beeNodeEndpoint)
	root.com.Reply(&plainRequest, true)
}

func (root *ZigbeeTiCc25xx) readBindingTable(rawMessage json.RawMessage) {
	plainRequest := com.PlainRequest{}
	json.Unmarshal(rawMessage, &plainRequest)
	beeNodeEndpoint := BeeNodeEndpoint{}
	json.Unmarshal(plainRequest.Plain, &beeNodeEndpoint)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ieeeAddress := zigbee.IEEEAddress(HexToUInt64("7cb03eaa0a0ac4af"))
	root.zStack.ReadBindingTable(ctx, ieeeAddress)
}

func (root *ZigbeeTiCc25xx) explore(rawMessage json.RawMessage) {
	root.exploreBee()
}

func (root *ZigbeeTiCc25xx) permitJoin(rawMessage json.RawMessage) {
	plainRequest := com.PlainRequest{}
	json.Unmarshal(rawMessage, &plainRequest)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := errors.New("BUS Zigbee non connesso")
	if root.zStack != nil {
		err = root.zStack.PermitJoin(ctx, true)
	}
	root.com.Reply(&plainRequest, err == nil)
}

func (root *ZigbeeTiCc25xx) commandExecute(rawMessage json.RawMessage) {
	plainRequest := com.PlainRequest{}
	json.Unmarshal(rawMessage, &plainRequest)
	beeCommandTransferObject := BeeCommandTransferObject{}
	json.Unmarshal(plainRequest.Plain, &beeCommandTransferObject)
	beeCommand := BeeCommand{}
	result := root.database.Where(&BeeCommand{
		ClusterId: beeCommandTransferObject.Behaviour.ClusterId,
		Command:   beeCommandTransferObject.Command,
	}).First(&beeCommand)
	if result.RowsAffected == 0 {
		return
	}
	TransactionId++
	bb := bitbuffer.NewBitBuffer()
	header := zcl.Header{
		Control: zcl.Control{
			Reserved:               0,
			DisableDefaultResponse: false,
			Direction:              zcl.ClientToServer,
			ManufacturerSpecific:   beeCommand.CommandMfccode > 0,
			FrameType:              zcl.FrameLocal,
		},
		Manufacturer:        zigbee.ManufacturerCode(beeCommand.CommandMfccode),
		TransactionSequence: TransactionId,
		CommandIdentifier:   zcl.CommandIdentifier(beeCommand.CommandId),
	}
	TransactionId = (TransactionId + 1) % 0xFF
	bytecodec.MarshalToBitBuffer(bb, header)
	appMessage := zigbee.ApplicationMessage{
		ClusterID:           zigbee.ClusterID(beeCommand.ClusterId),
		SourceEndpoint:      MyEndpointId,
		DestinationEndpoint: zigbee.Endpoint(beeCommandTransferObject.Behaviour.EndpointId),
		Data:                bb.Bytes(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	ieee := HexToUInt64(beeCommandTransferObject.Behaviour.IEEEAddress)
	err := root.zStack.SendApplicationMessageToNode(ctx, zigbee.IEEEAddress(ieee), appMessage, true)
	root.com.Reply(&plainRequest, err == nil)
}

func (root *ZigbeeTiCc25xx) init() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	netCfg, _ := zigbee.GenerateNetworkConfiguration()
	netCfg.PANID = 0x1234
	netCfg.Channel = 11
	netCfg.ExtendedPANID = 0x1234
	netCfg.NetworkKey = zigbee.NetworkKey{0x5a, 0x69, 0x67, 0x42, 0x65, 0x65, 0x41, 0x6c, 0x6c, 0x69, 0x61, 0x6e, 0x63, 0x65, 0x30, 0x39}
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	serialPort, err := serial.Open("COM10", mode)
	if err != nil {
		//root.com.Broadcast("bee/init", false)
		return
	}
	root.checkNodeTable()
	root.zStack = zstack.New(serialPort, root.nodeTable)
	root.CommandRegistry = zcl.NewCommandRegistry()
	global.Register(root.CommandRegistry)
	onoff.Register(root.CommandRegistry)
	root.Communicator = communicator.NewCommunicator(root.zStack, root.CommandRegistry)
	root.zStack.Initialise(ctx, netCfg)
	err = root.zStack.RegisterAdapterEndpoint(ctx, MyEndpointId, 0x0104, 0x0001, 0x0001, []zigbee.ClusterID{0x1, 0x3, 0x6, 0x402, 0x405}, []zigbee.ClusterID{0x1, 0x3, 0x6, 0x402, 0x405})
	if err != nil {
		fmt.Printf("Err: %s\n", err.Error())
		return
	}
	go root.events()
}

func (root *ZigbeeTiCc25xx) events() {
	for {
		ctx := context.Background()
		event, err := root.zStack.ReadEvent(ctx)
		if err != nil {
			return
		}

		switch e := event.(type) {
		case zigbee.NodeJoinEvent:
			log.Printf("+ %s", e.Node.IEEEAddress.String())
			removed[e.IEEEAddress.String()] = false
			root.createNodeTable(e.Node)
			ieeeAddress := e.IEEEAddress.String()
			root.exploreBeeNodeByIeee(ieeeAddress, true)
		case zigbee.NodeLeaveEvent:
			log.Printf("- %s", e.Node.IEEEAddress.String())
			removed[e.IEEEAddress.String()] = true
			root.deleteNodeTable(e.Node)
		case zigbee.NodeUpdateEvent:
			log.Printf("u %s", e.Node.IEEEAddress.String())
			if value, found := removed[e.IEEEAddress.String()]; found && value {
				return
			}
			root.updateNodeTable(e.Node)
			ieeeAddress := e.IEEEAddress.String()
			root.exploreBeeNodeByIeee(ieeeAddress, false)
		case zigbee.NodeIncomingMessageEvent:
			log.Printf("m %s", e.Node.IEEEAddress.String())
			root.Communicator.ProcessIncomingMessage(e)
			applicationMessage, err := root.CommandRegistry.Unmarshal(e.ApplicationMessage)
			if err == nil && applicationMessage.FrameType == zcl.FrameGlobal && applicationMessage.CommandIdentifier == global.ReportAttributesID {
				root.reportAttribute(e, applicationMessage)
			}
			if err == nil && applicationMessage.FrameType == zcl.FrameLocal {
				root.command(e, applicationMessage)
			}
			ieeeAddress := e.IEEEAddress.String()
			root.exploreBeeNodeByIeee(ieeeAddress, false)
		}
	}
}

func (root *ZigbeeTiCc25xx) reportAttribute(event zigbee.NodeIncomingMessageEvent, message zcl.Message) {
	reportAttributes := message.Command.(*global.ReportAttributes)
	for _, reportAttribute := range reportAttributes.Records {
		beeAttribute := BeeAttribute{
			AttributeId:     uint16(reportAttribute.Identifier),
			AttributeMfcode: uint16(message.Manufacturer),
			ClusterId:       uint16(message.ClusterID),
		}
		beeNodeEndpointCluster := BeeNodeEndpointCluster{
			EndpointId:  uint16(message.SourceEndpoint),
			ClusterId:   uint16(message.ClusterID),
			IEEEAddress: event.IEEEAddress.String(),
		}
		root.database.First(&beeAttribute, &beeAttribute)
		root.database.First(&beeNodeEndpointCluster, &beeNodeEndpointCluster)
		bytes, _ := json.Marshal(reportAttribute.DataTypeValue.Value)
		logicProperty := infra.LogicProperty{
			Channel:   beeNodeEndpointCluster.Channel,
			Behaviour: beeNodeEndpointCluster.Behaviour,
			Property:  beeAttribute.Property,
			UpdateAt:  time.Now(),
			Value:     string(bytes),
		}
		log.Printf("ra: %s\\%s\\%s", logicProperty.Channel, logicProperty.Property, logicProperty.Value)
		root.com.Reply(&com.PlainRequest{}, com.ResultResponse{
			EventName: "channels/properties/status",
			Result:    logicProperty,
		})
		payload, _ := json.Marshal(&logicProperty)
		root.com.MqttClientInt.Publish("channels/properties/status", 0, false, payload)
		root.database.Clauses(clause.OnConflict{UpdateAll: true}).Create(&logicProperty)
	}
}

func (root *ZigbeeTiCc25xx) command(event zigbee.NodeIncomingMessageEvent, message zcl.Message) {
	beeCommand := BeeCommand{}
	root.database.First(&beeCommand, map[string]any{"cluster_id": event.ApplicationMessage.ClusterID, "command_id": event.ApplicationMessage.Data[2]})
	beeNodeEndpointCluster := BeeNodeEndpointCluster{}
	root.database.First(&beeNodeEndpointCluster, map[string]any{"endpoint_id": event.ApplicationMessage.SourceEndpoint, "cluster_id": event.ApplicationMessage.ClusterID, "ieee_address": event.IEEEAddress.String()})
	logicCommand := infra.LogicCommand{
		Channel:   beeNodeEndpointCluster.Channel,
		Behaviour: beeNodeEndpointCluster.Behaviour,
		Command:   beeCommand.CommandName,
	}
	resultResponse := com.ResultResponse{
		EventName: "channels/commands/fire",
		Result:    logicCommand,
	}
	root.com.Reply(&com.PlainRequest{}, &resultResponse)

}

func (root *ZigbeeTiCc25xx) checkNodeTable() {
	root.nodeTable = zstack.NewNodeTable()
	var srcs []BeeNode
	var dsts []zigbee.Node
	root.database.Find(&srcs)
	for _, src := range srcs {
		dsts = append(dsts, zigbee.Node{
			IEEEAddress:    zigbee.IEEEAddress(HexToUInt64(src.IEEEAddress)),
			NetworkAddress: zigbee.NetworkAddress(src.NetworkAddress),
			LogicalType:    zigbee.LogicalType(src.LogicalType),
			LQI:            src.LQI,
			Depth:          src.Depth,
			LastDiscovered: src.LastDiscovered,
			LastReceived:   src.LastReceived,
		})
	}
	root.nodeTable.Load(dsts)
}

func (root *ZigbeeTiCc25xx) createNodeTable(src zigbee.Node) {
	beeNode := BeeNode{
		IEEEAddress:    src.IEEEAddress.String(),
		NetworkAddress: uint16(src.NetworkAddress),
		LogicalType:    uint8(src.LogicalType),
		LQI:            src.LQI,
		Depth:          src.Depth,
		LastDiscovered: src.LastDiscovered,
		LastReceived:   src.LastReceived,
		Explored:       false,
	}
	root.database.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&beeNode)
	//root.database.Where(&BeeNode{IEEEAddress: src.IEEEAddress.String()}).Update("explored", false)
	root.database.Where(&BeeNodeEndpoint{IEEEAddress: src.IEEEAddress.String()}).Update("explored", false)
	root.database.Where(&BeeNodeEndpointCluster{IEEEAddress: src.IEEEAddress.String()}).Update("explored", false)
}

func (root *ZigbeeTiCc25xx) updateNodeTable(src zigbee.Node) {
	beeNode := BeeNode{
		IEEEAddress:    src.IEEEAddress.String(),
		NetworkAddress: uint16(src.NetworkAddress),
		LogicalType:    uint8(src.LogicalType),
		LQI:            src.LQI,
		Depth:          src.Depth,
		LastDiscovered: src.LastDiscovered,
		LastReceived:   src.LastReceived,
	}
	root.database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ieee_address"}},
		DoUpdates: clause.AssignmentColumns([]string{"network_address", "logical_type", "lqi", "depth", "last_discovered", "last_received"}),
	}).Create(&beeNode)
}

func (root *ZigbeeTiCc25xx) deleteNodeTable(src zigbee.Node) {
	dst := BeeNode{
		IEEEAddress: src.IEEEAddress.String(),
	}
	root.database.Delete(&dst)
}

func (root *ZigbeeTiCc25xx) channels(rawMessage json.RawMessage) {
	plainRequest := com.PlainRequest{}
	json.Unmarshal(rawMessage, &plainRequest)
	beeNodeEndpoint := BeeNodeEndpoint{}
	json.Unmarshal(plainRequest.Plain, &beeNodeEndpoint)
	var result []BeeNodeEndpoint
	root.database.
		Model(&BeeNodeEndpoint{}).
		Where(&BeeNodeEndpoint{Enabled: true}).
		Preload("Behaviours", "enabled = ?", true).
		Preload("Behaviours.Properties").
		Find(&result)
	root.com.Reply(&plainRequest, result)
}

func (root *ZigbeeTiCc25xx) channelsItems(rawMessage json.RawMessage) {
	plainRequest := com.PlainRequest{}
	json.Unmarshal(rawMessage, &plainRequest)
	beeNodeEndpoint := BeeNodeEndpoint{}
	json.Unmarshal(plainRequest.Plain, &beeNodeEndpoint)
	var result []BeeNodeEndpoint
	root.database.
		Model(&BeeNodeEndpoint{}).
		Preload("Behaviours").
		Find(&result)
	root.com.Reply(&plainRequest, result)
}
