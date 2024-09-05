package base

import (
	"encoding/hex"

	"github.com/OLProtocol/ordx/common"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
)

func (b *BaseIndexer) fetchBlock(height int) *common.Block {
	hash, err := getBlockHash(uint64(height))
	if err != nil {
		common.Log.Errorf("getBlockHash %d failed. %v", height, err)
		return nil
		//common.Log.Fatalln(err)
	}

	rawBlock, err := getRawBlock(hash)
	if err != nil {
		common.Log.Errorf("getRawBlock %d %s failed. %v", height, hash, err)
		return nil
		//common.Log.Fatalln(err)
	}
	blockData, err := hex.DecodeString(rawBlock)
	if err != nil {
		common.Log.Errorf("DecodeString %d %s failed. %v", height, rawBlock, err)
		return nil
		//common.Log.Panicf("BaseIndexer.fetchBlock-> Failed to decode block: %v", err)
	}

	// Deserialize the bytes into a btcutil.Block.
	block, err := btcutil.NewBlockFromBytes(blockData)
	if err != nil {
		common.Log.Errorf("NewBlockFromBytes %d failed. %v", height, err)
		return nil
		//common.Log.Panicf("BaseIndexer.fetchBlock-> Failed to parse block: %v", err)
	}

	transactions := block.Transactions()
	txs := make([]*common.Transaction, len(transactions))
	for i, tx := range transactions {
		inputs := []*common.Input{}
		outputs := []*common.Output{}

		for _, v := range tx.MsgTx().TxIn {
			txid := v.PreviousOutPoint.Hash.String()
			vout := v.PreviousOutPoint.Index
			input := &common.Input{Txid: txid, Vout: int64(vout), Witness: v.Witness}
			inputs = append(inputs, input)
		}

		// parse the raw tx values
		for j, v := range tx.MsgTx().TxOut {
			// Determine the type of the script and extract the address
			scyptClass, addrs, _, err := txscript.ExtractPkScriptAddrs(v.PkScript, b.chaincfgParam)
			if err != nil {
				common.Log.Errorf("ExtractPkScriptAddrs %d failed. %v", height, err)
				return nil
				//common.Log.Panicf("BaseIndexer.fetchBlock-> Failed to extract address: %v", err)
			}

			addrsString := make([]string, len(addrs))
			for i, x := range addrs {
				addrsString[i] = x.EncodeAddress()
			}

			var receiver common.ScriptPubKey

			if len(addrs) == 0 {
				address := "UNKNOWN"
				if scyptClass == txscript.NullDataTy {
					address = "OP_RETURN"
				}
				receiver = common.ScriptPubKey{
					Addresses: []string{address},
					Type:      scyptClass,
				}
			} else {
				receiver = common.ScriptPubKey{
					Addresses: addrsString,
					Type:      scyptClass,
				}
			}

			output := &common.Output{Height: height, TxId: i, Value: v.Value, Address: &receiver, N: int64(j)}
			outputs = append(outputs, output)
		}

		txs[i] = &common.Transaction{
			Txid:    tx.Hash().String(),
			Inputs:  inputs,
			Outputs: outputs,
		}
	}

	t := block.MsgBlock().Header.Timestamp
	bl := &common.Block{
		Timestamp:     t,
		Height:        height,
		Hash:          block.Hash().String(),
		PrevBlockHash: block.MsgBlock().Header.PrevBlock.String(),
		Transactions:  txs,
	}

	return bl
}

// Prefetches blocks from bitcoind and sends them to the blocksChan
func (b *BaseIndexer) spawnBlockFetcher(startHeigh int, endHeight int, stopChan chan struct{}) {
	currentHeight := startHeigh
	for currentHeight <= endHeight {
		select {
		case <-stopChan:
			return
		default:
			block := b.fetchBlock(currentHeight)
			b.blocksChan <- block
			currentHeight += 1
		}
	}

	<-stopChan
}

func (b *BaseIndexer) drainBlocksChan() {
	for {
		select {
		case <-b.blocksChan:
		default:
			return
		}
	}
}
