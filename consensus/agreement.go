package dpos
import (
	"https://github.com/adamnite/go-adamnite/crypto"
	"https://github.com/adamnite/go-adamnite/common"
    "https://github.com/adamnite/go-adamnite/core/types"

)

type AgreementProtocol struct {
    WitnessPool *WitnessPool
    CurrentBlock *Block
    Certificates map[string]Certificate // map of witness addresses to certificates
}

type Certificate struct {
    Signature string
}

func (ap *AgreementProtocol) Start() error {
    // Select a block to agree upon
    block, err := ap.selectBlock()
    if err != nil {
        return fmt.Errorf("error selecting block: %s", err)
    }
    ap.CurrentBlock = block

    // Broadcast block to all witnesses
    err = gossip.Broadcast([]byte(fmt.Sprintf("BLOCK:%s", block)))
    if err != nil {
        return fmt.Errorf("error broadcasting block: %s", err)
    }



	func validate_block(block *Block, witness_state map[string]*WitnessState, tx_messages []*TxMessage) error {
		// Check if all transactions are valid based on the witness's current state view
		for _, tx := range block.Transactions {
			witness := tx.Witness
			witness_state := witness_state[witness]
			if err := validate_transaction(tx, witness_state); err != nil {
				return fmt.Errorf("invalid transaction: %s", err)
			}
		}
	
		// Check if transactions are not conflicting with one another
		for i, tx1 := range block.Transactions {
			for j, tx2 := range block.Transactions {
				if i != j && tx1.Witness == tx2.Witness {
					if err := check_for_conflicts(tx1, tx2); err != nil {
						return fmt.Errorf("conflicting transactions: %s", err)
					}
				}
			}
		}
	
		// Check if transactions are in line with the transaction messages received by the peer
		for _, tx_msg := range tx_messages {
			if !contains(block.Transactions, tx_msg.Transaction) {
				return fmt.Errorf("transaction not found in block: %s", tx_msg.Transaction.Hash())
			}
		}
	
		return nil
	}
	
	func validate_transaction(tx *Transaction, witness_state *WitnessState) error {
		// Check if the transaction is valid based on the witness's current state view
		if err := tx.Validate(witness_state); err != nil {
			return fmt.Errorf("invalid transaction: %s", err)
		}
	
		return nil
	}
	
	func check_for_conflicts(tx1 *Transaction, tx2 *Transaction) error {
		// Check if the two transactions are conflicting with each other
		if tx1.Type == tx2.Type && tx1.Hash() == tx2.Hash() {
			return fmt.Errorf("conflicting transactions: %s", tx1.Hash())
		}
	
		return nil
	}
	
	func contains(transactions []*Transaction, target *Transaction) bool {
		for _, tx := range transactions {
			if tx.Hash() == target.Hash() {
				return true
			}
		}
		return false
	}
	

    // Wait for responses from all witnesses
    responses, err := gossip.BroadcastAndWait([]byte("CERTIFICATE"), len(ap.WitnessPool.AllWitnesses), 10*time.Second)
    if err != nil {
        return fmt.Errorf("error waiting for certificates: %s", err)
    }

    // Process certificates
    for _, response := range responses {
        parts := strings.Split(string(response), ":")
        if len(parts) != 2 || parts[0] != "CERTIFICATE" {
            continue
        }
        addr := parts[1]
        ap.Certificates[addr] = Certificate{Signature: parts[2]}
    }

    // Check if at least 2/3 of witnesses approved the block
    if len(ap.Certificates) < len(ap.WitnessPool.AllWitnesses)*2/3 {
        return fmt.Errorf("not enough witnesses approved the block")
    }

    // Add block to local blockchain
    for _, cert := range ap.Certificates {
        if !verifySignature(cert.Signature, ap.CurrentBlock.Hash()) {
            return fmt.Errorf("invalid signature in certificate")
        }
    }
    err = statedb.ApplyStateTransitions(ap.CurrentBlock.StateTransitions)
    if err != nil {
        return fmt.Errorf("error applying state transitions: %s", err)
    }

    // Broadcast block confirmation to all witnesses
    err = gossip.Broadcast([]byte("BLOCK_CONFIRMED"))
    if err != nil {
        return fmt.Errorf("error broadcasting block confirmation: %s", err)
    }

    return nil
}