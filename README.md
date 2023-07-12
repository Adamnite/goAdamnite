# goAdamnite

goAdamnite is an implementation of the Adamnite Protocol in the GoLang programming language. 

Please look at all the branches if you are an interested contributor to determine where you want to best add value. This repo is open-sourced, and copyrighted under the GoAdamnite Authors, 2022.

Also, be sure to join the Contributor Discord if you are interested in contributing to the project as a whole.

If you are interested in using the private environment to develop a project for the Adamnite Blockchain, please look at the instructions below. 

## Compiling A1 code to ADM bytecode for running on the Adamnite Blockchain

1. Run `./cli debug --from-file <location of file>` from the VM/CLI folder which should return a response similar to:
`cli % ./cli debug --from-file examples/sum.ao 9703bdb17a160ed80486a83aa3c413c1 ===> i64 (i64, i64)`
 This response shows the hash of the code, parameters, as well as types.

2. Using the returned hash, run the cli execute command like so:
`cli % ./cli execute --from-file examples/sum.ao --call-args 0x123,1 --gas 1000000 --function 9703bdb17a160ed80486a83aa3c413c1`

3. To then build `cd` into the core/VM/cli folder and run `go build`

You have now compiled and ran your smart contract in a private Adamnite environment.
       
           
