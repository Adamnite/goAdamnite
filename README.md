# goAdamnite

goAdamnite is an implementation of the Adamnite Protocol in the GoLang programming language. The current plan is to include an implementation of DPOS, a P2P protocol that defines accounts and transactions, and setting up the overall project for POC_1.

Please look at all the branches if you are an interested contributor to determine where you want to best add value. This repo is open-sourced, and copryrighted under the GoAdamnite Authors, 2021.

Also, be sure to join the Contributor Discord if you are interested in contributing to the POC as a whole.




## Project Structure

A proposed project structure follows. The structure essentially outlines the various directories that will make up the proof of concept. Some of these directories 
have already been finalized for the POC stage, while others are stil in progress or have not been made yet. The core functionality of the Proof of Concept should depend on these directories, which should be formalized as packages. The corresponding directories and their primary functionality are included below:

### GoAdamnite
          {accounts: Contains files related to the account structure and supports the signing/manipulation of messages.
           build: Will contain various code determining the development environment and functions to run the core scripts of the platform.
           gnite: Will contain the major code for the core blockchain engine that will power the distirbuted state machine and the blockchain. 
           common: Will contain common packages such as mathemtical calculations and functions to determine parsing that could be helpful in the core protocol.
           core: Will contain the code for the actual blockchain: this will include definitions for the core blockchain, transactions (both regular and voting),                    DPOS Consensus, a raw database implementing storage in Merkle Trees, and a high-level implementation of the Virtual Machine (VM).
           crypto: Will contain cryptographic protocols and packages that will be used in the core blockchain. This will include cryptography for generating                          signatures and new public addresses. 
           p2p: Will contain the p2p protocols that will determine how nodes are structured and communicate with each other. Ths will include protocols for                       voting and sending messages between nodes.
           serialization: Will conain code for defining various serialization protocols. Serialization is the process of transforming various interfaces into                               bytecode. This directory will also contain deserialization code. 
           }
           
          
           
