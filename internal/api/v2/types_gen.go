package api

// GENERATED BY go run ./internal/cmd/genmarshal. DO NOT EDIT.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AccumulateNetwork/accumulate/internal/encoding"
)

type ChainIdQuery struct {
	ChainId []byte `json:"chainId,omitempty" form:"chainId" query:"chainId" validate:"required"`
}

type KeyPage struct {
	Height uint64 `json:"height,omitempty" form:"height" query:"height" validate:"required"`
	Index  uint64 `json:"index,omitempty" form:"index" query:"index" validate:"required"`
}

type MerkleState struct {
	Count uint64   `json:"count,omitempty" form:"count" query:"count" validate:"required"`
	Roots [][]byte `json:"roots,omitempty" form:"roots" query:"roots" validate:"required"`
}

type MetricsQuery struct {
	Metric   string        `json:"metric,omitempty" form:"metric" query:"metric" validate:"required"`
	Duration time.Duration `json:"duration,omitempty" form:"duration" query:"duration" validate:"required"`
}

type MetricsResponse struct {
	Value interface{} `json:"value,omitempty" form:"value" query:"value" validate:"required"`
}

type QueryMultiResponse struct {
	Items []*QueryResponse `json:"items,omitempty" form:"items" query:"items" validate:"required"`
	Start uint64           `json:"start,omitempty" form:"start" query:"start" validate:"required"`
	Count uint64           `json:"count,omitempty" form:"count" query:"count" validate:"required"`
	Total uint64           `json:"total,omitempty" form:"total" query:"total" validate:"required"`
}

type QueryResponse struct {
	Type        string       `json:"type,omitempty" form:"type" query:"type" validate:"required"`
	MerkleState *MerkleState `json:"merkleState,omitempty" form:"merkleState" query:"merkleState" validate:"required"`
	Data        interface{}  `json:"data,omitempty" form:"data" query:"data" validate:"required"`
	Sponsor     string       `json:"sponsor,omitempty" form:"sponsor" query:"sponsor" validate:"required"`
	KeyPage     KeyPage      `json:"keyPage,omitempty" form:"keyPage" query:"keyPage" validate:"required"`
	Txid        []byte       `json:"txid,omitempty" form:"txid" query:"txid" validate:"required"`
	Signer      Signer       `json:"signer,omitempty" form:"signer" query:"signer" validate:"required"`
	Sig         []byte       `json:"sig,omitempty" form:"sig" query:"sig" validate:"required"`
	Status      interface{}  `json:"status,omitempty" form:"status" query:"status" validate:"required"`
}

type DirectoryQueryResult struct {
	Entries         []string         `json:"entries,omitempty" form:"entries" query:"entries" validate:"required"`
	ExpandedEntries []*QueryResponse `json:"expandedEntries,omitempty" form:"expandedEntries" query:"expandedEntries" `
}

type Signer struct {
	PublicKey []byte `json:"publicKey,omitempty" form:"publicKey" query:"publicKey" validate:"required"`
	Nonce     uint64 `json:"nonce,omitempty" form:"nonce" query:"nonce" validate:"required"`
}

type TokenDeposit struct {
	Url    string `json:"url,omitempty" form:"url" query:"url" validate:"required"`
	Amount uint64 `json:"amount,omitempty" form:"amount" query:"amount" validate:"required"`
	Txid   []byte `json:"txid,omitempty" form:"txid" query:"txid" validate:"required"`
}

type TokenSend struct {
	From string         `json:"from,omitempty" form:"from" query:"from" validate:"required"`
	To   []TokenDeposit `json:"to,omitempty" form:"to" query:"to" validate:"required"`
}

type TxIdQuery struct {
	Txid []byte `json:"txid,omitempty" form:"txid" query:"txid" validate:"required"`
}

type TxRequest struct {
	CheckOnly bool        `json:"checkOnly,omitempty" form:"checkOnly" query:"checkOnly"`
	Sponsor   string      `json:"sponsor,omitempty" form:"sponsor" query:"sponsor" validate:"required,acc-url"`
	Signer    Signer      `json:"signer,omitempty" form:"signer" query:"signer" validate:"required"`
	Signature []byte      `json:"signature,omitempty" form:"signature" query:"signature" validate:"required"`
	KeyPage   KeyPage     `json:"keyPage,omitempty" form:"keyPage" query:"keyPage" validate:"required"`
	Payload   interface{} `json:"payload,omitempty" form:"payload" query:"payload" validate:"required"`
}

type TxResponse struct {
	Txid      []byte   `json:"txid,omitempty" form:"txid" query:"txid" validate:"required"`
	Hash      [32]byte `json:"hash,omitempty" form:"hash" query:"hash" validate:"required"`
	Code      uint64   `json:"code,omitempty" form:"code" query:"code" validate:"required"`
	Message   string   `json:"message,omitempty" form:"message" query:"message" validate:"required"`
	Delivered bool     `json:"delivered,omitempty" form:"delivered" query:"delivered" validate:"required"`
}

type UrlQuery struct {
	Url string `json:"url,omitempty" form:"url" query:"url" validate:"required,acc-url"`
}

type QueryPagination struct {
	Start uint64 `json:"start,omitempty" form:"start" query:"start"`
	Count uint64 `json:"count,omitempty" form:"count" query:"count"`
}

type QueryOptions struct {
	ExpandChains bool `json:"expandChains,omitempty" form:"expandChains" query:"expandChains"`
}

func (v *MetricsQuery) BinarySize() int {
	var n int

	n += encoding.StringBinarySize(v.Metric)

	n += encoding.DurationBinarySize(v.Duration)

	return n
}

func (v *MetricsQuery) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.Write(encoding.StringMarshalBinary(v.Metric))

	buffer.Write(encoding.DurationMarshalBinary(v.Duration))

	return buffer.Bytes(), nil
}

func (v *MetricsQuery) UnmarshalBinary(data []byte) error {
	if x, err := encoding.StringUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding Metric: %w", err)
	} else {
		v.Metric = x
	}
	data = data[encoding.StringBinarySize(v.Metric):]

	if x, err := encoding.DurationUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding Duration: %w", err)
	} else {
		v.Duration = x
	}
	data = data[encoding.DurationBinarySize(v.Duration):]

	return nil
}

func (v *ChainIdQuery) MarshalJSON() ([]byte, error) {
	u := struct {
		ChainId *string `json:"chainId,omitempty"`
	}{}
	u.ChainId = encoding.BytesToJSON(v.ChainId)
	return json.Marshal(&u)
}

func (v *MerkleState) MarshalJSON() ([]byte, error) {
	u := struct {
		Count uint64    `json:"count,omitempty"`
		Roots []*string `json:"roots,omitempty"`
	}{}
	u.Count = v.Count
	u.Roots = make([]*string, len(v.Roots))
	for i, x := range v.Roots {
		u.Roots[i] = encoding.BytesToJSON(x)
	}
	return json.Marshal(&u)
}

func (v *MetricsQuery) MarshalJSON() ([]byte, error) {
	u := struct {
		Metric   string      `json:"metric,omitempty"`
		Duration interface{} `json:"duration,omitempty"`
	}{}
	u.Metric = v.Metric
	u.Duration = encoding.DurationToJSON(v.Duration)
	return json.Marshal(&u)
}

func (v *QueryResponse) MarshalJSON() ([]byte, error) {
	u := struct {
		Type        string       `json:"type,omitempty"`
		MerkleState *MerkleState `json:"merkleState,omitempty"`
		Data        interface{}  `json:"data,omitempty"`
		Sponsor     string       `json:"sponsor,omitempty"`
		KeyPage     KeyPage      `json:"keyPage,omitempty"`
		Txid        *string      `json:"txid,omitempty"`
		Signer      Signer       `json:"signer,omitempty"`
		Sig         *string      `json:"sig,omitempty"`
		Status      interface{}  `json:"status,omitempty"`
	}{}
	u.Type = v.Type
	u.MerkleState = v.MerkleState
	u.Data = v.Data
	u.Sponsor = v.Sponsor
	u.KeyPage = v.KeyPage
	u.Txid = encoding.BytesToJSON(v.Txid)
	u.Signer = v.Signer
	u.Sig = encoding.BytesToJSON(v.Sig)
	u.Status = v.Status
	return json.Marshal(&u)
}

func (v *Signer) MarshalJSON() ([]byte, error) {
	u := struct {
		PublicKey *string `json:"publicKey,omitempty"`
		Nonce     uint64  `json:"nonce,omitempty"`
	}{}
	u.PublicKey = encoding.BytesToJSON(v.PublicKey)
	u.Nonce = v.Nonce
	return json.Marshal(&u)
}

func (v *TokenDeposit) MarshalJSON() ([]byte, error) {
	u := struct {
		Url    string  `json:"url,omitempty"`
		Amount uint64  `json:"amount,omitempty"`
		Txid   *string `json:"txid,omitempty"`
	}{}
	u.Url = v.Url
	u.Amount = v.Amount
	u.Txid = encoding.BytesToJSON(v.Txid)
	return json.Marshal(&u)
}

func (v *TxIdQuery) MarshalJSON() ([]byte, error) {
	u := struct {
		Txid *string `json:"txid,omitempty"`
	}{}
	u.Txid = encoding.BytesToJSON(v.Txid)
	return json.Marshal(&u)
}

func (v *TxRequest) MarshalJSON() ([]byte, error) {
	u := struct {
		CheckOnly bool        `json:"checkOnly,omitempty"`
		Sponsor   string      `json:"sponsor,omitempty"`
		Signer    Signer      `json:"signer,omitempty"`
		Signature *string     `json:"signature,omitempty"`
		KeyPage   KeyPage     `json:"keyPage,omitempty"`
		Payload   interface{} `json:"payload,omitempty"`
	}{}
	u.CheckOnly = v.CheckOnly
	u.Sponsor = v.Sponsor
	u.Signer = v.Signer
	u.Signature = encoding.BytesToJSON(v.Signature)
	u.KeyPage = v.KeyPage
	u.Payload = v.Payload
	return json.Marshal(&u)
}

func (v *TxResponse) MarshalJSON() ([]byte, error) {
	u := struct {
		Txid      *string `json:"txid,omitempty"`
		Hash      string  `json:"hash,omitempty"`
		Code      uint64  `json:"code,omitempty"`
		Message   string  `json:"message,omitempty"`
		Delivered bool    `json:"delivered,omitempty"`
	}{}
	u.Txid = encoding.BytesToJSON(v.Txid)
	u.Hash = encoding.ChainToJSON(v.Hash)
	u.Code = v.Code
	u.Message = v.Message
	u.Delivered = v.Delivered
	return json.Marshal(&u)
}

func (v *ChainIdQuery) UnmarshalJSON(data []byte) error {
	u := struct {
		ChainId *string `json:"chainId,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if x, err := encoding.BytesFromJSON(u.ChainId); err != nil {
		return fmt.Errorf("error decoding ChainId: %w", err)
	} else {
		v.ChainId = x
	}
	return nil
}

func (v *MerkleState) UnmarshalJSON(data []byte) error {
	u := struct {
		Count uint64    `json:"count,omitempty"`
		Roots []*string `json:"roots,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Count = u.Count
	v.Roots = make([][]byte, len(u.Roots))
	for i, x := range u.Roots {
		if x, err := encoding.BytesFromJSON(x); err != nil {
			return fmt.Errorf("error decoding Roots[%d]: %w", i, err)
		} else {
			v.Roots[i] = x
		}
	}
	return nil
}

func (v *MetricsQuery) UnmarshalJSON(data []byte) error {
	u := struct {
		Metric   string      `json:"metric,omitempty"`
		Duration interface{} `json:"duration,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Metric = u.Metric
	if x, err := encoding.DurationFromJSON(u.Duration); err != nil {
		return fmt.Errorf("error decoding Duration: %w", err)
	} else {
		v.Duration = x
	}
	return nil
}

func (v *QueryResponse) UnmarshalJSON(data []byte) error {
	u := struct {
		Type        string       `json:"type,omitempty"`
		MerkleState *MerkleState `json:"merkleState,omitempty"`
		Data        interface{}  `json:"data,omitempty"`
		Sponsor     string       `json:"sponsor,omitempty"`
		KeyPage     KeyPage      `json:"keyPage,omitempty"`
		Txid        *string      `json:"txid,omitempty"`
		Signer      Signer       `json:"signer,omitempty"`
		Sig         *string      `json:"sig,omitempty"`
		Status      interface{}  `json:"status,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Type = u.Type
	v.MerkleState = u.MerkleState
	v.Data = u.Data
	v.Sponsor = u.Sponsor
	v.KeyPage = u.KeyPage
	if x, err := encoding.BytesFromJSON(u.Txid); err != nil {
		return fmt.Errorf("error decoding Txid: %w", err)
	} else {
		v.Txid = x
	}
	v.Signer = u.Signer
	if x, err := encoding.BytesFromJSON(u.Sig); err != nil {
		return fmt.Errorf("error decoding Sig: %w", err)
	} else {
		v.Sig = x
	}
	v.Status = u.Status
	return nil
}

func (v *DirectoryQueryResult) UnmarshalJSON(data []byte) error {
	var lenEntries uint64
	if x, err := encoding.UvarintUnmarshalBinary(data); err != nil {
		return fmt.Errorf("error decoding Entries: %w", err)
	} else {
		lenEntries = x
	}
	data = data[encoding.UvarintBinarySize(lenEntries):]

	v.Entries = make([]string, lenEntries)
	for i := range v.Entries {
		if x, err := encoding.StringUnmarshalBinary(data); err != nil {
			return fmt.Errorf("error decoding Entries[%d]: %w", i, err)
		} else {
			v.Entries[i] = x
		}
		data = data[encoding.StringBinarySize(v.Entries[i]):]

	}

	return nil
}

func (v *Signer) UnmarshalJSON(data []byte) error {
	u := struct {
		PublicKey *string `json:"publicKey,omitempty"`
		Nonce     uint64  `json:"nonce,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if x, err := encoding.BytesFromJSON(u.PublicKey); err != nil {
		return fmt.Errorf("error decoding PublicKey: %w", err)
	} else {
		v.PublicKey = x
	}
	v.Nonce = u.Nonce
	return nil
}

func (v *TokenDeposit) UnmarshalJSON(data []byte) error {
	u := struct {
		Url    string  `json:"url,omitempty"`
		Amount uint64  `json:"amount,omitempty"`
		Txid   *string `json:"txid,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Url = u.Url
	v.Amount = u.Amount
	if x, err := encoding.BytesFromJSON(u.Txid); err != nil {
		return fmt.Errorf("error decoding Txid: %w", err)
	} else {
		v.Txid = x
	}
	return nil
}

func (v *TxIdQuery) UnmarshalJSON(data []byte) error {
	u := struct {
		Txid *string `json:"txid,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if x, err := encoding.BytesFromJSON(u.Txid); err != nil {
		return fmt.Errorf("error decoding Txid: %w", err)
	} else {
		v.Txid = x
	}
	return nil
}

func (v *TxRequest) UnmarshalJSON(data []byte) error {
	u := struct {
		CheckOnly bool        `json:"checkOnly,omitempty"`
		Sponsor   string      `json:"sponsor,omitempty"`
		Signer    Signer      `json:"signer,omitempty"`
		Signature *string     `json:"signature,omitempty"`
		KeyPage   KeyPage     `json:"keyPage,omitempty"`
		Payload   interface{} `json:"payload,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.CheckOnly = u.CheckOnly
	v.Sponsor = u.Sponsor
	v.Signer = u.Signer
	if x, err := encoding.BytesFromJSON(u.Signature); err != nil {
		return fmt.Errorf("error decoding Signature: %w", err)
	} else {
		v.Signature = x
	}
	v.KeyPage = u.KeyPage
	v.Payload = u.Payload
	return nil
}

func (v *TxResponse) UnmarshalJSON(data []byte) error {
	u := struct {
		Txid      *string `json:"txid,omitempty"`
		Hash      string  `json:"hash,omitempty"`
		Code      uint64  `json:"code,omitempty"`
		Message   string  `json:"message,omitempty"`
		Delivered bool    `json:"delivered,omitempty"`
	}{}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if x, err := encoding.BytesFromJSON(u.Txid); err != nil {
		return fmt.Errorf("error decoding Txid: %w", err)
	} else {
		v.Txid = x
	}
	if x, err := encoding.ChainFromJSON(u.Hash); err != nil {
		return fmt.Errorf("error decoding Hash: %w", err)
	} else {
		v.Hash = x
	}
	v.Code = u.Code
	v.Message = u.Message
	v.Delivered = u.Delivered
	return nil
}
