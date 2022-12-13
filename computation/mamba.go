package computation

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var programs = []string{"avg", "max", "stats", "linear_regression", "k-means"}
var allowedParams = map[string]bool{"COLS": true, "LEN": true, "NUM_CLUSTERS": true}

func PrepareMambaProgram(nodeId int, funcName string, paramsMap map[string]string, sm string) error {
	var err error
	check := true
	for _, e := range programs {
		if funcName == e {
			check = false
			break
		}
	}
	if check {
		return fmt.Errorf("function not supported")
	}

	check = true
	for key, element := range paramsMap {
		if _, ok := allowedParams[key]; ok {
			_, err = strconv.Atoi(element)
			if err != nil {
				check = false
				break
			}
		} else {
			check = false
			break
		}
	}
	if check == false {
		return fmt.Errorf("parameters not supported")
	}

	// remove previous compiled program if there
	log.Debug("Cleaning files.")
	cmd := exec.Command("rm", "Programs/MPCService/node"+strconv.Itoa(nodeId)+
		"/node"+strconv.Itoa(nodeId)+".mpc")
	cmd.Dir = sm
	_ = cmd.Run()

	// prepare the MAMBA program in the proper folder
	cmd = exec.Command("/bin/sh", "-c", "cp  Programs/MPCService/functions/"+funcName+
		".mpc Programs/MPCService/node"+strconv.Itoa(nodeId)+"/node"+strconv.Itoa(nodeId)+"new.mpc")
	cmd.Dir = sm
	if log.GetLevel() == log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	// write all the parameters to the MAMBA program
	log.Debug("Setting parameters.")
	s := "new"
	for key, val := range paramsMap {
		cmd = exec.Command("/bin/sh", "-c", "sed 's/"+key+"/"+val+
			"/g' Programs/MPCService/node"+
			strconv.Itoa(nodeId)+"/node"+strconv.Itoa(nodeId)+s+".mpc >> Programs/MPCService/node"+
			strconv.Itoa(nodeId)+"/node"+strconv.Itoa(nodeId)+s+"new"+".mpc")
		cmd.Dir = sm
		if log.GetLevel() == log.DebugLevel {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		err = cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command("rm", "Programs/MPCService/node"+strconv.Itoa(nodeId)+
			"/node"+strconv.Itoa(nodeId)+s+".mpc")
		cmd.Dir = sm
		_ = cmd.Run()

		s = s + "new"
	}

	log.Debug("Setting the file.")
	cmd = exec.Command("/bin/sh", "-c", "cp Programs/MPCService/node"+
		strconv.Itoa(nodeId)+"/node"+strconv.Itoa(nodeId)+s+".mpc  Programs/MPCService/node"+
		strconv.Itoa(nodeId)+"/node"+strconv.Itoa(nodeId)+".mpc")
	cmd.Dir = sm
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("rm", "Programs/MPCService/node"+strconv.Itoa(nodeId)+
		"/node"+strconv.Itoa(nodeId)+s+".mpc")
	cmd.Dir = sm
	err = cmd.Run()
	if err != nil {
		return err
	}

	// compile the MAMBA program
	log.Debug("Compiling.")
	cmd = exec.Command("./compile.sh", "Programs/MPCService/node"+strconv.Itoa(nodeId))
	cmd.Dir = sm
	if log.GetLevel() == log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
