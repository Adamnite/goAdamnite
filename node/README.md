# Directory for core node software
The core node software for Adamnite should be developed here. The node software should utilize the gossip protocol to exchange messages relevant to consensus, store transaction mempools (pending transactions), and run the Adamnite consensus protocol. 

The node directory should include an implementation of a full node, a regular node that can catchup from a full node, and a seed node. Individual nodes should be able to send transactions, get information about other transactions, register for consensus, and delegate their stake to other accounts who have registered. 



## Node Specification
The node should act as a middle layer between the pure networking library, and the consensus mechanism. That is, nodes should use the networking library to communicate and reach agreement in the context of the consensus protocol. A full node within the Adamnite protocol should store incoming transaction messages in a pool, store newly committed blocks, and run the consensus protocol. A node should be initalized with a specified data directory, and should have an active worker designated to create blocks in the case that it is selected as a witness. Each of these commands should be implemented as a service within the node.

A node should be able to use the networking protocol to brodcast messsages to other nodes, track votes, and other consensus related messages. Note that the consensus protocol should only be used as a pure agreement service (implemented seperately). The node should store data, mostly implemented in the adm directory. This data should be stored according to the storage protocol. 



