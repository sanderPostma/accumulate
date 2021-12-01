package state

import (
	"errors"
	"fmt"
	"github.com/AccumulateNetwork/accumulate/smt/managed"
	"github.com/AccumulateNetwork/accumulate/smt/pmt"
	"github.com/AccumulateNetwork/accumulate/smt/storage"
	"github.com/AccumulateNetwork/accumulate/smt/storage/database"
	"github.com/AccumulateNetwork/accumulate/types"
	"github.com/tendermint/tendermint/libs/log"
	"sync"
)

type transactionStateInfo struct {
	Object  *Object
	ChainId types.Bytes
	TxId    types.Bytes
}

type transactionLists struct {
	validatedTx []*transactionStateInfo //list of validated transaction chain state objects for block
	pendingTx   []*transactionStateInfo //list of pending transaction chain state objects for block
	synthTxMap  map[types.Bytes32]*[]transactionStateInfo
}

// reset will (re)initialize the transaction lists, this should be done on startup and at the end of each block
func (t *transactionLists) reset() {
	t.pendingTx = nil
	t.validatedTx = nil
	t.synthTxMap = make(map[types.Bytes32]*[]transactionStateInfo)
}

type bucket string

const (
	bucketEntry            = bucket("StateEntries")
	bucketTx               = bucket("Transactions")
	bucketMainToPending    = bucket("MainToPending") //main TXID to PendingTXID
	bucketPendingTx        = bucket("PendingTx")     //Store pending transaction
	bucketStagedSynthTx    = bucket("StagedSynthTx") //store the staged synthetic transactions
	bucketTxToSynthTx      = bucket("TxToSynthTx")   //TXID to synthetic TXID
	bucketMinorAnchorChain = bucket("MinorAnchorChain")

	markPower = int64(8)
)

//bucket SynthTx stores a list of synth tx's derived from a tx

func (b bucket) String() string { return string(b) }
func (b bucket) Bytes() []byte  { return []byte(b) }

type blockUpdates struct {
	bucket    bucket
	txId      []*types.Bytes32
	stateData *Object //the latest chain state object modified from a tx
}

// StateDB the state DB will only retrieve information out of the database.  To store stuff use PersistentStateDB instead
type StateDB struct {
	dbMgr      *database.Manager
	merkleMgr  *managed.MerkleManager
	debug      bool
	bptMgr     *pmt.Manager //pbt is the global patricia trie for the application
	TimeBucket float64
	mutex      sync.Mutex
	sync       sync.WaitGroup
	logger     log.Logger
}

func (s *StateDB) init(debug bool) error {
	s.debug = debug

	s.bptMgr = pmt.NewBPTManager(s.dbMgr)
	_, err := managed.NewMerkleManager(s.dbMgr, markPower)
	return err
}

// Open database to manage the smt and chain states
func (s *StateDB) Open(dbFilename string, debug bool) (err error) {

	s.dbMgr, err = database.NewDBManager("badger", dbFilename)
	if err != nil {
		return err
	}

	return s.initMerkleMgr(err, debug)
}

func (s *StateDB) OpenInMemory(debug bool) (err error) {

	s.dbMgr, err = database.NewDBManager("memory", "memory")
	if err != nil {
		return err
	}

	return s.initMerkleMgr(err, debug)
}

func (s *StateDB) initMerkleMgr(err error, debug bool) error {
	s.merkleMgr, err = managed.NewMerkleManager(s.dbMgr, markPower)
	if err != nil {
		return err
	}

	err = s.init(debug)
	if err != nil {
		return err
	}
	return nil
}

func (s *StateDB) Load(db storage.KeyValueDB, debug bool) (err error) {
	s.dbMgr = new(database.Manager)
	s.dbMgr.InitWithDB(db)
	s.merkleMgr, err = managed.NewMerkleManager(s.dbMgr, markPower)
	if err != nil {
		return err
	}

	err = s.init(debug)
	if err != nil {
		return err
	}
	return nil
}

func (s *StateDB) GetDB() *database.Manager {
	return s.dbMgr
}

func (s *StateDB) Sync() {
	s.sync.Wait()
}

//GetTxRange get the transaction id's in a given range
func (s *StateDB) GetTxRange(chainId *types.Bytes32, start int64, end int64) (resultHashes []types.Bytes32, maxAvailable int64, err error) {
	defer s.mutex.Unlock()

	s.mutex.Lock()
	hashes, err := s.merkleMgr.GetRange(chainId[:], start, end)
	if err != nil {
		return nil, 0, err
	}
	err = s.merkleMgr.SetChainID(chainId[:])
	if err != nil {
		return nil, 0, err
	}
	maxAvailable = s.merkleMgr.GetElementCount()
	for i := range hashes {
		resultHashes = append(resultHashes, hashes[i].Bytes32())
	}

	return resultHashes, maxAvailable, nil
}

//GetTx get the transaction by transaction ID
func (s *StateDB) GetTx(txId []byte) (tx []byte, err error) {
	tx, err = s.dbMgr.Key(bucketTx, txId).Get()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

//GetPendingTx get the pending transactions by primary transaction ID
func (s *StateDB) GetPendingTx(txId []byte) (pendingTx []byte, err error) {

	pendingTxId, err := s.dbMgr.Key(bucketMainToPending, txId).Get()
	if err != nil {
		return nil, err
	}
	pendingTx, err = s.dbMgr.Key(bucketPendingTx, pendingTxId).Get()
	if err != nil {
		return nil, err
	}

	return pendingTx, nil
}

// GetSyntheticTxIds get the transaction id list by the transaction ID that spawned the synthetic transactions
func (s *StateDB) GetSyntheticTxIds(txId []byte) (syntheticTxIds []byte, err error) {

	syntheticTxIds, err = s.dbMgr.Key(bucketTxToSynthTx, txId).Get()
	if err != nil {
		//this is not a significant error. Synthetic transactions don't usually have other synth tx's.
		//TODO: Fixme, this isn't an error
		return nil, err
	}

	return syntheticTxIds, nil
}

//GetPersistentEntry will pull the data from the database for the StateEntries bucket.
func (s *StateDB) GetPersistentEntry(chainId []byte, verify bool) (*Object, error) {
	_ = verify
	s.Sync()

	if s.dbMgr == nil {
		return nil, fmt.Errorf("database has not been initialized")
	}

	data, err := s.dbMgr.Key("StateEntries", chainId).Get()
	if errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%w: no state defined for %X", storage.ErrNotFound, chainId)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get state entry %X: %v", chainId, err)
	}

	ret := &Object{}
	err = ret.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state for %x", chainId)
	}
	//if verify {
	//todo: generate and verify data to make sure the state matches what is in the patricia trie
	//}
	return ret, nil
}

// GetTransaction loads the state of the given transaction.
func (s *StateDB) GetTransaction(txid []byte) (*Object, error) {
	s.Sync()

	if s.dbMgr == nil {
		return nil, fmt.Errorf("database has not been initialized")
	}

	data, err := s.dbMgr.Key(bucketTx, txid).Get()
	if errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%w: no transaction defined for %X", storage.ErrNotFound, txid)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction %X: %v", txid, err)
	}

	ret := &Object{}
	err = ret.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state for %x", txid)
	}

	return ret, nil
}

func (s *StateDB) getAnchorHead() (*AnchorMetadata, error) {
	err := s.merkleMgr.SetChainID([]byte(bucketMinorAnchorChain))
	if err != nil {
		return nil, err
	}

	if s.merkleMgr.MS.Count == 0 {
		return nil, storage.ErrNotFound
	}

	data, err := s.merkleMgr.Get(s.merkleMgr.MS.Count - 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read anchor chain element %d", s.merkleMgr.MS.Count-1)
	}

	head := new(AnchorMetadata)
	err = head.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return head, nil
}
func (s *StateDB) BlockIndex() (int64, error) {
	head, err := s.getAnchorHead()
	if err != nil {
		return 0, err
	}

	return head.Index, nil
}

func (s *StateDB) RootHash() []byte {
	h := s.bptMgr.Bpt.Root.Hash // Make a copy
	return h[:]                 // Return a reference to the copy
}

func (s *StateDB) EnsureRootHash() []byte {
	s.bptMgr.Bpt.EnsureRootHash()
	return s.RootHash()
}

func (s *StateDB) SetLogger(logger log.Logger) {
	if logger != nil {
		logger = logger.With("module", "dbMgr")
	}
	s.logger = logger
}

func (s *StateDB) logInfo(msg string, keyVals ...interface{}) {
	if s.logger != nil {
		// TODO Maybe this should be Debug?
		s.logger.Info(msg, keyVals...)
	}
}
