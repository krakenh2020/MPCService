package data_management

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/krakenh2020/MPCService/computation"
	"github.com/krakenh2020/MPCService/key_management"
)

// MPCPrime is the prime used in the chosen Shamir secret sharing protocol used in SCALE-MAMBA
var MPCPrime, _ = new(big.Int).SetString("340282366920938463463374607431768211507", 10)
var MPCPrimeHalf = new(big.Int).Div(MPCPrime, big.NewInt(2))

func NewUniformRandomVector(n int, max *big.Int) ([]*big.Int, error) {
	v := make([]*big.Int, n)
	var err error
	for i, _ := range v {
		v[i], err = rand.Int(rand.Reader, max)
		if err != nil {
			return nil, err
		}
	}
	return v, err
}

// CreateSharesShamir is a helping function that splits a vector
// input into 3 random parts x_1, x_2, x_3, such that
// f(i) = x_i and f(0) = x, for a linear f
func CreateSharesShamir(input []*big.Int) ([][]*big.Int, error) {
	// f(i) = ai + x
	a, err := NewUniformRandomVector(len(input), MPCPrime)
	if err != nil {
		return nil, err
	}

	res := make([][]*big.Int, 3)
	for i := int64(0); i < 3; i++ {
		res[i] = make([]*big.Int, len(input))
		// linear function going through input[j]
		for j := 0; j < len(input); j++ {
			res[i][j] = new(big.Int).Mul(a[j], big.NewInt(i+1))
			val := new(big.Int).Set(input[j])
			if new(big.Int).Abs(val).Cmp(MPCPrimeHalf) > 0 {
				return nil, fmt.Errorf("error: input value too big")
			}
			// in case input is negative
			if val.Sign() < 0 {
				val.Add(MPCPrime, val)
			}
			res[i][j].Add(val, res[i][j])
			res[i][j].Mod(res[i][j], MPCPrime)
		}
	}

	return res, nil
}

func JoinSharesShamir(input [][]*big.Int) ([]*big.Int, error) {
	res := make([]*big.Int, len(input[0]))

	for i, _ := range input[0] {
		res[i] = new(big.Int).Mul(input[0][i], big.NewInt(2))
		res[i].Sub(res[i], input[1][i])
		res[i].Mod(res[i], MPCPrime)

		check := new(big.Int).Mul(input[1][i], big.NewInt(3))
		check.Sub(check, new(big.Int).Mul(input[2][i], big.NewInt(2)))
		check.Mod(check, MPCPrime)
		if check.Cmp(res[i]) != 0 {
			return nil, fmt.Errorf("joining faild, inconsistent shares")
		}

		check2 := new(big.Int).Mul(input[0][i], big.NewInt(3))
		check2.Sub(check2, input[2][i])
		check2.Mod(check2, MPCPrime)
		twiceRI := new(big.Int).Mul(res[i], big.NewInt(2))
		twiceRI.Mod(twiceRI, MPCPrime)
		if check2.Cmp(twiceRI) != 0 {
			return nil, fmt.Errorf("joining faild, inconsistent shares")
		}

		if res[i].Cmp(MPCPrimeHalf) > 0 {
			res[i].Sub(res[i], MPCPrime)
		}
	}

	return res, nil
}

func JoinSharesShamirFloat(input [][]*big.Int) []float64 {
	res := make([]float64, len(input[0]))

	for i, _ := range input[0] {
		f := new(big.Int).Mul(input[0][i], big.NewInt(2))
		f.Sub(f, input[1][i])
		f.Mod(f, MPCPrime)
		if f.Cmp(MPCPrimeHalf) > 0 {
			f.Sub(f, MPCPrime)
		}
		res[i] = FixIntToFloat(f.Int64())
	}

	return res
}

type VecEnc struct {
	Key []byte
	Iv  []byte
	Val []byte
}

func EncryptVec(input []*big.Int, pubKey []byte) (string, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	// encrypt vec
	keyEncBytes, err := key_management.Encrypt(inputBytes, pubKey)
	if err != nil {
		return "", err
	}
	keyEnc := base64.StdEncoding.EncodeToString(keyEncBytes)

	return keyEnc, nil
}

func DecVec(encVec string, pubKey, secKey []byte) ([]*big.Int, error) {
	encVecBytes, err := base64.StdEncoding.DecodeString(encVec)
	if err != nil {
		return nil, err
	}

	dec, err := key_management.Decrypt(encVecBytes, pubKey, secKey)
	if err != nil {
		return nil, err
	}

	var res []*big.Int
	err = json.Unmarshal(dec, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func CsvToVec(file string) ([]*big.Int, []string, []float64, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, nil, err
	}

	countLines := 0
	vec := make([]*big.Int, 0)
	vecFloat := make([]float64, 0)
	scan := bufio.NewScanner(f)
	var cols []string
	for scan.Scan() {
		countLines++
		text := scan.Text()
		if countLines == 1 {
			cols = strings.Split(text, ",")
			continue
		}

		vals := strings.Split(text, ",")

		for _, e := range vals {
			f, err := strconv.ParseFloat(e, 64)
			if err != nil {
				return nil, nil, nil, err
			}
			vecFloat = append(vecFloat, f)
			i, err := FloatToFixInt(f)
			if err != nil {
				return nil, nil, nil, err
			}
			val := new(big.Int).SetInt64(i)

			vec = append(vec, val)
		}
	}
	if err = scan.Err(); err != nil {
		return nil, nil, nil, err
	}
	err = f.Close()

	return vec, cols, vecFloat, nil
}

func CsvTxtToVec(csvTxt string) ([]*big.Int, []string, error) {
	lines := strings.Split(csvTxt, "\n")

	countLines := 0
	vec := make([]*big.Int, 0)
	var cols []string
	for _, e := range lines {
		countLines++
		if countLines == 1 {
			cols = strings.Split(e, ",")
			continue
		}
		if e == "" {
			continue
		}

		vals := strings.Split(e, ",")

		for _, e := range vals {
			f, err := strconv.ParseFloat(e, 64)
			if err != nil {
				return nil, nil, err
			}
			i, err := FloatToFixInt(f)
			if err != nil {
				return nil, nil, err
			}
			val := new(big.Int).SetInt64(i)

			vec = append(vec, val)
		}
	}

	return vec, cols, nil
}

func SplitCsvFile(file, output string, pubKeys [][]byte) ([]float64, [][]*big.Int, []string, error) {
	vec, cols, vecFloat, err := CsvToVec(file)
	if err != nil {
		return nil, nil, nil, err
	}

	shares, err := CreateSharesShamir(vec)
	if err != nil {
		return nil, nil, nil, err
	}

	w, err := os.Create(output)
	if err != nil {
		return nil, nil, nil, err
	}

	for i := int64(0); i < 3; i++ {
		msg, err := EncryptVec(shares[i], pubKeys[i])
		if err != nil {
			return nil, nil, nil, err
		}

		_, err = w.Write([]byte(msg))
		if err != nil {
			return nil, nil, nil, err
		}
		_, err = w.Write([]byte("\n"))
		if err != nil {
			return nil, nil, nil, err
		}
	}
	_, err = w.Write([]byte(strings.Join(cols, ",") + "\n"))
	if err != nil {
		return nil, nil, nil, err
	}
	err = w.Close()

	return vecFloat, shares, cols, err
}

func DownloadShare(fileLink, filePath string) error {
	cmdStr := "wget --output-document=" + filePath + " --no-check-certificate \"" + fileLink + "\""
	cmd := exec.Command("bash", "-c", cmdStr)
	out := new(bytes.Buffer)
	outErr := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Stderr = outErr
	err := cmd.Run()
	log.Debug(out)
	log.Debug(outErr)

	return err
}

func DeleteShare(filePath string) error {
	cmdStr := "rm " + filePath
	cmd := exec.Command("bash", "-c", cmdStr)
	out := new(bytes.Buffer)
	outErr := new(bytes.Buffer)
	cmd.Stdout = out
	cmd.Stderr = outErr

	err := cmd.Run()

	log.Debug(out)
	log.Debug(outErr)

	return err
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func ReadShare(file string, pubKey, secKey []byte, nodeId int) ([]*big.Int, []string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}

	reader := bufio.NewReader(f)

	countLines := 0
	var decVec []*big.Int
	for countLines < 3 {
		text, err := Readln(reader)

		if countLines != nodeId {
			countLines++
			continue
		}

		decVec, err = DecVec(text, pubKey, secKey)
		if err != nil {
			return nil, nil, err
		}
		countLines++
	}
	// columns info
	text, err := Readln(reader)
	if err != nil {
		return nil, nil, err
	}

	cols := strings.Split(text, ",")

	err = f.Close()

	return decVec, cols, err
}

func ReduceToCols(input []*big.Int, colsAll []string, val string) ([]*big.Int, []string, error) {
	cols := strings.Split(val, ",")
	colsMap := make(map[string]bool)
	for _, e := range cols {
		colsMap[e] = true
	}
	inputNew := make([]*big.Int, len(input)*len(cols)/len(colsAll))
	count := 0
	for i, e := range input {
		index := i % len(colsAll)
		if _, ok := colsMap[colsAll[index]]; ok {
			inputNew[count] = e
			count++
		}
	}

	colsNew := make([]string, len(cols))
	count = 0
	for _, e := range colsAll {
		if _, ok := colsMap[e]; ok {
			colsNew[count] = e
			count++
		}
	}

	return inputNew, colsNew, nil
}

func PrepareData(inputsLinks []string, inputVecs []string, inputCols [][]string, nodeId int, sm string, params map[string]string, pubKey, secKey []byte) (int, int, int, []string, string) {
	// download and read
	allInputs := make([]*big.Int, 0)

	var cols []string
	var input []*big.Int
	var err error
	for _, link := range inputsLinks {
		err = DownloadShare(link, "mpc_data"+strconv.Itoa(nodeId)+".txt")
		if err != nil {
			e := "error, computation failed, downloading data error "
			log.Error(e, err)
			return 0, 0, 0, nil, e
		}
		log.Info("Engine: Downloaded data from ", link)

		input, cols, err = ReadShare("mpc_data"+strconv.Itoa(nodeId)+".txt", pubKey, secKey, nodeId)
		if err != nil || len(input) == 0 {
			e := "error, computation failed, input error "
			log.Error("error, computation failed, input error ", err)
			return 0, 0, 0, nil, e
		}
		// clean from memory
		err = DeleteShare("mpc_data" + strconv.Itoa(nodeId) + ".txt")
		if err != nil {
			log.Error("computation failed, deleting data error ", err)
		}
		// reduce the input to specified columns
		if val, ok := params["cols"]; ok {
			inputNew, colsNew, err := ReduceToCols(input, cols, val)
			if err != nil {
				e := "error, computation failed, columns error "
				log.Error(e, err)
				return 0, 0, 0, nil, e
			}
			input = inputNew
			cols = colsNew
		}

		allInputs = append(allInputs, input...)
	}

	for i, encText := range inputVecs {
		input, err = DecVec(encText, pubKey, secKey)
		if err != nil {
			e := "error, computation failed, decrypting input "
			log.Error(e, err)
			return 0, 0, 0, nil, e
		}
		cols = inputCols[i]

		// reduce the input to specified columns
		if val, ok := params["cols"]; ok {
			inputNew, colsNew, err := ReduceToCols(input, cols, val)
			if err != nil {
				e := "error, computation failed, columns error "
				log.Error(e, err)
				return 0, 0, 0, nil, e
			}
			input = inputNew
			cols = colsNew
		}

		allInputs = append(allInputs, input...)
	}

	log.Info("MPC engine: data size: ", len(allInputs)/len(cols), " rows ", len(cols), " columns.")

	// prepare data for Scale
	err = computation.InputPrepare(nodeId, allInputs, nil, sm)
	if err != nil {
		e := "error, computation failed, input error "
		log.Error(e, err)
		return 0, 0, 0, nil, e
	}

	return len(inputsLinks), len(cols), len(allInputs), cols, ""
}

func ResultsToCsvText(vec []float64, cols []string, funcName string) (string, error) {
	text := ""
	switch funcName {
	case "stats":
		numLines := 4
		if len(cols)*numLines != len(vec) {
			return "", fmt.Errorf("vector length error")
		}

		values := []string{"average", "standard deviation", "min", "max"}

		firstLine := append([]string{""}, cols...)
		firstLineString := strings.Join(firstLine, ",")

		text = text + firstLineString + "\r\n"

		for i := 0; i < numLines; i++ {
			line := []string{values[i]}
			for j := i; j < len(vec); j = j + numLines {
				s := fmt.Sprint(vec[j])
				line = append(line, s)
			}

			lineString := strings.Join(line, ",")
			text = text + lineString + "\r\n"
		}
	case "max":
		numLines := 1
		if len(cols)*numLines != len(vec) {
			return "", fmt.Errorf("vector length error")
		}

		values := []string{"max value"}

		firstLine := append([]string{""}, cols...)
		firstLineString := strings.Join(firstLine, ",")

		text = text + firstLineString + "\r\n"

		for i := 0; i < numLines; i++ {
			line := []string{values[i]}
			for j := i; j < len(vec); j = j + numLines {
				s := fmt.Sprint(vec[j])
				line = append(line, s)
			}

			lineString := strings.Join(line, ",")
			text = text + lineString + "\r\n"
		}
	case "avg":
		numLines := 1
		if len(cols)*numLines != len(vec) {
			return "", fmt.Errorf("vector length error")
		}

		values := []string{"average value"}

		firstLine := append([]string{""}, cols...)
		firstLineString := strings.Join(firstLine, ",")

		text = text + firstLineString + "\r\n"

		for i := 0; i < numLines; i++ {
			line := []string{values[i]}
			for j := i; j < len(vec); j = j + numLines {
				s := fmt.Sprint(vec[j])
				line = append(line, s)
			}

			lineString := strings.Join(line, ",")
			text = text + lineString + "\r\n"
		}
	case "linear_regression":
		if len(cols) != len(vec)+1 {
			return "", fmt.Errorf("vector length error")
		}

		firstLine := append([]string{""}, cols[:len(cols)-1]...)
		firstLineString := strings.Join(firstLine, ",")

		text = text + firstLineString + "\r\n"

		line := []string{"coefficients of linear regression predicting " + cols[len(cols)-1]}
		for j := 0; j < len(vec); j++ {
			s := fmt.Sprint(vec[j])
			line = append(line, s)
		}

		lineString := strings.Join(line, ",")
		text = text + lineString + "\r\n"

	case "k-means":
		numLines := len(vec) / (len(cols) + 1)
		if len(vec)%(len(cols)+1) != 0 {
			return "", fmt.Errorf("vector length error")
		}

		firstLine := append([]string{"", "size", ""}, cols...)
		firstLineString := strings.Join(firstLine, ",")

		text = text + firstLineString + "\r\n"

		for i := 0; i < numLines; i++ {
			line := []string{"cluster " + strconv.Itoa(i+1), fmt.Sprint(vec[len(vec)-numLines+i]), "cluster center"}
			for j := i * (len(cols)); j < (i+1)*(len(cols)); j = j + 1 {
				s := fmt.Sprint(vec[j])
				line = append(line, s)
			}

			lineString := strings.Join(line, ",")
			text = text + lineString + "\r\n"
		}
	default:
		return "", fmt.Errorf("function not suported")
	}

	return text, nil
}
