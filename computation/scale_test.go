package computation_test

import (
	"math/big"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/krakenh2020/MPCService/computation"
)

func TestInputPrepare(t *testing.T) {
	inLen := 10
	for nodeId := 0; nodeId < 3; nodeId++ {
		in := make([]*big.Int, inLen)
		for i := 0; i < inLen; i++ {
			in[i] = big.NewInt(int64(i))
		}

		err := computation.InputPrepare(nodeId, in, nil, os.Getenv("SCALE_MAMBA_PATH"))
		if err != nil {
			t.Fatal("Preparing input failed", err)
		}
	}
}

func TestRunScale(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	for nodeId := 0; nodeId < 3; nodeId++ {
		params := map[string]string{"LEN": "10", "COLS": "5"}
		go computation.RunScale(nodeId, "max", params, "5550,5551,5552", os.Getenv("SCALE_MAMBA_PATH"))
	}

	time.Sleep(5 * time.Second)
}
