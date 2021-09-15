package proto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"sort"
	"testing"
)

var Seed = sha256.Sum256([]byte{1, 2, 3})

// GetKey
// Get a private key (where the first 32 bytes is the private key, the second is the public key)
func GetKey() []byte {
	Seed = sha256.Sum256(Seed[:])
	privateKey := ed25519.NewKeyFromSeed(Seed[:])
	return privateKey
}

func TestTokenTransaction(t *testing.T) {
	testurl := "acc://0x411abc253de31674f"
	trans := new(GenTransaction)
	if err := trans.SetRoutingChainID(testurl); err != nil {
		t.Fatal("could not create the Routing value")
	}

	trans.Transaction = []byte("this is a message to who ever is about")
	key := GetKey()
	eSig := new(ED25519Sig)
	eSig.Nonce = 0
	eSig.PublicKey = key[32:]
	transData, err := trans.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if err := eSig.Sign(key, transData); err != nil {
		t.Errorf("error signing tx %v", err)
	}

	trans.Signature = append(trans.Signature, eSig)

	{
		data, err := eSig.Marshal()
		if err != nil {
			t.Error(err)
		}
		s2 := new(ED25519Sig)
		s2.Unmarshal(data)
		if !eSig.Equal(s2) {
			t.Fatal("Can't marshal a signature")
		}
		s2Data, err := s2.Marshal()
		if err != nil {
			t.Error("fail to marshal")
		}
		data = append(data, s2Data...)
		if data == nil {
			t.Fatal("couldn't marshal an ED25519Sig struct")
		}
		s3 := new(ED25519Sig)
		s4 := new(ED25519Sig)
		var err1, err2 error
		data, err1 = s3.Unmarshal(data)
		data, err2 = s4.Unmarshal(data)
		if err1 != nil || err2 != nil {
			t.Errorf("err1: %v err2 %v", err1, err2)
		} else if !s2.Equal(s3) || !s3.Equal(s4) || len(data) != 0 {
			t.Fatal("Can't marshal a multi-signature")
		}
	}

	data, err := trans.Marshal()
	if err != nil {
		t.Error(err)
	}
	to := new(GenTransaction)
	to.UnMarshal(data)

	if !trans.Equal(to) {
		t.Error("should be equal")
	}

	if !to.ValidateSig() {
		t.Error("failed to validate signature")
	}
	to.Routing++
	if to.ValidateSig() {
		t.Error("failed to invalidate signature")
	}
}

const num = 10000
const bvcs = uint64(30)

func TestDistribution(t *testing.T) {
	var routes [256]uint64

	seed := sha256.Sum256([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})

	for i := 0; i < num; i++ {
		privatekey := ed25519.NewKeyFromSeed(seed[:])
		x := privatekey[32:]
		route := uint64(x[0])<<56 | uint64(x[1])<<48 | uint64(x[2])<<40 | uint64(x[3])<<32 |
			uint64(x[4])<<24 | uint64(x[5])<<16 | uint64(x[6])<<8 | uint64(x[7])
		routes[route%bvcs]++
		seed = sha256.Sum256(seed[:])
	}

	sort.Slice(routes[:], func(i, j int) bool {
		return routes[i] > routes[j]
	})
	fmt.Printf(
		"%d iterations mod error on 64 bits is %6.4f%%\n",
		num,
		(float64(routes[0])-float64(routes[bvcs-1]))/
			float64(routes[bvcs-1])*100)

}
