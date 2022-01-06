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