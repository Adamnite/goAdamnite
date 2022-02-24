## P2P Sync Algorithm (Subject to improvement)
 1. Reads known peers from config file. At least one node (i.e. bootstrap node) must be configured.
 2. It queries bootstrap nodes for its known peers list, and then updates its list
 3. It also listens in a port, when a request to get peer list is received, it responds with its known peer list

 ## Done
 1. two modes implemented **--client-node** and **--full-node**
 2. when in  **--client-node** it does not listen for any connection, it queries its peers for updated known peers list by sending MsgType for PeerRequest. Currently it does not update its own known-peers.json list, its intentional for testing purpose
 3. When in **--full-node**, it just listens for incomming peerSync request, and responds.

 ## TO DO
 1. During PeerSync request add a field or mode in mesg, to indicate that the node requesting is also a full node, so that responder node can add the requesting node as another peer.
 2. Better arguement handling
 3. Better management of code
 4. Once blocks structure are decided, peer request stuff


# DPOS Implementation

## To DO 
1. implement the structure of witness/block validator
2. implement the logic to pick winner based on voting
3. implement the voting logic




### More specific details follow below


The witness, or block validator, should have the ability to both propose and validate blocks. The structure of this has already been implemented in a high level,
but it specifically will be very similar to regular node, but will also have the addiitonal parameters of reputation and the total number of most recent votes.

The conenseus protocol is described below:

Every 200 blocks, voting should occur. All accounts should be allowed to participate in the voting process, and, should also be allowed to declare themselves as an 
election node. An election node is essentially "on the ballot"; they are not witnesses yet, but they can be selected. Nodes who were election nodes in the past are 
automatically witness nodes for the current cycle, unless they choose to opt-out. To be included in the witness pool, a node should have to be voted in every cycle. 
This incentivizes them to behave honestly even after they have been elected. Accounts essentially includes anyone who has access to an
Adamnite wallet. For most accounts, voting should be automatic, though the functionality for an account to do a manual vote should still be available. A vote will 
simply be the allocation of all the NITE in one's account to a specific address, or addresses. For automatic voting, this will automatically be distributed 
to randomly to one of the election nodes in the top 10th percentile of reputation. Examples for both automatic and manual voting follow.

Automatic Voting: Account X's 50 NITE get allocated randomly to Account Y, whose 95 reputation score puts them in the top 10% of all voters.

Manual Voting: Account X decides to allocate their 50 NITE to Account Z, who could be anyone on the voting list.

At the end of the voting process, the accounts with the top 10% of votes should be set as the witness pool. From this pool, 27 validators should be selected in 
a cryptographically-random process (through a Verifable Random Function (VRF)) every 50 blocks. For each block, the block producer should also be chosen randomly 
from the 27 validators weighed by the amount of votes they received and their reputation. 

This is essentially a cyclic process, and will be repeated for the life of the blockchain. For each block, a new block propser should be chosen randomly from the current list of 27 validators. Every 50 blocks, the 27 validators should be chosen randomly from the current pool of witnesses. Every 274 blocks, the witness pool should be updated with a new vote. In this context, "chosen randomly" means chosen using verifable random function (VRF).

Specific low level details concerning the actual implementation (such as data parsing for block proposal) will also have to be implemented. More details on this to
follow. 

An prefixing alrogirithm for both encodoing and decoding (serialization and deserialization) needs to be defined. This can simply be an expansion of the 
serialization.go file in the P2P folder developed by PiecePaper. Definitions will essentially include mappings for converting bytes, strings, arrays, and other data types to structures that can be parsed within the Adamnite Protocol. These rules will need to be defined, and encoded in a separate package that can be imported into the core blockchain and cryptography packages. 
