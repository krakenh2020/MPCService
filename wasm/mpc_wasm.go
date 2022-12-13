package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"syscall/js"

	"github.com/krakenh2020/MPCService/data_management"
	"github.com/krakenh2020/MPCService/key_management"
)

func registerCallbacks() {
	js.Global().Set("JoinSharesShamir", js.FuncOf(JoinSharesShamir))
	js.Global().Set("GenerateKeypair", js.FuncOf(GenerateKeypair))
	js.Global().Set("SplitCsvText", js.FuncOf(SplitCsvText))
	js.Global().Set("VecToCsvText", js.FuncOf(VecToCsvText))
}
func main() {
	c := make(chan struct{}, 0)
	// register functions
	registerCallbacks()
	fmt.Println("WASM Go Initialized")
	<-c
}

// Generates a keypair for the MPC Nodes, only for testing
// args pubKey, secKey, share0, share1, share2
func JoinSharesShamir(this js.Value, args []js.Value) interface{} {
	pubKey, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		panic("Error in JoinSharesShamir decoding pubKey")
	}
	secKey, err := base64.StdEncoding.DecodeString(args[1].String())
	if err != nil {
		panic("Error in JoinSharesShamir decoding secKey")
	}

	numNodes := 3
	sharesArray := make([][]*big.Int, numNodes)
	//var encVec data_management.VecEnc
	//var shareBytesEnc []byte
	for i := 2; i < numNodes+2; i++ {
		sharesArray[i-2], err = data_management.DecVec(args[i].String(), pubKey, secKey)
		if err != nil {
			fmt.Println("Error", err)
			panic("Error in JoinSharesShamir decrypting")
		}
	}
	//fmt.Println("shares", sharesArray)
	res := data_management.JoinSharesShamirFloat(sharesArray)
	//fmt.Println("result", res)

	ret, err := json.Marshal(res)
	retString := base64.StdEncoding.EncodeToString(ret)
	if err != nil {
		panic("Error in JoinSharesShamir marshalling")
	}

	return retString
}

func GenerateKeypair(this js.Value, args []js.Value) interface{} {
	pubKey, secKey := key_management.GenerateKeypair()
	pubKeyString := base64.StdEncoding.EncodeToString(pubKey)
	secKeyString := base64.StdEncoding.EncodeToString(secKey)

	return []interface{}{pubKeyString, secKeyString}
}

// Splits the txt into shares
// args txt, pubKey0, pubKey1, pubKey2
func SplitCsvText(this js.Value, args []js.Value) interface{} {
	numNodes := 3
	pubKeys := make([][]byte, numNodes)
	var err error
	for i := 0; i < numNodes; i++ {
		pubKeys[i], err = base64.StdEncoding.DecodeString(args[1+i].String())
		if err != nil {
			panic("Error in SplitCsvText decoding pubKey")
		}
	}

	txt := args[0].String()
	vec, cols, err := data_management.CsvTxtToVec(txt)
	if err != nil {
		panic("Error in SplitCsvText reading")
	}
	shares, err := data_management.CreateSharesShamir(vec)
	if err != nil {
		panic("Error in SplitCsvText splitting")
	}

	encShares := make([]string, 3)
	for i := int64(0); i < 3; i++ {
		encShares[i], err = data_management.EncryptVec(shares[i], pubKeys[i])
		if err != nil {
			panic("Error in SplitCsvText encrypting")
		}
		//encBytes, err := json.Marshal(enc)
		//encShares[i] = string(encBytes)

		if err != nil {
			panic("Error in SplitCsvText marshalling")
		}
	}
	colsString := strings.Join(cols, ",")

	return []interface{}{encShares[0], encShares[1], encShares[2], colsString}
}

func VecToCsvText(this js.Value, args []js.Value) interface{} {
	sharesFloatsString := args[0].String()
	colsString := args[1].String()
	funcName := args[2].String()

	sharesFloatsBytes, err := base64.StdEncoding.DecodeString(sharesFloatsString)
	if err != nil {
		panic("Error in VecToCsvText decoding vector")
	}
	var sharesFloats []float64
	err = json.Unmarshal(sharesFloatsBytes, &sharesFloats)
	if err != nil {
		panic("Error in VecToCsvText unmarshalling")
	}

	cols := strings.Split(colsString, ",")
	res, err := data_management.ResultsToCsvText(sharesFloats, cols, funcName)
	if err != nil {
		fmt.Println("Error:", err)
		panic("Error in VecToCsvText csv text")
	}
	return res
}
