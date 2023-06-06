# goAdamnite

goAdamnite is an implementation of the Adamnite Protocol in the GoLang programming language. 

Please look at all the branches if you are an interested contributor to determine where you want to best add value. This repo is open-sourced, and copyrighted under the GoAdamnite Authors, 2022.

Also, be sure to join the Contributor Discord if you are interested in contributing to the project as a whole.

If you are interested in using the private environment to develop a project for the Adamnite Blockchain, please look at the instructions below. 

## Compiling A1 code to ADM bytecode for running on the Adamnite Blockchain

1. Run `/app % go run main.go` to start the Adamnite interactive CLI. Use `help` to list commands available!

## Refurbishing code guidelines and notes
Delete anything that is associated with Ethereum or another blockchain, with the exception of the secp256k1 library. 

Move to modular importing (import metrics and logging from a go library rather than copying it over). Use specific functions in log15 and metrics as neccescary instead of actually importing the whole thing. 

The only exception here are C-libraries (again, crypto secp256k1 and ed25519 libraries), which need cgo and therefore need to be included. Move the main witness definitions in core to consensus; move types to adm (as defined in poc_dev).

Tests only has msgpack testing; we can add sanity testing here if need be but it looks fine for now. 

Main vs App???

Events: lets get rid of that and add into networking, if possible. On that note, add in everything from BARGossip into the updated part of networking. 

Node can stay for now; if we decide to overhaul it, it will be good as a reference implementation. If not, we can remove it and implement it in consensus and networking seperately (not recommeneded, will rather have node use consensus, networking, and adm-storage at once). The new node directory can also contain the validator and storage code from adm. 

ADM: Rework dpos worker in consensus or in the node directory. DPOS worker essentially processes transactions, state transitions, and new blocks. Replace it, as it is a direct copy from an unprofessional repo.
