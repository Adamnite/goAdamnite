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
          
           
