package mpc_engine

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/krakenh2020/MPCService/computation"
	"github.com/krakenh2020/MPCService/data_management"
)

// Todo: shutdown request, low priority

// Request is a struct defining how request to the node servers should
// be given.
type Request struct {
	Program        string
	InputLinks     []string
	InputVecs      []string
	InputCols      [][]string
	Params         string
	NodeId         int
	NodesNames     []string
	NodesAddrs     []string
	NodesPorts     string
	ScaleCerts     [][]byte
	ReceiverPubKey string // only used outside of engine
}

type Response struct {
	Vec  []*big.Int
	Cols []string
	Msg  string
}

func ScaleEngine(sm string, tasksBacklog chan Request, output chan Response,
	pubKey, secKey []byte, scalePort int, privateCert, certLoc string) {
	var err error
	var response Response
	for {
		req := <-tasksBacklog

		// prepare settings of SCALE-MAMBA
		err = computation.SetUpScale(req.NodeId, req.NodesNames, req.NodesAddrs, sm, req.ScaleCerts, privateCert, certLoc)
		if err != nil {
			log.Fatal("Error preparing SCALE: ", err)
		}

		// load parameters of the computation
		var params map[string]string
		if req.Params != "" {
			err = json.Unmarshal([]byte(req.Params), &params)
			if err != nil {
				e := "error, computation failed, parameters error "
				log.Error(e, err)
				response.Msg = e
				output <- response
				continue
			}
		} else {
			params = map[string]string{}
		}

		// download, read and prepare data for SCALE
		_, numCols, numInput, cols, e := data_management.PrepareData(req.InputLinks, req.InputVecs, req.InputCols, req.NodeId, sm, params, pubKey, secKey)
		if e != "" {
			response.Msg = e
			output <- response
			continue
		}

		// set up the parameters
		params["COLS"] = strconv.Itoa(numCols)
		params["LEN"] = strconv.Itoa(numInput)
		if _, ok := params["cols"]; ok {
			delete(params, "cols")
		}

		// execute the computation of the node
		if strconv.Itoa(scalePort) != strings.Split(req.NodesPorts, ",")[req.NodeId] {
			log.Error(fmt.Errorf("error in port specification " + strconv.Itoa(scalePort) + " " + strings.Split(req.NodesPorts, ",")[req.NodeId]))
			continue
		}
		err = computation.RunScale(req.NodeId, req.Program, params, req.NodesPorts, sm)
		if err != nil {
			e := "error, computation failed, node trigger error"
			response.Msg = e
			output <- response
			if err.Error() == "function not supported" {
				log.Error(e, err)
				continue
			} else {
				log.Fatal(e, err)
			}
		}

		// load result
		res, err := computation.LoadResultShares(req.NodeId, sm)
		if err != nil {
			e := "error, computation failed, error reading result"
			response.Msg = e
			output <- response
			log.Error(err)
			continue
		}

		response.Vec = res
		response.Cols = cols

		// todo clean data
		output <- response
	}
}
