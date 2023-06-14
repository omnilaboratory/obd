package lnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNftOutxo(t *testing.T) {
	r := rpcServer{}
	//c2fe83b53f4eb0b8ba2b4748884727887f840332ef02f3f79b455fcf3a3d11ebi0  aa063ac6d004a02e33b6757d87bf083540d931a115867b5f130c948ea1a6722ci0
	utxo, err := r.getInscriptionUtxo("aa063ac6d004a02e33b6757d87bf083540d931a115867b5f130c948ea1a6722ci0")
	if err != nil {
		t.Log(err)
	}
	t.Log(utxo)
}
func TestGetAllPermissions(t *testing.T) {
	perms := GetAllPermissions()

	// Currently there are there are 16 entity:action pairs in use.
	assert.Equal(t, len(perms), 16)
}
