package computation

import (
	"bufio"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// InputPrepare is a helping function that loads the data shares provided to
// the node for the MPC computation.
func InputPrepare(nodeId int, shares, privateIn []*big.Int, sm string) error {
	f, err := os.Create(sm + "/Input/input_shares" + strconv.Itoa(nodeId) + ".txt")
	if err != nil {
		return err
	}

	for j := 0; j < len(shares); j++ {
		_, err = f.WriteString(strconv.Itoa(nodeId) + " " + shares[j].String() + "\n")
		if err != nil {
			return err
		}
	}
	err = f.Close()
	if err != nil {
		return err
	}

	f2, err := os.Create(sm + "/Input/private_input" + strconv.Itoa(nodeId) + ".txt")
	if err != nil {
		return err
	}

	for j := 0; j < len(privateIn); j++ {
		_, err = f2.WriteString(strconv.Itoa(nodeId) + " " + privateIn[j].String() + "\n")
		if err != nil {
			return err
		}
	}
	err = f2.Close()

	return nil
}

// LoadResultShares is a helping function that loads the result obtained by
// MPC computation.
func LoadResultShares(nodeId int, sm string) ([]*big.Int, error) {
	f, err := os.Open(sm + "/Input/output_shares" + strconv.Itoa(nodeId) + ".txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	countLines := 0
	res := make([]*big.Int, countLines)
	for scanner.Scan() {
		countLines++
		text := scanner.Text()
		vals := strings.Split(text, " ")
		val, _ := new(big.Int).SetString(vals[1], 10)
		res = append(res, val)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return res, nil
}

// LoadResultPrivate is a helping function that loads private result obtained by
// MPC computation.
func LoadResultPrivate(nodeId int, sm string) ([]*big.Int, error) {
	f, err := os.Open(sm + "/Input/private_output.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	countLines := 0
	res := make([]*big.Int, countLines)
	for scanner.Scan() {
		countLines++
		text := scanner.Text()
		val, _ := new(big.Int).SetString(text, 10)
		res = append(res, val)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return res, nil
}

func RunScale(nodeId int, funcName string, paramsMap map[string]string, mpcPorts string, sm string) error {
	var err error
	start := time.Now()
	err = PrepareMambaProgram(nodeId, funcName, paramsMap, sm)
	elapsed := time.Since(start)
	log.Info("Mamba: Compiling took ", elapsed.Seconds(), " seconds")
	if err != nil {
		return err
	}

	// start SCALE node that will prepare itself for future computation
	cmdStr := "./Player.x " + strconv.Itoa(nodeId) + " -dOT -pns " + mpcPorts + " Programs/MPCService/node" + strconv.Itoa(nodeId)

	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = sm

	if log.GetLevel() == log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	start = time.Now()
	err = cmd.Run()
	elapsed = time.Since(start)
	log.Info("Scale: computation took ", elapsed.Seconds(), " seconds")
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// SetUpScale defines all the settings needed to start SCALE
func SetUpScale(nodeId int, nodeNames, nodesAddrs []string, sm string, scaleCerts [][]byte, certPrivate, certLoc string) error {
	// set up addresses of all MPC nodes
	w, err := os.Create(sm + "/Data/NetworkData.txt")
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("RootCA.crt\n3\n"))
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		iS := strconv.Itoa(i)
		ip1S := strconv.Itoa(i + 1)
		_, err = w.Write([]byte(iS + " " + nodesAddrs[i] + " Player" + ip1S + ".crt " + nodeNames[i] + "\n"))
		if err != nil {
			return err
		}
	}
	_, err = w.Write([]byte("0\n0\n"))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	// set up public certificates of all the MPC nodes
	for i := 0; i < 3; i++ {
		w, err = os.Create(sm + "/Cert-Store/Player" + strconv.Itoa(i+1) + ".crt")
		if err != nil {
			return err
		}
		_, err = w.Write(scaleCerts[i])
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}

	}

	// set up my private key
	o, err := os.Open(certLoc + "/" + certPrivate + ".key")
	buf := make([]byte, 2000)
	n, err := o.Read(buf)
	if err != nil {
		return err
	}
	w, err = os.Create(sm + "/Cert-Store/Player" + strconv.Itoa(nodeId+1) + ".key")
	if err != nil {
		return err
	}
	_, err = w.Write(buf[:n])
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	// set up CA cert
	o, err = os.Open(certLoc + "/RootCA.crt")
	buf = make([]byte, 2000)
	n, err = o.Read(buf)
	if err != nil {
		return err
	}
	w, err = os.Create(sm + "/Cert-Store/RootCA.crt")
	if err != nil {
		return err
	}
	_, err = w.Write(buf[:n])
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}
