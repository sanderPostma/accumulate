package state

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/AccumulateNetwork/accumulate/internal/logging"
	"github.com/AccumulateNetwork/accumulate/smt/common"
	"github.com/AccumulateNetwork/accumulate/smt/managed"
	"github.com/AccumulateNetwork/accumulate/smt/storage"
	"github.com/AccumulateNetwork/accumulate/smt/storage/database"
	"github.com/AccumulateNetwork/accumulate/types"
	"math/bits"
	"sort"
	"sync"
	"time"
)

type DBTransaction struct {
	state        *StateDB
	updates      map[types.Bytes32]*blockUpdates
	writes       map[storage.Key][]byte
	transactions transactionLists
}

func (s *StateDB) Begin() *DBTransaction {
	dbTx := &DBTransaction{
		state: s,
	}
	dbTx.updates = make(map[types.Bytes32]*blockUpdates)
	dbTx.writes = map[storage.Key][]byte{}
	dbTx.transactions.reset()
	return dbTx
}

// GetPersistentEntry calls StateDB.GetPersistentEntry(...).
func (tx *DBTransaction) GetPersistentEntry(chainId []byte, verify bool) (*Object, error) {
	return tx.state.GetPersistentEntry(chainId, verify)
}

// GetDB calls StateDB.GetDB().
func (tx *DBTransaction) GetDB() *database.Manager {
	return tx.state.GetDB()
}

// Sync calls StateDB.Sync().
func (tx *DBTransaction) Sync() {
	tx.state.Sync()
}

// RootHash calls StateDB.RootHash().
func (tx *DBTransaction) RootHash() []byte {
	return tx.state.RootHash()
}

// BlockIndex calls StateDB.BlockIndex().
func (tx *DBTransaction) BlockIndex() (int64, error) {
	return tx.state.BlockIndex()
}

// EnsureRootHash calls StateDB.EnsureRootHash().
func (tx *DBTransaction) EnsureRootHash() []byte {
	return tx.state.EnsureRootHash()
}

//AddSynthTx add the synthetic transaction which is mapped to the parent transaction
func (tx *DBTransaction) AddSynthTx(parentTxId types.Bytes, synthTxId types.Bytes, synthTxObject *Object) {
	tx.state.logInfo("AddSynthTx", "txid", logging.AsHex(synthTxId), "entry", logging.AsHex(synthTxObject.Entry))
	var val *[]transactionStateInfo
	var ok bool

	parentHash := parentTxId.AsBytes32()
	if val, ok = tx.transactions.synthTxMap[parentHash]; !ok {
		val = new([]transactionStateInfo)
		tx.transactions.synthTxMap[parentHash] = val
	}
	*val = append(*val, transactionStateInfo{synthTxObject, nil, synthTxId})
}

// AddTransaction queues (pending) transaction signatures and (optionally) an
// accepted transaction for storage to their respective chains.
func (tx *DBTransaction) AddTransaction(chainId *types.Bytes32, txId types.Bytes, txPending, txAccepted *Object) error {
	var txAcceptedEntry []byte
	if txAccepted != nil {
		txAcceptedEntry = txAccepted.Entry
	}
	tx.state.logInfo("AddTransaction", "chainId", logging.AsHex(chainId), "txid", logging.AsHex(txId), "pending", logging.AsHex(txPending.Entry), "accepted", logging.AsHex(txAcceptedEntry))

	chainType, _ := binary.Uvarint(txPending.Entry)
	if types.ChainType(chainType) != types.ChainTypePendingTransaction {
		return fmt.Errorf("expecting pending transaction chain type of %s, but received %s",
			types.ChainTypePendingTransaction.String(), types.TransactionType(chainType).String())
	}

	if txAccepted != nil {
		chainType, _ = binary.Uvarint(txAccepted.Entry)
		if types.ChainType(chainType) != types.ChainTypeTransaction {
			return fmt.Errorf("expecting pending transaction chain type of %s, but received %s",
				types.ChainTypeTransaction.String(), types.ChainType(chainType).String())
		}
	}

	//append the list of pending Tx's, txId's, and validated Tx's.
	tx.state.mutex.Lock()
	defer tx.state.mutex.Unlock()

	tsi := transactionStateInfo{txPending, chainId.Bytes(), txId}
	tx.transactions.pendingTx = append(tx.transactions.pendingTx, &tsi)

	if txAccepted != nil {
		tsi := transactionStateInfo{txAccepted, chainId.Bytes(), txId}
		tx.transactions.validatedTx = append(tx.transactions.validatedTx, &tsi)
	}
	return nil
}

// GetCurrentEntry retrieves the current state object from the database based upon chainId.  Current state either comes
// from a previously saves state for the current block, or it is from the database
func (tx *DBTransaction) GetCurrentEntry(chainId []byte) (*Object, error) {
	if chainId == nil {
		return nil, fmt.Errorf("chain id is invalid, thus unable to retrieve current entry")
	}
	var (
		ret        *Object
		err        error
		chainIdKey types.Bytes32
	)
	copy(chainIdKey[:32], chainId[:32])

	tx.state.mutex.Lock()
	currentState := tx.updates[chainIdKey]
	tx.state.mutex.Unlock()
	if currentState != nil {
		ret = currentState.stateData
	} else {
		currentState := blockUpdates{}
		currentState.bucket = bucketEntry
		//pull current state entry from the database.
		currentState.stateData, err = tx.GetPersistentEntry(chainId, false)
		if err != nil {
			return nil, err
		}
		//if we have valid data, store off the state
		ret = currentState.stateData
	}

	return ret, nil
}

// AddStateEntry append the entry to the chain, the subChainId is if the chain upon which
// the transaction is against touches another chain. One example would be an account type chain
// may change the state of the sigspecgroup chain (i.e. a sub/secondary chain) based on the effect
// of a transaction.  The entry is the state object associated with
func (tx *DBTransaction) AddStateEntry(chainId *types.Bytes32, txHash *types.Bytes32, object *Object) {
	tx.state.logInfo("AddStateEntry", "chainId", logging.AsHex(chainId), "txHash", logging.AsHex(txHash), "entry", logging.AsHex(object.Entry))
	begin := time.Now()

	tx.state.TimeBucket = tx.state.TimeBucket + float64(time.Since(begin))*float64(time.Nanosecond)*1e-9

	tx.state.mutex.Lock()
	updates := tx.updates[*chainId]
	tx.state.mutex.Unlock()

	if updates == nil {
		updates = new(blockUpdates)
		tx.updates[*chainId] = updates
	}

	updates.txId = append(updates.txId, txHash)
	updates.stateData = object
}

func (tx *DBTransaction) writeTxs(mutex *sync.Mutex, group *sync.WaitGroup) error {
	defer group.Done()

	err := tx.writeValidatedTxs(mutex)
	if err != nil {
		return err
	}

	tx.writePendingTxs(mutex)

	//clear out the transactions after they have been processed
	tx.transactions.validatedTx = nil
	tx.transactions.pendingTx = nil
	tx.transactions.synthTxMap = make(map[types.Bytes32]*[]transactionStateInfo)
	return nil
}

func (tx *DBTransaction) writeValidatedTxs(mutex *sync.Mutex) error {
	for _, txn := range tx.transactions.validatedTx {
		data, _ := txn.Object.MarshalBinary()
		//store the transaction

		txHash := txn.TxId.AsBytes32()
		if synthTxInfos, ok := tx.transactions.synthTxMap[txHash]; ok {
			var synthData []byte
			for _, synthTxInfo := range *synthTxInfos {
				synthData = append(synthData, synthTxInfo.TxId...)
				synthTxData, err := synthTxInfo.Object.MarshalBinary()
				if err != nil {
					return err
				}

				tx.state.dbMgr.Key(bucketStagedSynthTx, "", synthTxInfo.TxId).PutBatch(synthTxData)

				//store the hash of th synthObject in the bptMgr, will be removed after synth tx is processed
				tx.state.bptMgr.Bpt.Insert(synthTxInfo.TxId.AsBytes32(), sha256.Sum256(synthTxData))
			}
			//store a list of txid to list of synth txid's
			tx.state.dbMgr.Key(bucketTxToSynthTx, txn.TxId).PutBatch(synthData)
		}

		mutex.Lock()
		//store the transaction in the transaction bucket by txid
		tx.state.dbMgr.Key(bucketTx, txn.TxId).PutBatch(data)
		//insert the hash of the tx object in the BPT
		tx.state.bptMgr.Bpt.Insert(txHash, sha256.Sum256(data))
		mutex.Unlock()
	}
	return nil
}

func (tx *DBTransaction) writePendingTxs(mutex *sync.Mutex) {
	for _, txn := range tx.transactions.pendingTx {
		//marshal the pending transaction state
		data, _ := txn.Object.MarshalBinary()
		//hash it and add to the merkle state for the pending chain
		pendingHash := sha256.Sum256(data)

		mutex.Lock()
		//Store the mapping of the Transaction hash to the pending transaction hash which can be used for
		// validation so we can find the pending transaction
		tx.state.dbMgr.Key("MainToPending", txn.TxId).PutBatch(pendingHash[:])

		//store the pending transaction by the pending tx hash
		tx.state.dbMgr.Key(bucketPendingTx, pendingHash[:]).PutBatch(data)
		mutex.Unlock()
	}
}

// Commit will push the data to the database and update the patricia trie
func (tx *DBTransaction) Commit(blockHeight int64, timestamp time.Time) ([]byte, error) {
	//build a list of keys from the map
	currentStateCount := len(tx.updates)
	if currentStateCount == 0 {
		//only attempt to record the block if we have any data.
		return tx.RootHash(), nil
	}

	group := new(sync.WaitGroup)
	group.Add(1)
	group.Add(len(tx.updates))

	mutex := new(sync.Mutex)
	//to try the multi-threading add "go" in front of the next line
	err := tx.writeTxs(mutex, group)
	if err != nil {
		return nil, err
	}

	orderedUpdates := orderUpdateList(tx.updates)
	err = tx.commitUpdates(orderedUpdates, err, group, mutex)
	if err != nil {
		return nil, err
	}
	tx.commitTxWrites()

	// Don't write the anchor during the genesis TX
	err = tx.writeAnchors(mutex, blockHeight, timestamp, orderedUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to make anchor: %v", err)
	}

	group.Wait()

	tx.state.bptMgr.Bpt.Update()

	//reset out block update buffer to get ready for the next round
	tx.state.sync.Add(1)
	//to enable threaded batch writes, put go in front of next line.
	tx.writeBatches()

	tx.updates = make(map[types.Bytes32]*blockUpdates)

	//return the state of the BPT for the state of the block
	rh := types.Bytes(tx.state.RootHash()).AsBytes32()
	tx.state.logInfo("WriteStates", "height", blockHeight, "hash", logging.AsHex(rh))
	return tx.state.RootHash(), nil
}

func (tx *DBTransaction) commitUpdates(orderedUpdates []types.Bytes32, err error, group *sync.WaitGroup, mutex *sync.Mutex) error {
	for _, chainId := range orderedUpdates {
		err = tx.writeChainState(group, mutex, tx.state.merkleMgr, chainId)
		if err != nil {
			return err
		}

		//TODO: figure out how to do this with new way state is derived
		//if len(currentState.pendingTx) != 0 {
		//	mdRoot := v.PendingChain.MS.GetMDRoot()
		//	if mdRoot == nil {
		//		//shouldn't get here, but will reject if I do
		//		panic(fmt.Sprintf("shouldn't get here on writeState() on chain id %X obtaining merkle state", chainId))
		//	}
		//	//todo:  Determine how we purge pending tx's after 2 weeks.
		//	s.bptMgr.Bpt.Insert(chainId, *mdRoot)
		//}
	}
	return nil
}

func (tx *DBTransaction) commitTxWrites() {
	// Process pending writes
	writeOrder := make([]storage.Key, 0, len(tx.writes))
	for k := range tx.writes {
		writeOrder = append(writeOrder, k)
	}
	sort.Slice(writeOrder, func(i, j int) bool {
		return bytes.Compare(writeOrder[i][:], writeOrder[j][:]) < 0
	})
	for _, k := range writeOrder {
		tx.state.GetDB().Key(k).PutBatch(tx.writes[k])
	}
	// The compiler optimizes this into a constant-time operation
	for k := range tx.writes {
		delete(tx.writes, k)
	}
}

func orderUpdateList(updateList map[types.Bytes32]*blockUpdates) []types.Bytes32 {
	// Create an ordered list of chain IDs that need updating. The iteration
	// order of maps in Go is random. Randomly ordering database writes is bad,
	// because that leads to consensus errors between nodes, since each node
	// will have a different random order. So we need updates to have some
	// consistent order, regardless of what it is.
	updateOrder := make([]types.Bytes32, 0, len(updateList))
	for id := range updateList {
		updateOrder = append(updateOrder, id)
	}
	sort.Slice(updateOrder, func(i, j int) bool {
		return bytes.Compare(updateOrder[i][:], updateOrder[j][:]) < 0
	})
	return updateOrder
}

func (tx *DBTransaction) writeChainState(group *sync.WaitGroup, mutex *sync.Mutex, mm *managed.MerkleManager, chainId types.Bytes32) error {
	defer group.Done()

	err := tx.state.merkleMgr.SetChainID(chainId[:])
	if err != nil {
		return err
	}

	// We get ChainState objects here, instead. And THAT will hold
	//       the MerkleStateManager for the chain.
	//mutex.Lock()
	currentState := tx.updates[chainId]
	//mutex.Unlock()

	if currentState == nil {
		panic(fmt.Sprintf("Chain state is nil meaning no updates were stored on chain %X for the block. Should not get here!", chainId[:]))
	}

	//add all the transaction states that occurred during this block for this chain (in order of appearance)
	for _, txn := range currentState.txId {
		//store the txHash for the chains, they will be mapped back to the above recorded tx's
		tx.state.logInfo("AddHash", "hash", logging.AsHex(tx))
		mm.AddHash(managed.Hash((*txn)[:]))
	}

	if currentState.stateData != nil {
		//store the state of the main chain in the state object
		count := uint64(mm.MS.Count)
		currentState.stateData.Height = count
		currentState.stateData.Roots = make([][]byte, 64-bits.LeadingZeros64(count))
		for i := range currentState.stateData.Roots {
			if count&(1<<i) == 0 {
				// Only store the hashes we need
				continue
			}
			currentState.stateData.Roots[i] = mm.MS.Pending[i].Copy()
		}

		//now store the state object
		chainStateObject, err := currentState.stateData.MarshalBinary()
		if err != nil {
			panic("failed to marshal binary for state data")
		}

		mutex.Lock()
		tx.GetDB().Key(bucketEntry, chainId.Bytes()).PutBatch(chainStateObject)
		// The bptMgr stores the hash of the ChainState object hash.
		tx.state.bptMgr.Bpt.Insert(chainId, sha256.Sum256(chainStateObject))
		mutex.Unlock()
	}
	//TODO: figure out how to do this with new way state is derived
	//if len(currentState.pendingTx) != 0 {
	//	mdRoot := v.PendingChain.MS.GetMDRoot()
	//	if mdRoot == nil {
	//		//shouldn't get here, but will reject if I do
	//		panic(fmt.Sprintf("shouldn't get here on writeState() on chain id %X obtaining merkle state", chainId))
	//	}
	//	//todo:  Determine how we purge pending tx's after 2 weeks.
	//	s.bptMgr.Bpt.Insert(chainId, *mdRoot)
	//}

	return nil
}

func (tx *DBTransaction) writeAnchors(mutex *sync.Mutex, blockIndex int64, timestamp time.Time, chainsThatUpdated []types.Bytes32) error {
	// Load the previous anchor chain head
	prevHead, err := tx.state.getAnchorHead()
	if errors.Is(err, storage.ErrNotFound) {
		prevHead = &AnchorMetadata{Index: -1}
	} else if err != nil {
		return err
	}

	// Make sure the block index is increasing
	if prevHead.Index >= blockIndex {
		panic(fmt.Errorf("Current height is %d but the next block height is %d!", prevHead.Index, blockIndex))
	}

	// Metadata
	head := new(AnchorMetadata)
	head.Index = blockIndex
	head.PreviousHeight = tx.state.merkleMgr.MS.Count
	head.Timestamp = timestamp
	head.Chains = make([][32]byte, len(chainsThatUpdated))

	// Add an anchor for each updated chain to the anchor chain
	for i, chainId := range chainsThatUpdated {
		head.Chains[i] = chainId

		err := tx.state.merkleMgr.SetChainID(chainId[:])
		if err != nil {
			return err
		}
		root := tx.state.merkleMgr.MS.GetMDRoot()

		err = tx.state.merkleMgr.SetChainID([]byte(bucketMinorAnchorChain))
		if err != nil {
			return err
		}
		tx.state.merkleMgr.AddHash(root)
	}

	data, err := head.MarshalBinary()
	if err != nil {
		return err
	}

	// Add the anchor head to the anchor chain
	tx.state.merkleMgr.AddHash(data)

	// Index the anchor chain against the block index
	tx.GetDB().Key(bucketMinorAnchorChain, "Index", blockIndex).PutBatch(common.Int64Bytes(tx.state.merkleMgr.MS.Count))

	// Update the Patricia tree
	var id [32]byte
	copy(id[:], []byte(bucketMinorAnchorChain.String()))
	tx.state.bptMgr.Bpt.Insert(id, sha256.Sum256(data))
	return nil
}

func (tx *DBTransaction) writeBatches() {
	defer tx.state.sync.Done()
	tx.state.dbMgr.EndBatch()
	tx.state.bptMgr.DBManager.EndBatch()
}
