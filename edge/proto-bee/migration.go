package main

import (
	_ "embed"
	"fmt"
	"gadu/shared/m/v2/infra"
	"github.com/antchfx/xmlquery"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

//go:embed models.xml
var models string

func applyAttributeValue(db *gorm.DB, client bool, beeAttribute *BeeAttribute, attributeValue *xmlquery.Node) {
	value := attributeValue.SelectAttr("value")
	if strings.HasPrefix(value, "0x") == false {
		value = fmt.Sprintf("0x%s", value)
	}
	beeAttributeValue := BeeAttributeValue{
		ClusterId:           beeAttribute.ClusterId,
		ClusterClient:       client,
		AttributeId:         beeAttribute.AttributeId,
		AttributeMfcode:     beeAttribute.AttributeMfcode,
		AttributeValueName:  attributeValue.SelectAttr("name"),
		AttributeValueValue: HexToUInt64(attributeValue.SelectAttr("value")),
	}
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&beeAttributeValue)
}

func applyAttribute(db *gorm.DB, beeCluster *BeeCluster, client bool, attribute *xmlquery.Node) {
	beeAttribute := BeeAttribute{
		ClusterId:         beeCluster.ClusterId,
		ClusterClient:     client,
		AttributeId:       HexToUInt16(attribute.SelectAttr("id")),
		AttributeName:     attribute.SelectAttr("name"),
		AttributeType:     attribute.SelectAttr("type"),
		AttributeAccess:   attribute.SelectAttr("access"),
		AttributeDefault:  attribute.SelectAttr("default"),
		AttributeRequired: attribute.SelectAttr("required") == "m",
		AttributeShowas:   attribute.SelectAttr("showas"),
		AttributeMfcode:   HexToUInt16(attribute.SelectAttr("mfcode")),
	}
	if minInterval, err := strconv.Atoi(attribute.SelectAttr("minInterval")); err == nil {
		beeAttribute.MinInterval = uint16(minInterval)
	}
	if maxInterval, err := strconv.Atoi(attribute.SelectAttr("maxInterval")); err == nil {
		beeAttribute.MaxInterval = uint16(maxInterval)
	}
	if reportableChange, err := strconv.Atoi(attribute.SelectAttr("reportableChange")); err == nil {
		beeAttribute.ReportableChange = uint16(reportableChange)
	}
	rangee := attribute.SelectAttr("range")
	if rangee != "" {
		tokens := strings.Split(rangee, ",")
		if strings.HasPrefix(tokens[0], "0x") {
			beeAttribute.AttributeRange0 = HexToUInt64(tokens[0])
		} else {
			beeAttribute.AttributeRange0, _ = strconv.ParseUint(tokens[0], 10, 64)
		}
		if len(tokens) > 1 && strings.HasPrefix(tokens[1], "0x") {
			beeAttribute.AttributeRange1 = HexToUInt64(tokens[1])
		} else if len(tokens) > 1 {
			beeAttribute.AttributeRange1, _ = strconv.ParseUint(tokens[1], 10, 64)
		}
	}
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&beeAttribute)
	attributeValues := attribute.SelectElements("/value")
	for _, attributeValue := range attributeValues {
		applyAttributeValue(db, client, &beeAttribute, attributeValue)
	}
}

func applyCommandAttribute(db *gorm.DB, beeCluster *BeeCluster, client bool, beeCommand *BeeCommand, attribute *xmlquery.Node) {
	beeAttribute := BeeCommandAttribute{
		ClusterId: beeCluster.ClusterId,
		//ClusterMfCode:            beeCluster.ClusterMfCode,
		ClusterClient:            client,
		CommandAttributeId:       HexToUInt16(attribute.SelectAttr("id")),
		CommandAttributeName:     attribute.SelectAttr("name"),
		CommandAttributeType:     attribute.SelectAttr("type"),
		CommandAttributeDefault:  HexToUInt16(attribute.SelectAttr("default")),
		CommandAttributeRequired: attribute.SelectAttr("required") == "m",
		CommandAttributeShowas:   attribute.SelectAttr("showas"),
		CommandId:                beeCommand.CommandId,
	}
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&beeAttribute)
}

func applyCommand(db *gorm.DB, beeCluster *BeeCluster, client bool, command *xmlquery.Node) {
	beeCommand := BeeCommand{
		ClusterId: beeCluster.ClusterId,
		//ClusterMfCode:    beeCluster.ClusterMfCode,
		ClusterClient:    client,
		CommandId:        HexToUInt8(command.SelectAttr("id")),
		CommandMfccode:   HexToUInt16(command.SelectAttr("vendor")),
		CommandName:      command.SelectAttr("name"),
		CommandRequired:  command.SelectAttr("required") == "m",
		CommandDirection: command.SelectAttr("dir"),
	}
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&beeCommand)
	attributes := command.SelectElements("/payload/attribute")
	for _, attribute := range attributes {
		applyCommandAttribute(db, beeCluster, client, &beeCommand, attribute)
	}
}

func applyCluster(db *gorm.DB, cluster *xmlquery.Node) {
	beeCluster := BeeCluster{
		ClusterId:   HexToUInt16(cluster.SelectAttr("id")),
		ClusterName: cluster.SelectAttr("name"),
		//ClusterMfCode: utils.HexToUInt16(cluster.SelectAttr("mfcode")),
	}
	description := cluster.SelectElement("/description")
	if description != nil {
		beeCluster.ClusterDescription = description.InnerText()
	}
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&beeCluster)
	attributes := cluster.SelectElements("/server/attribute")
	for _, attribute := range attributes {
		applyAttribute(db, &beeCluster, false, attribute)
	}
	commands := cluster.SelectElements("/server/command")
	for _, command := range commands {
		applyCommand(db, &beeCluster, false, command)
	}
	attributes = cluster.SelectElements("/client/attribute")
	for _, attribute := range attributes {
		applyAttribute(db, &beeCluster, true, attribute)
	}
}

func migrate(db *gorm.DB) {
	db.AutoMigrate(&infra.LogicProperty{})
	db.AutoMigrate(&BeeCluster{})
	db.AutoMigrate(&BeeCommand{})
	db.AutoMigrate(&BeeCommandAttribute{})
	db.AutoMigrate(&BeeAttribute{})
	db.AutoMigrate(&BeeAttributeValue{})
	db.AutoMigrate(&BeeNode{})
	db.AutoMigrate(&BeeNodeEndpoint{})
	db.AutoMigrate(&BeeNodeEndpointCluster{})
	xml, _ := xmlquery.Parse(strings.NewReader(models))
	xmlquery.FindEach(xml, "//cluster", func(i int, cluster *xmlquery.Node) {
		applyCluster(db, cluster)
	})
}
