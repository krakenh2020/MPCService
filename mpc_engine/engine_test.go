package mpc_engine_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/krakenh2020/MPCService/key_management"

	"github.com/stretchr/testify/assert"
	"github.com/krakenh2020/MPCService/data_management"
	"github.com/krakenh2020/MPCService/mpc_engine"
)

func TestEngine(t *testing.T) {
	//run server
	log.SetLevel(log.DebugLevel)
	queue := make([]chan mpc_engine.Request, 3)
	out := make([]chan mpc_engine.Response, 3)

	var err error
	nodeNames := []string{"Ljubljana_node", "Berlin_node", "Paris_node"}
	for nodeId := 0; nodeId < 3; nodeId++ {
		pubKey, secKey, _, err := key_management.LoadKeysFromCertKey("../key_management/keys_certificates", nodeNames[nodeId])
		if err != nil {
			t.Fatal(err)
		}
		queue[nodeId] = make(chan mpc_engine.Request, 1)
		out[nodeId] = make(chan mpc_engine.Response, 1)
		go mpc_engine.ScaleEngine(os.Getenv("SCALE_MAMBA_PATH"), queue[nodeId], out[nodeId], pubKey, secKey, 5012+nodeId,
			nodeNames[nodeId], "../key_management/keys_certificates")
	}

	time.Sleep(1 * time.Second)

	scaleCerts := make([][]byte, 3)
	for nodeId := 0; nodeId < 3; nodeId++ {
		scaleCerts[nodeId], err = key_management.LoadCertificate(nodeNames[nodeId], "../key_management/keys_certificates")
		if err != nil {
			t.Fatal(err)
		}
	}

	params := map[string]string{"cols": "age,male,TenYearCHD,glucose", "NUM_CLUSTERS": "3"}
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}
	for nodeId := 0; nodeId < 3; nodeId++ {
		req := mpc_engine.Request{"k-means",
			[]string{"https://unilj-my.sharepoint.com/:t:/g/personal/tilen_marc_fmf_uni-lj_si/ERIeB11IHdNEm6XZJIDIO_gB5frXkd70ygnaJJUYboUnJw?e=vxSOoV&download=1"},
			nil,
			nil,
			string(paramsBytes),
			nodeId,
			nodeNames,
			[]string{"localhost", "localhost", "localhost"},
			"5012,5013,5014",
			scaleCerts,
			"",
		}
		queue[nodeId] <- req
	}

	var ret [3]mpc_engine.Response
	var shares = make([][]*big.Int, 3)
	for nodeId := 0; nodeId < 3; nodeId++ {
		ret[nodeId] = <-out[nodeId]
		shares[nodeId] = ret[nodeId].Vec
	}

	res := data_management.JoinSharesShamirFloat(shares)
	fmt.Println("Result", res)
	//assert.Equal(t, []float64{1, 63, 4, 1, 30, 0, 0, 1, 0, 313, 180, 110, 33.109999656677246, 95, 103, 1}, res)
	assert.Equal(t, 4*3+3, len(res))
}

//func TestEngineErrors(t *testing.T) {
//	time.Sleep(5 * time.Second)
//	//run server
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		go mpc_engine.RunEngine(nodeId, os.Getenv("SM_PATH"), "info", "../logging/engine_log.log",
//			5012+nodeId, "test", "../key_management/keys", []string{"localhost", "localhost", "localhost"}, []int{5000, 5001, 5002})
//	}
//
//	time.Sleep(5 * time.Second)
//	tcpAddr := make([]*net.TCPAddr, 3)
//	var err error
//	c := make([]*net.TCPConn, 3)
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		tcpAddr[nodeId], err = net.ResolveTCPAddr("tcp", "localhost:"+strconv.Itoa(5012+nodeId))
//		c[nodeId], err = net.DialTCP("tcp", nil, tcpAddr[nodeId])
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	// set parameters
//	//params := map[string]string{"COLS": "3"}
//	//paramsBytes, err := json.Marshal(params)
//	//if err != nil {
//	//	t.Fatal(err)
//	//}
//	paramsBytes := []byte{}
//
//	// send bad link
//	req := computation.ScaleRequest{"max_sfix",
//		"https://unilj-my.sharepoint.com/:t:/g/personal/tilen_marc_fmf_uni-lj_si/wrong_link&download=1",
//		string(paramsBytes)}
//	reqBytes, err := json.Marshal(req)
//	if err != nil {
//		t.Fatal(err)
//	}
//	b := make([]byte, 1024*1024)
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		_, err := c[nodeId].Write([]byte("ping"))
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Read(b)
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Write(reqBytes)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	// receive error
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		n, err := c[nodeId].Read(b)
//		if err != nil {
//			t.Fatal(err)
//		}
//		//
//		//fmt.Println(string(b[:n]))
//		assert.Equal(t, "error", string(b[:n])[:5])
//	}
//
//	// send good link, bad parameters
//	req = computation.ScaleRequest{"max_sfix",
//		"https://unilj-my.sharepoint.com/:t:/g/personal/tilen_marc_fmf_uni-lj_si/EVXan3OtjOdJmYyM7J7lqJYBK6aDPoN9Bku8fEk9dcu4Ig?e=FMmUWk&download=1",
//		"bla"}
//	reqBytes, err = json.Marshal(req)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		_, err := c[nodeId].Write([]byte("ping"))
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Read(b)
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Write(reqBytes)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	// receive error
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		n, err := c[nodeId].Read(b)
//		if err != nil {
//			t.Fatal(err)
//		}
//		assert.Equal(t, "error", string(b[:n])[:5])
//	}
//
//	// send good link, good parameters, nonexistent program
//	req = computation.ScaleRequest{"unknown_program",
//		"https://unilj-my.sharepoint.com/:t:/g/personal/tilen_marc_fmf_uni-lj_si/EVXan3OtjOdJmYyM7J7lqJYBK6aDPoN9Bku8fEk9dcu4Ig?e=FMmUWk&download=1",
//		string(paramsBytes)}
//	reqBytes, err = json.Marshal(req)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		_, err := c[nodeId].Write([]byte("ping"))
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Read(b)
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Write(reqBytes)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	// receive error
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		n, err := c[nodeId].Read(b)
//		if err != nil {
//			t.Fatal(err)
//		}
//		//
//		fmt.Println(string(b[:n]))
//		assert.Equal(t, "error", string(b[:n])[:5])
//	}
//
//	// stop the engine
//	req = computation.ScaleRequest{"exit", "", ""}
//	reqBytes, err = json.Marshal(req)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for nodeId := 0; nodeId < 3; nodeId++ {
//		_, err := c[nodeId].Write([]byte("ping"))
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Read(b)
//		if err != nil {
//			t.Fatal(err)
//		}
//		_, err = c[nodeId].Write(reqBytes)
//		if err != nil {
//			t.Error(err)
//		}
//		n, err := c[nodeId].Read(b)
//		if err != nil {
//			t.Error(err)
//		}
//		assert.Equal(t, "exit", string(b[:n]))
//	}
//}
