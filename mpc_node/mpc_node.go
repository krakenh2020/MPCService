package mpc_node

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/krakenh2020/MPCService/data_management"
	"github.com/krakenh2020/MPCService/key_management"
	"github.com/krakenh2020/MPCService/logging"
	"github.com/krakenh2020/MPCService/manager"
	"github.com/krakenh2020/MPCService/mpc_engine"
)

// RunNode starts a node server at localhost.
func RunNode(name string, myAddr string, scalePort int, certFolder, sm string, logLevel, logFile string,
	managerAddr string, description string) {
	// set up logging
	logging.LogSetUp(logLevel, logFile)
	log.Info("MPC "+name+" is running with scale port ", scalePort, "; address ", myAddr,
		"; key name: ", "; certificate location: ", certFolder,
		"; using SCALE-MAMBA ", sm, "; connecting to manager on address ", managerAddr)

	// make a queue for MPC computation requests
	queue := make(chan mpc_engine.Request, 100)
	out := make(chan mpc_engine.Response, 100)

	pubKey, secKey, sig, err := key_management.LoadKeysFromCertKey(certFolder, name)
	if err != nil {
		log.Fatal(err)
	}

	if managerAddr != "" {
		go managerConn(name, myAddr, managerAddr, pubKey, certFolder, sig, scalePort, queue, out, description)
	}

	mpc_engine.ScaleEngine(sm, queue, out, pubKey, secKey, scalePort, name, certFolder)
}

func managerConn(name, myAddr, managerAddr string, pubKey []byte, certFolder string, sig []byte,
	scalePort int, queue chan mpc_engine.Request, out chan mpc_engine.Response, description string) {
	u := url.URL{Scheme: "wss", Host: managerAddr, Path: "/connect_mpc"}

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

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("Data provider: error dialing the manager")
		return
	}
	defer conn.Close()

	// todo: double load
	scaleCrt, err := key_management.LoadCertificate(name, certFolder)
	if err != nil {
		log.Error("error loading pub key")
		return
	}

	msg := manager.MPCNode{Name: name, Address: myAddr,
		ScaleCert:   scaleCrt,
		MpcPubKey:   pubKey,
		SigPubKey:   sig,
		Description: description,
		ScalePort:   scalePort,
	}
	err = conn.WriteJSON(msg)
	if err != nil {
		log.Error("error sending info to the manager")
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

		var msg mpc_engine.Request
		err = json.Unmarshal(b, &msg)

		errMsg := manager.ReturnMsg{Error: "",
			Result: ""}
		if err != nil {
			log.Error("failed to read the message: ", err)
			errMsg.Error = "failed to read the message: " + err.Error()
			err = conn.WriteJSON(errMsg)
			if err != nil {
				log.Error("failed to return a response:", err)
			} else {
				log.Info("Error response sent")
			}
			return
		}

		log.Info("Server: Received a request to start computation of "+msg.Program+" from ", conn.RemoteAddr())
		retMsg, err := RequestComputation(msg, queue, out)

		if err != nil {
			log.Error("failed to do a computation: ", err)
			errMsg.Error = "failed to do a computation: " + err.Error()
			err = conn.WriteJSON(errMsg)
			if err != nil {
				log.Error("failed to return a response:", err)
			} else {
				log.Info("Error response sent")
			}
			return
		}
		log.Info("Server: Computation successful")

		log.Debug("return of computation:", retMsg)
		err = conn.WriteJSON(retMsg)
		if err != nil {
			log.Error("failed to return a response: ", err)
		} else {
			log.Info("Server: Return message sent")
		}
	}
}

func RequestComputation(msg mpc_engine.Request, queue chan mpc_engine.Request, out chan mpc_engine.Response) (manager.ReturnMsg, error) {
	// TODO: concurrency, it assumes that the computation requests come in the right order
	queue <- msg
	// get output
	var resEnc = ""
	var errMsg = ""
	res := <-out
	if res.Msg != "exit" && (len(res.Msg) != 5 || res.Msg[:5] != "error") {
		// encrypt output
		pubKey, err := base64.StdEncoding.DecodeString(msg.ReceiverPubKey)
		if err != nil {
			log.Error("Decoding public key failed, ", err)
			return manager.ReturnMsg{}, err
		}

		resEnc, err = data_management.EncryptVec(res.Vec, pubKey)
		if err != nil {
			log.Error("Encrypting result failed, ", err)
			return manager.ReturnMsg{}, err
		}

	} else if len(res.Msg) >= 5 && res.Msg[:5] == "error" {
		errMsg = res.Msg
	}

	ret := manager.ReturnMsg{Error: errMsg, Result: resEnc, Cols: strings.Join(res.Cols, ",")}

	return ret, nil
}
