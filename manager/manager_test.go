package manager_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/krakenh2020/MPCService/data_provider"
	"github.com/krakenh2020/MPCService/mpc_node"

	"github.com/krakenh2020/MPCService/data_management"
	"github.com/krakenh2020/MPCService/manager"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/krakenh2020/MPCService/key_management"
)

func TestRequestComputationWithManager(t *testing.T) {
	go manager.RunManager(5007, 5008, "../manager/assets", "info", "../logging/log.log",
		"../key_management/keys_certificates")
	time.Sleep(1 * time.Second)

	nodeNames := []string{"Berlin_node", "Paris_node", "Ljubljana_node", "Rome_node", "Leuven_node",
		"Bristol_node", "Madrid_node"}

	trustedNodes := []string{"Berlin_node", "Paris_node", "Ljubljana_node", "Rome_node"}
	// run data provider
	go data_provider.RunDatasetProvider("Data_provider1", "../data_provider/datasets", "debug",
		"../logging/log.log", "localhost:5008", "../key_management/keys_certificates", trustedNodes)
	time.Sleep(1 * time.Second)

	// run MPC nodes
	for nodeId := 0; nodeId < len(nodeNames); nodeId++ {
		go mpc_node.RunNode(nodeNames[nodeId], "localhost", 5040+nodeId,
			"../key_management/keys_certificates", os.Getenv("SCALE_MAMBA_PATH"),
			"debug", "../logging/log.log",
			"localhost:5008", "An MPC node deployed for tests.")
	}
	time.Sleep(1 * time.Second)

	buyerPubKey, buyerSecKey := key_management.GenerateKeypair()

	response := requestComputationToManager("max", buyerPubKey)

	resVecs := make([][]*big.Int, 3)
	var err error
	for i := 0; i < 3; i++ {
		assert.Equal(t, "", response[i].Error)
		resVecs[i], err = data_management.DecVec(response[i].Result, buyerPubKey, buyerSecKey)
		assert.NoError(t, err)
	}

	res := data_management.JoinSharesShamirFloat(resVecs)
	assert.Equal(t, []float64{1, 65, 225, 1}, res)
}

func requestComputationToManager(program string, pubKey []byte) [3]manager.ReturnMsg {
	params := map[string]string{"cols": "age,male,TenYearCHD,glucose"}
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}

	// client sends a JSON message to the server
	msg := manager.ComputationRequest{
		NodesNames: "Berlin_node,Paris_node,Ljubljana_node",
		//NodesNames:     "Rome_node,Leuven_node,Madrid_node",
		Program: program,
		//DatasetNames:   "framingham_heart_study_dataset1.csv,framingham_heart_study_dataset2.csv",
		DatasetNames: "framingham_heart_study_dataset1.csv",
		//DatasetNames:   "breast_cancer_dataset.csv",
		Params:         string(paramsBytes),
		ReceiverPubKey: base64.StdEncoding.EncodeToString(pubKey),
	}

	urlAddr := "http://localhost:5007/compute"
	payloadBuf := new(bytes.Buffer)
	err = json.NewEncoder(payloadBuf).Encode(msg)
	req, err := http.NewRequest("POST", urlAddr, payloadBuf)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var resBytes []byte
	resBytes, err = ioutil.ReadAll(response.Body)

	var res [3]manager.ReturnMsg
	err = json.Unmarshal(resBytes, &res)

	return res
}
