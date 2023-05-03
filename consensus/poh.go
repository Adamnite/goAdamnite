package dpos

import(
    "https://github.com/adamnite/go-adamnite/crypto"
	"https://github.com/adamnite/go-adamnite/common"
    "https://github.com/adamnite/go-adamnite/core/types"
)

type StateTransition struct {
    ContractAddress string
    CallerAddress   string
    Input           []byte
}


func applyStateTransitions(transitions []*StateTransition, POHTable map[int][]byte) error {
    for _, transition := range transitions {
        // Apply state transition
        switch transition.Type {
        case ContractCreation:
            // Create new contract instance
            contract := &Contract{
                Code:    transition.Code,
                Storage: make(map[string][]byte),
            }
            err := contractDB.Write(contract.ID, contract)
            if err != nil {
                return fmt.Errorf("error writing contract to database: %s", err)
            }
        case ContractCall:
            // Retrieve contract instance from database
            contract := &Contract{}
            err := contractDB.Read(transition.ContractID, contract)
            if err != nil {
                return fmt.Errorf("error reading contract from database: %s", err)
            }
            // Execute contract call
            output, err := executeContractCall(contract, transition.Function, transition.Arguments)
            if err != nil {
                return fmt.Errorf("error executing contract call: %s", err)
            }
            // Update contract storage with output
            contract.Storage[transition.OutputVariable] = output
            err = contractDB.Write(contract.ID, contract)
            if err != nil {
                return fmt.Errorf("error writing contract to database: %s", err)
            }
        default:
            return fmt.Errorf("unknown state transition type")
        }
        // Create new POH record for this state transition
        pohData := []byte(fmt.Sprintf("%d:%x", transition.Timestamp, sha256.Sum256([]byte(transition.String()))))
        POHTable[transition.Timestamp] = pohData
    }
    return nil
}



func (ap *AgreementProtocol) processStateTransitions(transitions []*StateTransition) error {
    // Generate POH table and final state
    pohTable := make(map[int]Hash)
    currentState := ap.CurrentBlock.LastState
    for i, t := range transitions {
        // Apply state transition
        err := applyStateTransition(currentState, t)
        if err != nil {
            return fmt.Errorf("error applying state transition %d: %s", i, err)
        }
        // Generate hash and update POH table
        hash := generateHash(currentState)
        pohTable[i] = hash
        currentState = hash
    }

    // Split state transitions into packets
    packets := splitIntoPackets(transitions)

    // Send packets to witnesses
    for _, packet := range packets {
        // Select witness
        witness, err := ap.WitnessPool.SelectWitness()
        if err != nil {
            return fmt.Errorf("error selecting witness: %s", err)
        }
        // Send packet and POH table to witness
        err = gossip.SendTo(witness.Address, encodePacketAndPOH(packet, pohTable))
        if err != nil {
            return fmt.Errorf("error sending packet to witness %s: %s", witness.Address, err)
        }
    }

    // Wait for certificates from witnesses
    certificates := make(map[string]Certificate)
    for i := 0; i < len(packets); i++ {
        response, err := gossip.ReceiveWait(time.Second)
        if err != nil {
            return fmt.Errorf("error waiting for certificate: %s", err)
        }
        parts := strings.Split(string(response), ":")
        if len(parts) != 2 || parts[0] != "CERTIFICATE" {
            return fmt.Errorf("invalid response from witness: %s", response)
        }
        addr := parts[1]
        certificates[addr] = Certificate{Signature: parts[2]}
    }

    // Check if all witnesses approved the state transitions
    if len(certificates) < len(ap.WitnessPool.AllWitnesses) {
        return fmt.Errorf("not all witnesses approved the state transitions")
    }

    // Create signed batch
    batch := &SignedBatch{
        Certificates: certificates,
        StateTransitions: transitions,
    }
    err := signBatch(batch)
    if err != nil {
        return fmt.Errorf("error signing batch: %s", err)
    }

    // Broadcast batch to all witnesses
    err = gossip.Broadcast(encodeBatch(batch))
    if err != nil {
        return fmt.Errorf("error broadcasting batch: %s", err)
    }

    return nil
}



func splitIntoPackets(transitions []*StateTransition) []*Packet {
    // Sort state transitions based on their dependencies
    sortedTransitions := sortBasedOnDependencies(transitions)
    
    // Initialize packet list
    packets := []*Packet{}
    
    // Initialize current packet
    currentPacket := &Packet{
        Hashes:     []string{},
        Transitions: []*StateTransition{},
    }
    
    // Loop through sorted transitions and add them to packets
    for _, transition := range sortedTransitions {
        // Check if adding the transition will cause the packet to exceed the maximum packet size
        if exceedsPacketSize(currentPacket, transition) {
            // Add current packet to packet list
            packets = append(packets, currentPacket)
            
            // Create new packet
            currentPacket = &Packet{
                Hashes:     []string{},
                Transitions: []*StateTransition{},
            }
        }
        
        // Add transition to current packet
        currentPacket.Transitions = append(currentPacket.Transitions, transition)
        
        // Add POH hash to current packet
        currentPacket.Hashes = append(currentPacket.Hashes, POH_table[transition])
    }
    
    // Add last packet to packet list
    packets = append(packets, currentPacket)
    
    return packets
}



type transitionNode struct {
    transition *StateTransition
    dependencies map[string]bool // map of transition hashes that this node depends on
}

func sortBasedOnDependencies(transitions []*StateTransition) ([]*StateTransition, error) {
    // Build a graph of the transitions and their dependencies
    graph := make(map[string]*transitionNode)
    for _, t := range transitions {
        node := &transitionNode{
            transition: t,
            dependencies: make(map[string]bool),
        }
        for _, dep := range t.Dependencies {
            depHash := dep.Hash()
            if _, ok := graph[depHash]; !ok {
                return nil, fmt.Errorf("dependency not found: %s", depHash)
            }
            node.dependencies[depHash] = true
        }
        graph[t.Hash()] = node
    }

    // Perform a topological sort to order the transitions
    sortedTransitions := make([]*StateTransition, 0, len(transitions))
    for len(graph) > 0 {
        // Find a node that has no remaining dependencies
        var nextNode *transitionNode
        for _, node := range graph {
            if len(node.dependencies) == 0 {
                nextNode = node
                break
            }
        }
        if nextNode == nil {
            return nil, fmt.Errorf("circular dependency detected")
        }

        // Remove the node and its dependencies from the graph
        delete(graph, nextNode.transition.Hash())
        for _, node := range graph {
            delete(node.dependencies, nextNode.transition.Hash())
        }

        // Add the transition to the sorted list
        sortedTransitions = append(sortedTransitions, nextNode.transition)
    }

    return sortedTransitions, nil
}

func exceedsPacketSize(packet *Packet, transition *StateTransition) bool {
    // TODO: Implement check if adding the transition will cause the packet to exceed the maximum packet size
}