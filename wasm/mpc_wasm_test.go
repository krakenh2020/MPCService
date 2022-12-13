package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"syscall/js"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallbacks(t *testing.T) {
	registerCallbacks()
}

func TestRunMain(t *testing.T) {
	timeout := time.After(3 * time.Second)
	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()
	select {
	case <-timeout: //main() never returns, so we run into the timeout
		assert.True(t, true)
	case <-done: //if main() returns something went wrong
		assert.True(t, false)
	}
}

func TestGenerateKeypair(t *testing.T) {
	keypair := GenerateKeypair(js.Null(), []js.Value{})
	sk, pk := spreadTwoEncodedStrings(keypair)
	assert.True(t, len(sk) > 0)
	assert.True(t, len(pk) > 0)
	assert.True(t, sk[0] != 0)
	assert.True(t, pk[0] != 0)
	assert.False(t, bytes.Equal(pk, sk))
}

func TestSplitCsvText(t *testing.T) {
	keypair1 := GenerateKeypair(js.Null(), []js.Value{})
	_, pk1 := spreadTwoEncodedStrings(keypair1)
	keypair2 := GenerateKeypair(js.Null(), []js.Value{})
	_, pk2 := spreadTwoEncodedStrings(keypair2)
	keypair3 := GenerateKeypair(js.Null(), []js.Value{})
	_, pk3 := spreadTwoEncodedStrings(keypair3)
	csvStringBytes := []byte("Test,CSV,moreTest")
	encodedString := base64.StdEncoding.EncodeToString(csvStringBytes)
	arguments := []js.Value{js.ValueOf(encodedString), js.ValueOf(base64.StdEncoding.EncodeToString(pk1)),
		js.ValueOf(base64.StdEncoding.EncodeToString(pk2)), js.ValueOf(base64.StdEncoding.EncodeToString(pk3))}
	retVal := js.ValueOf(SplitCsvText(js.Null(), arguments))
	share1 := retVal.Index(0).String()
	share2 := retVal.Index(1).String()
	share3 := retVal.Index(2).String()
	colsString := retVal.Index(3).String()
	colsStringBytes, err := base64.StdEncoding.DecodeString(colsString)
	if err != nil {
		fmt.Println("error decoding colsString")
	}
	assert.True(t, len(share1) > 0)
	assert.True(t, len(share2) > 0)
	assert.True(t, len(share3) > 0)
	assert.True(t, bytes.Equal(colsStringBytes, csvStringBytes))
}

func TestJoinSharesShamir(t *testing.T) {
	keypair1 := GenerateKeypair(js.Null(), []js.Value{})
	sk1, pk1 := spreadTwoEncodedStrings(keypair1)
	csvStringBytes := []byte("Test,CSV,moreTest")
	encodedString := base64.StdEncoding.EncodeToString(csvStringBytes)
	arguments := []js.Value{js.ValueOf(encodedString), js.ValueOf(base64.StdEncoding.EncodeToString(pk1)),
		js.ValueOf(base64.StdEncoding.EncodeToString(pk1)), js.ValueOf(base64.StdEncoding.EncodeToString(pk1))}
	retVal := js.ValueOf(SplitCsvText(js.Null(), arguments))
	share1 := retVal.Index(0).String()
	share2 := retVal.Index(1).String()
	share3 := retVal.Index(2).String()

	arguments = []js.Value{js.ValueOf(base64.StdEncoding.EncodeToString(pk1)), js.ValueOf(base64.StdEncoding.EncodeToString(sk1)),
		js.ValueOf(share1), js.ValueOf(share2), js.ValueOf(share3)}

	_ = JoinSharesShamir(js.Null(), arguments)
	//fmt.Println(js.ValueOf(joinedShares).Get("length").Int())

}

func arrayToJs(input []byte) interface{} {
	tmpJS := js.Global().Get("Uint8Array").New(len(input))
	js.CopyBytesToJS(tmpJS, input)
	return tmpJS
}
func spreadTwoArrays(input interface{}) (sk, pk []byte) {
	jsVal := js.ValueOf(input)
	sk = make([]byte, jsVal.Index(0).Get("length").Int())
	js.CopyBytesToGo(sk, jsVal.Index(0))
	pk = make([]byte, jsVal.Index(1).Get("length").Int())
	js.CopyBytesToGo(pk, jsVal.Index(1))
	return
}

func spreadTwoEncodedStrings(input interface{}) ([]byte, []byte) {
	jsVal := js.ValueOf(input)
	pk, err := base64.StdEncoding.DecodeString(jsVal.Index(0).String())
	if err != nil {
		fmt.Println("Error in decoding publicKey")
	}
	sk, err := base64.StdEncoding.DecodeString(jsVal.Index(1).String())
	if err != nil {
		fmt.Println("Error in decoding secretKey")
	}
	return sk, pk
}

func TestVecToCsvText(t *testing.T) {
	a := []float64{1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54}

	ret, err := json.Marshal(a)
	assert.NoError(t, err)
	retString := base64.StdEncoding.EncodeToString(ret)

	cols := "male,age,education,currentSmoker,cigsPerDay"
	funcName := "stats"

	arguments := []js.Value{js.ValueOf(retString), js.ValueOf(cols), js.ValueOf(funcName)}

	_ = VecToCsvText(js.Null(), arguments)
	assert.NoError(t, err)
}
