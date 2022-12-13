package key_management

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/nacl/box"
)

func GenerateKeypair() (publicKey, privateKey []byte) {
	publicKeyTmp, privateKeyTmp, err := box.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	publicKey = publicKeyTmp[:]
	privateKey = privateKeyTmp[:]
	return
}

func KeysFromCertKey(cerKey []byte, key *rsa.PrivateKey) ([]byte, []byte, []byte, error) {
	hash := sha256.New()
	_, err := hash.Write(cerKey)
	if err != nil {
		return nil, nil, nil, err
	}

	privateKey := hash.Sum(nil)
	reader := bytes.NewReader(privateKey)
	publicKeyTmp, _, err := box.GenerateKey(reader)

	publicKey := publicKeyTmp[:]
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, publicKey)

	return publicKey, privateKey, sig, nil
}

func LoadKeysFromCertKey(certFolder, name string) ([]byte, []byte, []byte, error) {
	certKey, err := ioutil.ReadFile(certFolder + "/" + name + ".key")
	if err != nil {
		return nil, nil, nil, err
	}
	cert, err := tls.LoadX509KeyPair(certFolder+"/"+name+".crt", certFolder+"/"+name+".key")
	privateKey := cert.PrivateKey.(*rsa.PrivateKey)

	return KeysFromCertKey(certKey, privateKey)
}

func Encrypt(input, pubkey []byte) ([]byte, error) {
	var key [32]byte
	copy(key[:], pubkey)
	encrypted, err := box.SealAnonymous(nil, input, &key, nil)
	return encrypted, err
}

func Decrypt(message, inputPublicKey, inputPrivateKey []byte) ([]byte, error) {
	var publicKey [32]byte
	var privateKey [32]byte
	copy(publicKey[:], inputPublicKey)
	copy(privateKey[:], inputPrivateKey)
	out, ok := box.OpenAnonymous(nil, message, &publicKey, &privateKey)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}
	return out, nil
}

func LoadCertificate(name, loc string) ([]byte, error) {
	f1, err := os.Open(loc + "/" + name + ".crt")
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 2000)
	n, err := f1.Read(buf)

	return buf[:n], err
}
