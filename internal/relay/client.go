package relay

import (
	"context"
	"fmt"

	"github.com/tendermint/tendermint/libs/service"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Client interface {
	service.Service
	client.ABCIClient
	client.EventsClient

	// From client.SignClient
	Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error)
}

type Batchable interface {
	Client
	NewBatch() Batch
}

type Batch interface {
	client.ABCIClient
	Send(context.Context) ([]interface{}, error)
	Count() int
	Clear() int
}

type rpcClient struct {
	*http.HTTP
}

func (c rpcClient) NewBatch() Batch {
	return c.HTTP.NewBatch()
}

type TransactionInfo struct {
	ReferenceId []byte //tendermint reference id
	NetworkId   int    //network the transaction will be submitted to
	QueueIndex  int    //the transaction's place in line
}

type DispatchStatus struct {
	NetworkId int
	Returns   []interface{}
	Err       error
}

type BatchedStatus struct {
	Status []DispatchStatus
}

func (bs *BatchedStatus) ResolveTransactionResponse(ti TransactionInfo) (*ctypes.ResultBroadcastTx, error) {
	for _, s := range bs.Status {
		if ti.NetworkId == s.NetworkId {
			if len(s.Returns) < ti.QueueIndex {
				return nil, fmt.Errorf("invalid queue length for batch dispatch response, unable to find transaction")
			}
			r, ok := s.Returns[ti.QueueIndex].(*ctypes.ResultBroadcastTx)
			if !ok {
				return nil, fmt.Errorf("unable to resolve return interface as ctypes.ResultBroadcastTx")
			}
			return r, s.Err
		}
	}
	return nil, fmt.Errorf("transaction response not found from batch dispatch")
}
