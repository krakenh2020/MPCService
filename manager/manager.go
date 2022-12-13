package manager

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/krakenh2020/MPCService/logging"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/krakenh2020/MPCService/data_provider"
	"github.com/krakenh2020/MPCService/mpc_engine"

	"github.com/gorilla/mux"
)

type MPCNode struct {
	Name        string `json:"name"`
	ScalePort   int    `json:"scale_port"`
	Address     string `json:"address"`
	MpcPubKey   []byte `json:"mpc_pub_key"`
	ScaleCert   []byte `json:"scale_cert"`
	SigPubKey   []byte `json:"sig_pub_key"`
	Description string `json:"description"`
}

type MPCNodes struct {
	mu          sync.Mutex
	list        []MPCNode
	nameToIndex map[string]int
	reqChan     []chan mpc_engine.Request
	outChan     []chan ReturnMsg
}

type Datasets struct {
	mu          sync.Mutex
	list        []data_provider.Dataset
	nameToIndex map[string]int
	reqChan     []chan data_provider.DatasetRequest
	outChan     []chan data_provider.DatasetReturn
}

type ComputationRequest struct {
	NodesNames     string
	Program        string
	DatasetNames   string
	Params         string
	ReceiverPubKey string
}

var mpcNodes MPCNodes
var datasets Datasets

// ReturnMsg is a struct defining how returns of the node server will
// be structured
type ReturnMsg struct {
	Error  string
	Result string
	Cols   string
}

func getMPCNodesHandler(w http.ResponseWriter, r *http.Request) {
	nodesListBytes, err := json.Marshal(mpcNodes.list)

	if err != nil {
		log.Error(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(nodesListBytes)
	if err != nil {
		log.Error(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func addDatasetHandler(w http.ResponseWriter, r *http.Request) {
	dataset := data_provider.Dataset{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &dataset)
	if err != nil {
		panic(err)
	}

	// Append our existing list of datasets
	datasets.mu.Lock()
	datasets.list = append(datasets.list, dataset)
	datasets.reqChan = append(datasets.reqChan, nil)
	datasets.outChan = append(datasets.outChan, nil)
	datasets.nameToIndex[dataset.Name] = len(datasets.list) - 1
	datasets.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusFound)
}

func getDatasetsHandler(w http.ResponseWriter, r *http.Request) {
	datasetsListBytes, err := json.Marshal(datasets.list)
	if err != nil {
		log.Error(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(datasetsListBytes)
	if err != nil {
		log.Error(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func requestComputation(w http.ResponseWriter, r *http.Request) {
	// todo: columns management, make a queue for requests
	var ret [3]ReturnMsg
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var req ComputationRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Error("cannot read request", err)
		return
	}
	log.Info("Manager: received a request for MPC computation")
	log.Debug("Manager: request", req)

	inputVecs := make([][]string, 3)
	for i := 0; i < 3; i++ {
		inputVecs[i] = make([]string, 0)
	}
	inputCols := make([][]string, 0)
	inputLinks := make([]string, 0)
	datasetNames := strings.Split(req.DatasetNames, ",")
	chosenNodes := strings.Split(req.NodesNames, ",")
	for _, dataName := range datasetNames {
		dataIndex := datasets.nameToIndex[dataName]
		pubKeys := make([][]byte, 3)
		certs := make([][]byte, 3)
		sigs := make([][]byte, 3)
		for i := 0; i < 3; i++ {
			nodeIndex := mpcNodes.nameToIndex[chosenNodes[i]]
			pubKeys[i] = mpcNodes.list[nodeIndex].MpcPubKey
			certs[i] = mpcNodes.list[nodeIndex].ScaleCert
			sigs[i] = mpcNodes.list[nodeIndex].SigPubKey
		}
		if val := datasets.reqChan[dataIndex]; val != nil {
			inChan := datasets.reqChan[dataIndex]
			dataReq := data_provider.DatasetRequest{DatasetName: dataName, NodesNames: chosenNodes,
				NodesPubKeys: pubKeys, NodesCerts: certs, NodesPubKeysSignatures: sigs}
			inChan <- dataReq

			outChan := datasets.outChan[dataIndex]
			retData := <-outChan

			if len(retData.EncVecs) == 0 {
				ret[0].Error = "data provider denied access"
				retBytes, err := json.Marshal(ret)
				if err != nil {
					log.Error(err)
				}
				_, err = w.Write(retBytes)
				if err != nil {
					log.Error(err)
				}
				return
			}

			for i := 0; i < 3; i++ {
				inputVecs[i] = append(inputVecs[i], retData.EncVecs[i])
			}
			inputCols = append(inputCols, retData.Cols)
		} else {
			inputLinks = append(inputLinks, datasets.list[dataIndex].Link)
		}

	}

	nodesAddr := make([]string, 3)
	nodesPorts := make([]string, 3)
	scaleCerts := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		nodesAddr[i] = mpcNodes.list[mpcNodes.nameToIndex[chosenNodes[i]]].Address
		nodesPorts[i] = strconv.Itoa(mpcNodes.list[mpcNodes.nameToIndex[chosenNodes[i]]].ScalePort)
		scaleCerts[i] = mpcNodes.list[mpcNodes.nameToIndex[chosenNodes[i]]].ScaleCert
	}
	nodePortsString := strings.Join(nodesPorts, ",")

	for i := 0; i < 3; i++ {
		inChan := mpcNodes.reqChan[mpcNodes.nameToIndex[chosenNodes[i]]]
		reqI := mpc_engine.Request{Program: req.Program, InputLinks: inputLinks, Params: req.Params,
			NodeId: i, NodesNames: chosenNodes, NodesAddrs: nodesAddr, NodesPorts: nodePortsString,
			ReceiverPubKey: req.ReceiverPubKey, InputVecs: inputVecs[i], InputCols: inputCols,
			ScaleCerts: scaleCerts}
		inChan <- reqI
	}
	log.Info("Manager: sent request for computation to ", req.NodesNames)

	for i := 0; i < 3; i++ {
		outChan := mpcNodes.outChan[mpcNodes.nameToIndex[chosenNodes[i]]]
		out := <-outChan
		ret[i] = out
	}
	log.Info("Manager: computation response received")

	retBytes, err := json.Marshal(ret)
	if err != nil {
		log.Error(err)
	}

	_, err = w.Write(retBytes)
	if err != nil {
		log.Error(err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func datasetsConnection(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, http.Header{
		"Sec-websocket-Protocol": websocket.Subprotocols(r),
	})
	if err != nil {
		log.Fatal(err)
	}

	var newDatasets []data_provider.Dataset
	err = ws.ReadJSON(&newDatasets)
	if err != nil {
		log.Fatal(err)
	}

	inChan := make(chan data_provider.DatasetRequest, 2)
	outChan := make(chan data_provider.DatasetReturn, 2)

	datasets.mu.Lock()
	for _, data := range newDatasets {
		datasets.list = append(datasets.list, data)
		datasets.reqChan = append(datasets.reqChan, inChan)
		datasets.outChan = append(datasets.outChan, outChan)
		datasets.nameToIndex[data.Name] = len(datasets.list) - 1
	}
	datasets.mu.Unlock()

	for {
		if len(inChan) > 0 {
			req := <-inChan
			log.Info("Manager: sending request for data ", req.DatasetName)
			msg, err := json.Marshal(req)
			if err != nil {
				log.Error(err)
			}

			err = ws.WriteJSON(msg)
			if err != nil {
				log.Error(err)
			}
			var ret data_provider.DatasetReturn
			err = ws.ReadJSON(&ret)
			if len(ret.EncVecs) == 0 {
				log.Error("Manager: request for data denied")
			} else {
				log.Info("Manager: received data ", req.DatasetName)
			}
			outChan <- ret

		} else {
			// check if the engine is ready
			err := ws.WriteJSON([]byte("ping"))
			if err != nil {
				goto removeDataset
			}
			var b []byte
			err = ws.ReadJSON(&b)
			if err != nil {
				goto removeDataset
			}
		}

		time.Sleep(200 * time.Millisecond)
	}
removeDataset:
	datasets.mu.Lock()
	for _, data := range newDatasets {
		index := 0
		for i, e := range datasets.list {
			if e == data {
				index = i
				break
			}
		}
		datasets.list = append(datasets.list[:index], datasets.list[index+1:]...)
		datasets.reqChan = append(datasets.reqChan[:index], datasets.reqChan[index+1:]...)
		datasets.outChan = append(datasets.outChan[:index], datasets.outChan[index+1:]...)
		for key, val := range datasets.nameToIndex {
			if val > index {
				datasets.nameToIndex[key] = val - 1
			}
		}
		delete(datasets.nameToIndex, data.Name)
	}
	datasets.mu.Unlock()
}

func mpcNodeConnection(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, http.Header{
		"Sec-websocket-Protocol": websocket.Subprotocols(r),
	})
	if err != nil {
		log.Fatal(err)
	}

	var msg MPCNode
	err = ws.ReadJSON(&msg)
	if err != nil {
		log.Fatal(err)
	}

	inChan := make(chan mpc_engine.Request, 2)
	outChan := make(chan ReturnMsg, 2)

	mpcNodes.mu.Lock()
	mpcNodes.list = append(mpcNodes.list, msg)
	mpcNodes.reqChan = append(mpcNodes.reqChan, inChan)
	mpcNodes.outChan = append(mpcNodes.outChan, outChan)
	mpcNodes.nameToIndex[msg.Name] = len(mpcNodes.list) - 1
	mpcNodes.mu.Unlock()

	for {
		if len(inChan) > 0 {
			req := <-inChan
			msg, err := json.Marshal(req)
			if err != nil {
				log.Error(err)
			}

			err = ws.WriteJSON(msg)
			if err != nil {
				log.Error(err)
			}
			var ret ReturnMsg
			err = ws.ReadJSON(&ret)
			outChan <- ret

		} else {
			// check if the engine is ready
			err := ws.WriteJSON([]byte("ping"))
			if err != nil {
				goto removeMPCnode
			}
			var b []byte
			err = ws.ReadJSON(&b)
			if err != nil {
				goto removeMPCnode
			}
		}

		time.Sleep(200 * time.Millisecond)
	}
removeMPCnode:
	mpcNodes.mu.Lock()
	index := 0
	for i, e := range mpcNodes.list {
		if e.Name == msg.Name {
			index = i
			break
		}
	}
	mpcNodes.list = append(mpcNodes.list[:index], mpcNodes.list[index+1:]...)
	mpcNodes.reqChan = append(mpcNodes.reqChan[:index], mpcNodes.reqChan[index+1:]...)
	mpcNodes.outChan = append(mpcNodes.outChan[:index], mpcNodes.outChan[index+1:]...)
	for key, val := range mpcNodes.nameToIndex {
		if val > index {
			mpcNodes.nameToIndex[key] = val - 1
		}
	}
	delete(mpcNodes.nameToIndex, msg.Name)

	mpcNodes.mu.Unlock()
}

func newRouters(assets string) (*mux.Router, *mux.Router) {
	r1 := mux.NewRouter()
	r1.HandleFunc("/hello", handler).Methods("GET")
	r1.HandleFunc("/nodes", getMPCNodesHandler).Methods("GET")
	r1.HandleFunc("/datasets", getDatasetsHandler).Methods("GET")
	r1.HandleFunc("/datasets", addDatasetHandler).Methods("POST")
	r1.HandleFunc("/compute", requestComputation).Methods("POST")

	var staticFileDirectory http.Dir
	if assets == "" {
		staticFileDirectory = http.Dir("./manager/assets/")
	} else {
		staticFileDirectory = http.Dir(assets)
	}
	staticFileHandler := http.FileServer(staticFileDirectory)
	r1.PathPrefix("/").Handler(staticFileHandler).Methods("GET")

	r2 := mux.NewRouter()
	r2.HandleFunc("/connect_mpc", mpcNodeConnection)
	r2.HandleFunc("/connect_data", datasetsConnection)

	return r1, r2
}

func RunManager(guiPort, servicePort int, assets string, logLevel, logFile, caFolder string) {
	// set up logging
	logging.LogSetUp(logLevel, logFile)

	mpcNodes.nameToIndex = make(map[string]int)
	datasets.nameToIndex = make(map[string]int)
	// The router is now formed by calling the `newRouter` constructor function
	// that we defined above. The rest of the code stays the same
	log.Info("Manager running on localhost:", guiPort)
	r1, r2 := newRouters(assets)

	go http.ListenAndServe(":"+strconv.Itoa(guiPort), r1)

	caCert, err := ioutil.ReadFile(caFolder + "/RootCA.crt")
	caCertPool := x509.NewCertPool()
	_ = caCertPool.AppendCertsFromPEM(caCert)
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(servicePort),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Second,
		TLSConfig: &tls.Config{ServerName: "manager", ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs: caCertPool},
		Handler: r2,
	}

	err = server.ListenAndServeTLS(caFolder+"/Manager.crt", caFolder+"/Manager.key")

	if err != nil {
		panic(err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}
