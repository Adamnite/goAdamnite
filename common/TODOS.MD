# Immediate TODOs for cleanup

### Move common functionality to utils (Archie)

### Get Chamber A working with genesis block (Marin, Jonah)
See if we can get a better format for getting node IDs as well (those are also copied over from ETH)

Proposed block structure is in genesis.go, under blockchain.

### Get Chamber B working (Jonah)

### Remove all references/dependcies on old code. This includes the old bar gossip structure, storage, ADM, blockchain, RPC. Reimplement transaction lists/mempools if need be. 

### Implement finality and networking additions (Jonah/Marin/Archie)


Additions for things to reimplment coming along soon. Off the top of my head, it will be making sure storage is up to par, redoing anything copied over from node, and removing all dependicies on RLP serialization (use msgpack).