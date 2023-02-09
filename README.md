goAdamnite is an implementation of the Adamnite Protocol in the GoLang programming language. As of Janurary 2023, the goAdamnite client implementation supports transactions, participating in consensus by staking your tokens, and creating smart contracts that correspond/interface with a virtual machine based on WebAssembly. 

The goal for the project from late Janurary to the end of March is to remove technical debt, and restructure various parts of the repository for cleanliness/code ettiquette. A proposed project structure, along with the type of files they should contain, follows. Note that this is just a recommendation: the goal is to make a majority of the code (with the exception of liscensed imports from open source platforms) original. The current branch (poc_dev) has been reformatted to include the blank directories mentioned below (the original directories still remain). Some of the directories have specifications that further describe the type of files they should have. If you are interested in helping with this process, please ping the respecitve assignees/leads in the Adamnite Discord Channel. If you can't reach them, ping Arch2230 or Toucan.





## Project Structure
### Consensus 
This directory should contain all files pertaining to consensus and agreement. Specifically, it should implement witness pools, the selection of both core chain witnesses and database witnesses, and the agreement process that defines how witnesses reach consensus. More information can be found in the Consensus Directory. 
Current assignees/leads: Jonah and SDMG.

### Gossip 
This directory should contain the files currently in bargossip. The code should be rewritten/refactored to be made original: any leftover code from the Ethereum Protocol should be deleted. Proof of Misbehavior and node reputation for blacklisting/whitelisting should be improved and implemented properly, and checked to ensure that they follow the protocol described here. Efficiency improvements to decrease latency, if time permits, should also be implemented. 
Assignee/Lead: Tsiamfei

### Wallet
The wallet directory should contain the files needed to manage and create an account. It ideally replaces the account directory, and also contains some of the files from core that define the creation of a new account. Again, double-checking to ensure that the code is completely original and meets the requirements for protocol parameters is the goal. 
Lead/Assignee: Tsiamfei

### ADM
This directory will retain the same functionality it does right now. It should contain an original implementation of the binary trie (something that is much cleaner than the current implementation lifted from Ethereum's EIP), along with original code for other state database functions. Original implementations of the LevelDB and MemoryDB (we can follow a different structure if needed as well to simplify things) should also be implemented. Ideally, refactoring this will involve getting rid of all legacy code, and making code more human readable/friendly. 
Lead/Assignees: SDMG, Tsiamfei will support

### Crypto
This directory should contain unique implementations (with the exception of the elliptic curve cryptography library used by all blockchains and cryptocurrencies) of the various cryptographic functions that Adamnite uses. This includes the Verifiable Random Function (VRF), SHA-512, RIPMED-160, etc. 
Assignees: Pyro and Arch, Tsiamfei will support

### Blockchain
This directory should contain code that defines transactions and blocks, thus defining the actual ledger/blockchain that nodes will read from and store locally. This should also contain state transition information that pertains to smart contract data and storage, read from the offchain database. 
Lead/Assignee: Tsiamfei

### CMD
Command Line Functions. Already implemented. Focus should be on making display text and other information more orignal. 
Lead/Assignees: SDMG

### Utils
Common things (math functions, database APIs, etc) that we need. Again, focus on originality.
Lead/Assignee: Pyro and SDMG

### VM
Same as right now. A better environment that allows for easier access to both the VM and database could be implemented if time permits.
Lead/Assignee: SDMG

Note that these are the main directories. Additional directories (such as those suppporting parameters and nodes) should also be implemented. The goal here is to get rid of external code and make everything original/clean in anticipation of a full open-source release at the end of March. The primary focus should be on orignality and cleaning up. 


