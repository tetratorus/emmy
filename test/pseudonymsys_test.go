package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/xlab-si/emmy/client"
	"github.com/xlab-si/emmy/config"
	"github.com/xlab-si/emmy/pseudonymsys"
	"testing"
)

// TestPseudonymsys requires a running server (it is started in communication_test.go).
func TestPseudonymsys(t *testing.T) {
	dlog := config.LoadDLog("pseudonymsys")
	caClient, err := client.NewPseudonymsysCAClient(testGrpcServerEndpont)
	if err != nil {
		t.Errorf("Error when initializing NewPseudonymsysCAClient")
	}

	userSecret := config.LoadPseudonymsysUserSecret("user1")

	p, _ := dlog.Exponentiate(dlog.G, userSecret) // this is user's public key
	masterNym := pseudonymsys.NewPseudonym(dlog.G, p)
	caCertificate, err := caClient.ObtainCertificate(userSecret, masterNym)
	if err != nil {
		t.Errorf("Error when registering with CA")
	}

	// usually the endpoint is different from the one used for CA:
	c1, err := client.NewPseudonymsysClient(testGrpcServerEndpont)
	nym1, err := c1.GenerateNym(userSecret, caCertificate)
	if err != nil {
		t.Errorf(err.Error())
	}

	orgName := "org1"
	h1, h2 := config.LoadPseudonymsysOrgPubKeys(orgName)
	orgPubKeys := pseudonymsys.NewOrgPubKeys(h1, h2)
	credential, err := c1.ObtainCredential(userSecret, nym1, orgPubKeys)
	if err != nil {
		t.Errorf(err.Error())
	}

	// register with org2
	// create a client to communicate with org2
	c2, err := client.NewPseudonymsysClient(testGrpcServerEndpont)
	nym2, err := c2.GenerateNym(userSecret, caCertificate)
	if err != nil {
		t.Errorf(err.Error())
	}

	authenticated, err := c2.TransferCredential(orgName, userSecret, nym2, credential)
	if err != nil {
		t.Errorf(err.Error())
	}

	assert.Equal(t, authenticated, true, "Pseudonymsys test failed")
}
