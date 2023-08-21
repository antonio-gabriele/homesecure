package main

import (
	"context"
	"fmt"
	"github.com/shimmeringbee/zcl"
	"github.com/shimmeringbee/zcl/commands/local/basic"
	"github.com/shimmeringbee/zigbee"
	"gorm.io/gorm/clause"
	"time"
)

func (root *ZigbeeTiCc25xx) provisioning(beeNodeEndpointCluster *BeeNodeEndpointCluster) bool {
	fmt.Printf("provisioning %s\\%d\\%d\n", beeNodeEndpointCluster.IEEEAddress, beeNodeEndpointCluster.EndpointId, beeNodeEndpointCluster.ClusterId)
	var beeAttributes []BeeAttribute
	result := root.database.Where("cluster_id = ? AND min_interval > 0 AND max_interval > 0 AND reportable_change > 0", beeNodeEndpointCluster.ClusterId).Find(&beeAttributes)
	ieeeAddress := zigbee.IEEEAddress(HexToUInt64(beeNodeEndpointCluster.IEEEAddress))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if result.RowsAffected > 0 {
		fmt.Printf("binding %s\\%d\\%d\n", beeNodeEndpointCluster.IEEEAddress, beeNodeEndpointCluster.EndpointId, beeNodeEndpointCluster.ClusterId)
		err := root.zStack.BindNodeToController(ctx, ieeeAddress, zigbee.Endpoint(beeNodeEndpointCluster.EndpointId), MyEndpointId, zigbee.ClusterID(beeNodeEndpointCluster.ClusterId))
		if err != nil {
			return false
		}
	}
	for _, beeAttribute := range beeAttributes {
		fmt.Printf("configure reporting %s\\%d\\%d\\%d\n", beeNodeEndpointCluster.IEEEAddress, beeNodeEndpointCluster.EndpointId, beeNodeEndpointCluster.ClusterId, beeAttribute.AttributeId)
		beeAttributeType := 0x10
		switch beeAttribute.AttributeType {
		case "u8":
			beeAttributeType = 0x20
		case "s16":
			beeAttributeType = 0x29
		}
		TransactionId = TransactionId + 1
		err := root.Communicator.Global().ConfigureReporting(ctx, ieeeAddress, true, zigbee.ClusterID(beeNodeEndpointCluster.ClusterId), 0, zigbee.Endpoint(beeNodeEndpointCluster.EndpointId), zigbee.Endpoint(MyEndpointId), TransactionId, zcl.AttributeID(beeAttribute.AttributeId), zcl.AttributeDataType(beeAttributeType), beeAttribute.MinInterval, beeAttribute.MaxInterval, beeAttribute.ReportableChange)
		if err != nil {
			fmt.Printf(err.Error())
			return false
		}
	}
	return true
}

func (root *ZigbeeTiCc25xx) exploreBeeNodeEndpoint(beeNodeEndpoint *BeeNodeEndpoint, force bool) {
	root.logger.Printf("exploreBeeNodeEndpoint %s\\%d", beeNodeEndpoint.IEEEAddress, beeNodeEndpoint.EndpointId)
	beeNode := BeeNode{IEEEAddress: beeNodeEndpoint.IEEEAddress}
	root.database.First(&beeNode, &beeNode)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ieee := zigbee.IEEEAddress(HexToUInt64(beeNodeEndpoint.IEEEAddress))
	endpointDescription, _ := root.zStack.QueryNodeEndpointDescription(ctx, ieee, zigbee.Endpoint(beeNodeEndpoint.EndpointId))
	root.deviceName(beeNodeEndpoint)
	beeNodeEndpoint.ProfileId = uint16(endpointDescription.ProfileID)
	beeNodeEndpoint.DeviceId = endpointDescription.DeviceID
	beeNodeEndpoint.DeviceVersion = endpointDescription.DeviceVersion
	explored := true
	for _, value := range endpointDescription.InClusterList {
		beeCluster := BeeCluster{}
		result := root.database.First(&beeCluster, map[string]any{"cluster_id": uint16(value)})
		if result.RowsAffected == 0 {
			continue
		}
		obj := BeeNodeEndpointCluster{IEEEAddress: beeNodeEndpoint.IEEEAddress, EndpointId: beeNodeEndpoint.EndpointId, ClusterId: uint16(value), ClusterClient: false, Behaviour: beeCluster.ClusterName}
		root.database.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&obj)
		explored1 := root.provisioning(&obj)
		explored = explored && explored1
	}
	for _, value := range endpointDescription.OutClusterList {
		beeCluster := BeeCluster{}
		result := root.database.First(&beeCluster, map[string]any{"cluster_id": uint16(value)})
		if result.RowsAffected == 0 {
			continue
		}
		obj := BeeNodeEndpointCluster{IEEEAddress: beeNodeEndpoint.IEEEAddress, EndpointId: beeNodeEndpoint.EndpointId, ClusterId: uint16(value), ClusterClient: true, Behaviour: beeCluster.ClusterName}
		root.database.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&obj)
		explored1 := root.provisioning(&obj)
		explored = explored && explored1
	}
	beeNodeEndpoint.Explored = explored
	root.database.Updates(&beeNodeEndpoint)
}

func (root *ZigbeeTiCc25xx) exploreBeeNode(beeNode *BeeNode) {
	root.logger.Printf("exploreBeeNode %s", beeNode.IEEEAddress)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ieee := zigbee.IEEEAddress(HexToUInt64(beeNode.IEEEAddress))
	endpoints, _ := root.zStack.QueryNodeEndpoints(ctx, ieee)
	for _, endpoint := range endpoints {
		dst := BeeNodeEndpoint{
			IEEEAddress: beeNode.IEEEAddress,
			EndpointId:  uint16(endpoint),
		}
		root.database.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&dst)
	}
	beeNode.Explored = true
	root.database.Updates(&beeNode)
}

func (root *ZigbeeTiCc25xx) exploreBeeNodeByIeee(ieee string, force bool) {
	beeNode := BeeNode{IEEEAddress: ieee}
	root.database.First(&beeNode, &beeNode)
	if beeNode.Explored && force == false {
		return
	}
	root.exploreBeeNode(&beeNode)
	var beeNodeEndpoints []BeeNodeEndpoint
	root.database.Find(&beeNodeEndpoints, &BeeNodeEndpoint{IEEEAddress: ieee})
	for _, beeNodeEndpoint := range beeNodeEndpoints {
		root.exploreBeeNodeEndpoint(&beeNodeEndpoint, force)
	}
}

func (root *ZigbeeTiCc25xx) exploreBee() {
	root.logger.Printf("scan")
	var beeNodes []BeeNode
	root.database.Debug().Find(&beeNodes, map[string]interface{}{"explored": 0})
	for _, beeNode := range beeNodes {
		root.exploreBeeNode(&beeNode)
	}
	var beeNodeEndpoints []BeeNodeEndpoint
	root.database.Find(&beeNodeEndpoints, map[string]interface{}{"explored": 0})
	for _, beeNodeEndpoint := range beeNodeEndpoints {
		root.exploreBeeNodeEndpoint(&beeNodeEndpoint, false)
	}
}

func (root *ZigbeeTiCc25xx) deviceName(beeNodeEndpoint *BeeNodeEndpoint) {
	TransactionId++
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ieee := HexToUInt64(beeNodeEndpoint.IEEEAddress)
	attributes, err := root.Communicator.Global().ReadAttributes(ctx, zigbee.IEEEAddress(ieee), true, zcl.BasicId, zigbee.NoManufacturer, MyEndpointId, zigbee.Endpoint(beeNodeEndpoint.EndpointId), TransactionId, []zcl.AttributeID{basic.ManufacturerName, basic.ModelIdentifier})
	if err == nil {
		beeNodeEndpoint.ManufacturerName = attributes[0].DataTypeValue.Value.(string)
		beeNodeEndpoint.ModelIdentifier = attributes[1].DataTypeValue.Value.(string)
	}
}
