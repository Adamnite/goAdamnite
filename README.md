# P2P research and dev
Current plan for this project (likely to update/change in the future) inculdes
 - Build a simple prototype p2p network for involving peer discovery, message broadcasting, peer syncing
 - optimize it as accordingly as seen on other p2p projects
## P2P Sync Algorithm (Subject to improvement)
 1. Reads known peers and bootstrap nodes from config file. At least one bootstrap nodes must be configured.
 2. It queries bootstrap nodes for its known peers list, and then updates its list
 3. It also listens in a port, when a request to get peer list is received, it responds with its known peer list

 ***Things To do: Query known peer for their known peer, limit on known peers ?***
