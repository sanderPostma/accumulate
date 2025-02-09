package state

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AccumulateNetwork/accumulate/types"
)

type KeyType byte

const (
	KeyTypeUnknown KeyType = iota
	KeyTypeSha256
	KeyTypePublic
	KeyTypeSha256d
	KeyTypeChain
)

func (kt KeyType) String() string {
	switch kt {
	case KeyTypeUnknown:
		return "Unknown"
	case KeyTypeSha256:
		return "SHA-256"
	case KeyTypePublic:
		return "Public"
	case KeyTypeSha256d:
		return "Double-SHA-256"
	case KeyTypeChain:
		return "Chain"
	}
	return fmt.Sprintf("KeyType:%d", kt)
}

type AdiState struct {
	ChainHeader

	KeyType KeyType     `json:"keyType"`
	KeyData types.Bytes `json:"keyData"`
	Nonce   uint64      `json:"nonce"`
}

// NewIdentityState this will eventually be the key groups and potentially just a multi-map of types to chain paths controlled by the identity
func NewIdentityState(adi types.String) *AdiState {
	r := &AdiState{}
	r.SetHeader(adi, types.ChainTypeIdentity)
	return r
}

func NewADI(url types.String, keyType KeyType, keyData []byte) *AdiState {
	r := new(AdiState)
	r.SetHeader(url, types.ChainTypeIdentity)
	r.KeyType = keyType
	r.KeyData = keyData
	return r
}

func (is *AdiState) VerifyAndUpdateNonce(nonce uint64) bool {
	if is.Nonce < nonce {
		is.Nonce = nonce
	}
	is.Nonce = nonce
	return true
}

func (is *AdiState) VerifyKey(key []byte) bool {
	//check if key is a valid public key for identity
	if key[0] == is.KeyData[0] {
		if bytes.Equal(is.KeyData, key) {
			return true
		}
	}

	//check if key is a valid sha256(key) for identity
	kh := sha256.Sum256(key)
	if kh[0] == is.KeyData[0] {
		if bytes.Equal(is.KeyData, kh[:]) {
			return true
		}
	}

	//check if key is a valid sha256d(key) for identity
	kh = sha256.Sum256(kh[:])
	if kh[0] == is.KeyData[0] {
		if bytes.Equal(is.KeyData, kh[:]) {
			return true
		}
	}

	return false
}

//SetKeyData currently key data is defined based upon keyType.  This will
//be replaced once a formal spec for key groups is established.
//we will also be storing references to a key group chain managed by the identity.
//a chain will most likely be just the chain paths mapped to chain types
func (is *AdiState) SetKeyData(keyType KeyType, data []byte) error {
	if len(data) > cap(is.KeyData) {
		is.KeyData = make([]byte, len(data))
	}
	is.KeyType = keyType
	copy(is.KeyData, data)

	return nil
}

//GetKeyData Currently this will just return the key information
//in the future the identity will hold links to a bunch of sub-chains
//managed by the identities.  one of them will be of key groups.
func (is *AdiState) GetKeyData() (KeyType, types.Bytes) {
	return is.KeyType, is.KeyData
}

func (is *AdiState) GetIdentityChainId() types.Bytes {
	h := types.GetIdentityChainFromIdentity(is.ChainUrl.AsString())
	if h == nil {
		return types.Bytes{}
	}
	return h[:]
}

func (is *AdiState) MarshalBinary() ([]byte, error) {

	headerData, err := is.ChainHeader.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal chain header for AdiState, %v", err)
	}

	var buffer bytes.Buffer
	buffer.Write(headerData)
	buffer.WriteByte(byte(is.KeyType))
	data, err := is.KeyData.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal key data for AdiState, %v", err)
	}

	buffer.Write(data)

	var nonce [8]byte
	n := binary.PutUvarint(nonce[:], is.Nonce)
	buffer.Write(nonce[:n])

	return buffer.Bytes(), nil
}

func (is *AdiState) UnmarshalBinary(data []byte) error {

	dLen := len(data)
	if dLen == 0 {
		return fmt.Errorf("cannot unmarshal Identity State, insuffient data")
	}
	i := 0

	err := is.ChainHeader.UnmarshalBinary(data)
	if err != nil {
		return fmt.Errorf("unable to unmarshal data for AdiState, %v", err)
	}

	i += is.ChainHeader.GetHeaderSize()

	is.KeyType = KeyType(data[i])
	i++
	if dLen <= i {
		return fmt.Errorf("cannot unmarshal key type for AdiState, insuffient data")
	}

	err = is.KeyData.UnmarshalBinary(data[i:])
	if err != nil {
		return fmt.Errorf("unable to unmarshal key data for AdiState, %v", err)
	}
	i += is.KeyData.Size(nil)

	if dLen <= i {
		return fmt.Errorf("cannot nonce for AdiState, insuffient data")
	}

	var n int
	is.Nonce, n = binary.Uvarint(data)
	if n <= 0 {
		return fmt.Errorf("error unmarshalling nonce for adi")
	}

	return nil
}

func (k *KeyType) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "public":
		*k = KeyTypePublic
	case str == "sha256":
		*k = KeyTypeSha256
	case str == "sha256d":
		*k = KeyTypeSha256d
	case str == "chain":
		*k = KeyTypeChain
	default:
		*k = KeyTypeSha256
	}

	return nil
}

func (k *KeyType) MarshalJSON() ([]byte, error) {
	var str string
	switch *k {
	case KeyTypePublic:
		str = "public"
	case KeyTypeSha256:
		str = "sha256"
	case KeyTypeSha256d:
		str = "sha256d"
	case KeyTypeChain:
		str = "chain"
	default:
		str = "sha256"
	}

	data, _ := json.Marshal(str)
	return data, nil
}
