package mpc_node_test

import (
	"os"
	"testing"
	"time"

	"github.com/krakenh2020/MPCService/manager"
	"github.com/krakenh2020/MPCService/mpc_node"
)

func TestRunNode(t *testing.T) {
	go manager.RunManager(5007, 5008, "../manager/assets", "info", "../logging/log.log",
		"../key_management/keys_certificates")
	time.Sleep(1 * time.Second)

	// run servers
	nodeNames := []string{"Berlin_node", "Paris_node", "Ljubljana_node", "Rome_node", "Leuven_node",
		"Bristol_node", "Madrid_node"}
	for nodeId := 0; nodeId < len(nodeNames); nodeId++ {
		go mpc_node.RunNode(nodeNames[nodeId], "localhost", 5040+nodeId, "../key_management/keys_certificates",
			os.Getenv("SCALE_MAMBA_PATH"), "info", "../logging/log.log",
			"localhost:5008", "some description")
	}
	time.Sleep(1 * time.Second)
}
