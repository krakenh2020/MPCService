package data_management

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/krakenh2020/MPCService/key_management"
)

func TestSharesShamir(t *testing.T) {
	n := 100
	a, err := NewUniformRandomVector(n, MPCPrimeHalf)
	assert.NoError(t, err)

	shares, err := CreateSharesShamir(a)
	assert.NoError(t, err)

	b, err := JoinSharesShamir(shares)
	assert.NoError(t, err)

	assert.Equal(t, a, b)
}

func TestEncVec(t *testing.T) {
	n := 100
	a, err := NewUniformRandomVector(n, MPCPrime)
	assert.NoError(t, err)

	pubKey, secKey := key_management.GenerateKeypair()

	e, err := EncryptVec(a, pubKey)
	assert.NoError(t, err)

	d, err := DecVec(e, pubKey, secKey)
	assert.NoError(t, err)

	assert.Equal(t, a, d)
}

func TestCsvFileSplitJoin(t *testing.T) {
	nodeNames := []string{"Berlin_node", "Paris_node", "Ljubljana_node"}
	pubKeys := make([][]byte, 3)
	secKeys := make([][]byte, 3)
	var err error
	for i := 0; i < 3; i++ {
		pubKeys[i], secKeys[i], _, err = key_management.LoadKeysFromCertKey("../key_management/keys_certificates", nodeNames[i])
		assert.NoError(t, err)
	}
	vec, _, _, err := SplitCsvFile("framingham_tiny.csv", "framingham_tiny_enc.txt", pubKeys)
	assert.NoError(t, err)

	shares := make([][]*big.Int, 3)
	for i := 0; i < 3; i++ {
		shares[i], _, err = ReadShare("framingham_tiny_enc.txt", pubKeys[i], secKeys[i], i)
		assert.NoError(t, err)
	}

	b := JoinSharesShamirFloat(shares)
	for i, _ := range vec {
		assert.Equal(t, math.Trunc(vec[i]*10), math.Trunc(b[i]*10))
	}
}

func TestFileDownload(t *testing.T) {
	err := DownloadShare("https://unilj-my.sharepoint.com/:t:/g/personal/tilen_marc_fmf_uni-lj_si/EVXan3OtjOdJmYyM7J7lqJYBK6aDPoN9Bku8fEk9dcu4Ig?e=auX45f&download=1", "test2_enc.txt")
	assert.NoError(t, err)

	shares := make([][]*big.Int, 3)
	nodeNames := []string{"Berlin_node", "Paris_node", "Ljubljana_node"}
	for i := 0; i < 3; i++ {
		pubKey, secKey, _, err := key_management.LoadKeysFromCertKey("../key_management/keys_certificates", nodeNames[i])
		shares[i], _, err = ReadShare("test2_enc.txt", pubKey, secKey, i)
		assert.NoError(t, err)
	}
	err = DeleteShare("test2_enc.txt")
	assert.NoError(t, err)

	b, err := JoinSharesShamir(shares)
	assert.NoError(t, err)

	assert.Equal(t, big.NewInt(40894464), b[1])
}

func TestResultsToCsvText(t *testing.T) {
	cols := []string{"male", "age", "education", "currentSmoker", "cigsPerDay"}

	a := []float64{1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54, 1.55, 23.4, 0, 54}
	_, err := ResultsToCsvText(a, cols, "stats")
	assert.NoError(t, err)

	b := []float64{1.55, 23.4, 0, 54}
	_, err = ResultsToCsvText(b, cols, "linear_regression")
	assert.NoError(t, err)
}
