package main

import (
	"encoding/json"
	"gadu/shared/m/v2/com"
	"gadu/shared/m/v2/infra"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type WORKFLOW struct {
	logger        *log.Logger
	com           *com.COM
	dataCtx       ast.IDataContext
	knowledgeBase *ast.KnowledgeBase
	engine        *engine.GruleEngine
}

func main() {
	logger := log.New(os.Stderr, "GW: ", log.LstdFlags)
	logger.Println("Starting...")
	NewWORKFLOW("a1b2c3", logger)
	logger.Println("Started...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logger.Println("Ending...")
}

const (
	GRL = `
rule CallingLog "Calling a log" {
	when
		true
	then
		Log("Hello Grule");
}
`
)

func NewWORKFLOW(identifier string, logger *log.Logger) {
	root := WORKFLOW{}
	root.dataCtx = ast.NewDataContext()
	root.logger = logger
	root.logger.Println("Starting...")
	root.com = com.NewCOM(identifier)
	knowledgeLibrary := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)

	drls := `
rule CheckValues "Check the default values" {
    when 
        channelPropertyStatus.Channel == "5c0272fffe3d88a1:1"
    then
		Log("Hello Grule");
		Retract("CheckValues");
}
`
	byteArr := pkg.NewBytesResource([]byte(drls))
	ruleBuilder.BuildRuleFromResource("Tutorial", "0.0.1", byteArr)
	root.knowledgeBase = knowledgeLibrary.NewKnowledgeBaseInstance("Tutorial", "0.0.1")
	root.engine = engine.NewGruleEngine()
	root.com.Subscribe("channels/properties/status", root.channelPropertyStatus)
}

func (root *WORKFLOW) channelPropertyStatus(rawMessage json.RawMessage) {
	logicProperty := infra.LogicProperty{}
	json.Unmarshal(rawMessage, &logicProperty)
	root.dataCtx.Add("channelPropertyStatus", &logicProperty)
	root.engine.Execute(root.dataCtx, root.knowledgeBase)

}
