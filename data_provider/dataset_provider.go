package data_provider

import (
	"crypto"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/krakenh2020/MPCService/data_management"
	"github.com/krakenh2020/MPCService/logging"
)

// todo: error management
// todo: share with selection
type Dataset struct {
	Name        string `json:"name"`
	Size        string `json:"size"`
	Cols        string `json:"cols"`
	SharedWith  string `json:"shared_with"`
	Link        string `json:"link"`
	Description string `json:"description"`
}

type DatasetRequest struct {
	DatasetName            string
	NodesNames             []string
	Params                 string
	NodesPubKeys           [][]byte
	NodesCerts             [][]byte
	NodesPubKeysSignatures [][]byte
}

type DatasetReturn struct {
	EncVecs []string
	Cols    []string
}

func RunDatasetProvider(name string, loc string, logLevel, logFile, managerAddr, certFolder string, sharedWith []string) {
	// set up logging
	logging.LogSetUp(logLevel, logFile)
	log.Info("Dataset server "+name+", dataset location: ", loc, ", manager address: ", managerAddr)
	datasets, locations := getDatasetsData(loc, sharedWith)

	managerConn(name, managerAddr, datasets, locations, certFolder, sharedWith)
}

func managerConn(name, managerAddr string, datasets []Dataset, locations map[string]string,
	certFolder string, sharedWith []string) {
	u := url.URL{Scheme: "wss", Host: managerAddr, Path: "/connect_data"}
	cert, err := tls.LoadX509KeyPair(certFolder+"/"+name+".crt", certFolder+"/"+name+".key")

	caCert, err := ioutil.ReadFile(certFolder + "/" + "/RootCA.crt")
	caCertPool := x509.NewCertPool()
	_ = caCertPool.AppendCertsFromPEM(caCert)

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	}

	sharedWithMap := make(map[string]bool, 0)
	for _, e := range sharedWith {
		sharedWithMap[e] = true
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("Data provider: error dialing the manager")
		return
	}
	defer conn.Close()

	err = conn.WriteJSON(datasets)
	if err != nil {
		log.Error("Data provider: error sending info to the manager")
		return
	}

	for {
		var b []byte
		err = conn.ReadJSON(&b)
		if err != nil {
			log.Error("lost connection with the manager 1 ", err)
			return
		}
		if string(b) == "ping" {
			err = conn.WriteJSON([]byte("pong"))
			if err != nil {
				log.Error("lost connection with the manager 2")
				return
			}
			continue
		}

		var msg DatasetRequest
		err = json.Unmarshal(b, &msg)
		log.Info("Data provider: received a request for dataset ", msg.DatasetName)

		log.Debug(sharedWithMap)
		log.Debug(msg)
		check, err := checkIfAllowed(msg, sharedWithMap, caCertPool)
		if check == false {
			log.Info("Data provider: access denied ", err)
			err = conn.WriteJSON([]byte{})
			if err != nil {
				log.Error("failed to return the response: ", err)
			}
			continue
		}

		response, err := prepareDataset(msg, locations)
		if err != nil {
			log.Fatal("error preparing data", err)
		}

		err = conn.WriteJSON(response)
		if err != nil {
			log.Error("failed to return a response: ", err)
		} else {
			log.Info("Data provider: dataset provided")
		}
	}
}

func getDatasetsData(loc string, sharedWith []string) ([]Dataset, map[string]string) {
	files, err := ioutil.ReadDir(loc)
	if err != nil {
		log.Fatal(err)
	}

	datasets := make([]Dataset, 0)
	locations := make(map[string]string)
	for _, file := range files {
		name := file.Name()

		_, cols, vec, err := data_management.CsvToVec(loc + "/" + name)
		if err != nil {
			log.Fatal(err)
		}

		dataset := Dataset{
			Name:       name,
			SharedWith: strings.Join(sharedWith, ","),
			Cols:       strings.Join(cols, ","),
			Size:       strconv.Itoa(len(vec)),
		}
		datasets = append(datasets, dataset)
		locations[name] = loc + "/" + name
	}

	return datasets, locations
}

func prepareDataset(req DatasetRequest, locations map[string]string) (*DatasetReturn, error) {
	vec, cols, _, err := data_management.CsvToVec(locations[req.DatasetName])
	if err != nil {
		return nil, err
	}

	shares, err := data_management.CreateSharesShamir(vec)
	if err != nil {
		return nil, err
	}

	var response DatasetReturn
	response.EncVecs = make([]string, 3)
	for i := int64(0); i < 3; i++ {
		response.EncVecs[i], err = data_management.EncryptVec(shares[i], req.NodesPubKeys[i])
		if err != nil {
			return nil, err
		}
	}
	response.Cols = cols
	return &response, nil
}

func checkIfAllowed(req DatasetRequest, sharedWith map[string]bool, caCertPool *x509.CertPool) (bool, error) {
	if sharedWith["all"] {
		return true, nil
	}

	nodesMap := make(map[string]bool, 0)
	for _, e := range req.NodesNames {
		nodesMap[e] = true
	}
	if len(nodesMap) != 3 {
		return false, fmt.Errorf("dataset must be shared among 3 different nodes")
	}

	for i, e := range req.NodesNames {
		log.Debug("Data provider: checking ", e)
		if ok, _ := sharedWith[e]; !ok {
			return false, fmt.Errorf("dataset not shared with node " + e)
		}

		block, _ := pem.Decode(req.NodesCerts[i])
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return false, err
		}
		if cert.Subject.CommonName != e {
			return false, fmt.Errorf("common name in cert does not match provided node name")
		}

		// check if cert was signed by the CA
		opts := x509.VerifyOptions{
			Roots: caCertPool,
		}
		_, err = cert.Verify(opts)
		if err != nil {
			return false, err
		}

		// check if the public key is confirmed by the node
		publicKey := cert.PublicKey.(*rsa.PublicKey)
		err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, req.NodesPubKeys[i], req.NodesPubKeysSignatures[i])
		if err != nil {
			return false, err
		}
	}

	return true, nil

}
