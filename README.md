# goAdamnite

goAdamnite is an implementation of the Adamnite Protocol in the GoLang programming language. The current plan is to include an implementation of DPOS, a P2P protocol that defines accounts and transactions, and setting up the overall project for POC_1.

Please look at all the branches if you are an interested contributor to determine where you want to best add value. This repo is open-sourced, and copyrighted under the GoAdamnite Authors, 2021.

Also, be sure to join the Contributor Discord if you are interested in contributing to the POC as a whole.

## Building A1 code on the Adamnite Blockchain

1. Run `./cli debug --from-file <location of file>` from the CLI folder which should return a response similar to:
`cli % ./cli debug --from-file examples/sum.ao 9703bdb17a160ed80486a83aa3c413c1 ===> i64 (i64, i64)`
 This response shows the hash of the code, parameters, as well as types.

2. Using the returned has, run the cli execute command like so:
`cli % ./cli execute --from-file examples/sum.ao --call-args 0x123,1 --gas 1000000 --function 9703bdb17a160ed80486a83aa3c413c1`

3. To then build `cd` into the core/VM/cli folder and run `go build`

You have now debugged and built your A1 code.

## Project Structure and Specifications

A proposed project structure (focused on mainly on the DPOS consensus mechanism follows. The structure essentially outlines the various directories that will make up the proof of concept. Some of these directories 
have already been finalized for the POC stage, while others are still in progress or have not been made yet. The core functionality of the Proof of Concept should depend on these directories, which should be formalized as packages. For the private beta/POC 2, the following functionality should be made available to anyone with a copy of the software: the ability to get on-boarded to the gossip P2P network via a seed node hosted on AWS, the ability to participate in the consensus process by creating a staking account and then delegating their stake to an eligible address, the ability to get elected within the DPOS consensus process and subsequently having the ability to create/propose blocks while building the reputation. A weighted verifiable random function should be used to select the validators (witnesses) for a round (162 blocks), and then again used to select a new block proposer every 6 blocks within the round. Finally, the POC should anyone with a copy of the software to send a transaction that is finalized on the blockchain.

For contributors, the below structure should serve as a good overview of what packages do what. Note that we only focus on the main directories needed for the core blockchain:

- The "Crypto" package contains code for cryptographic functions that are used in the core blockchain. This includes code for the SECP256K1 curve used for public/private key cryptography, various hashing functions, verifiable random functions, and signature algorithms. 
- The "DPOS" package should contain code for the core consensus mechanism: it should define how the witness pool, the validators, and the actual block proposers that are elected for any given round. It should also define how these proposers propose and add blocks to the blockchain.
- The "Core" package contains code for defining transactions, blocks, witnesses/validators, and other crucial parts of the blockchain. In the future, it will most likely also include code for the VM, and associated state transitions. 
          
           
