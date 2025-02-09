package synthetic

import (
	"bytes"
	"fmt"

	"github.com/AccumulateNetwork/accumulate/smt/common"
	"github.com/AccumulateNetwork/accumulate/types"
)

// Deprecated: Use protocol.SyntheticCreateChain
type AdiStateCreate struct {
	Header
	PublicKeyHash types.Bytes32 `json:"publicKeyHash" form:"publicKeyHash" query:"publicKeyHash" validate:"required"`
}

// Deprecated: Use protocol.SyntheticCreateChain
func NewAdiStateCreate(txId types.Bytes, from *types.String, to *types.String, keyHash *types.Bytes32) *AdiStateCreate {
	ctas := &AdiStateCreate{}
	ctas.SetHeader(txId, from, to)
	ctas.PublicKeyHash = *keyHash

	return ctas
}

func (a *AdiStateCreate) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	buf.Write(common.Uint64Bytes(types.TxSyntheticIdentityCreate.AsUint64()))
	data, err := a.Header.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("cannot marshal header for Adi State Create message, %v", err)
	}

	buf.Write(data)
	buf.Write(a.PublicKeyHash[:])
	return buf.Bytes(), nil
}

func (a *AdiStateCreate) UnmarshalBinary(data []byte) (err error) {
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("error marshaling Pending Transaction State %v", err)
		}
	}()

	txType, data := common.BytesUint64(data)
	if txType != types.TxSyntheticIdentityCreate.AsUint64() {
		return fmt.Errorf("data is not of a synthetic identity creation type, expected %s, but received %s",
			types.TxSyntheticDepositTokens.Name(), types.TxType(txType).Name())
	}

	err = a.Header.UnmarshalBinary(data)
	if err != nil {
		return fmt.Errorf("insufficient data to unmarshal Adi State Create message header, %v", err)
	}
	i := a.Header.Size()
	if len(data) < i+32 {
		return fmt.Errorf("insufficient data to unmarshal Adi State Create message key hash")
	}
	copy(a.PublicKeyHash[:], data[i:])

	return err
}
