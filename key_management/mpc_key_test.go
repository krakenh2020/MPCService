package key_management_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/krakenh2020/MPCService/key_management"
)

func TestKeyGenLoad(t *testing.T) {
	pubKey, secKey := key_management.GenerateKeypair()
	msg := []byte("blabla")
	encMsg, err := key_management.Encrypt(msg, pubKey)
	assert.NoError(t, err)

	msg2, err := key_management.Decrypt(encMsg, pubKey, secKey)
	assert.NoError(t, err)
	assert.Equal(t, string(msg), string(msg2))

	pubKey, secKey, sig, err := key_management.LoadKeysFromCertKey("keys_certificates", "Ljubljana_node")
	fmt.Println(sig)
	assert.NoError(t, err)

	msg = []byte("blablabla")
	encMsg, err = key_management.Encrypt(msg, pubKey)
	assert.NoError(t, err)

	msg2, err = key_management.Decrypt(encMsg, pubKey, secKey)
	assert.NoError(t, err)
	assert.Equal(t, string(msg), string(msg2))
}
