# goAdamnite

An implementation of the Adamnite Protocol in the GoLang programming language.

## Running A1 smart contract

In case you want to run your A1 smart contract on the Adamnite blockchain, make sure you go through the following steps:

1. Use A1 compiler to compile smart contract to the ADVM bytecode. Follow instructions specified [here](https://github.com/Adamnite/A1).
2. Build the CLI application by running the following commands:

```sh
$ cd VM/cli
$ go build
```

3. Once you have a binary file, run the following command to get code hash, parameters and types:

```sh
$ ./cli debug --from-file <path-to-binary-file>

# if you want to try out our examples, try using hexadecimal content of .a1 files in examples directory
$ ./cli debug --from-hex <hexadecimal-string>
```

4. Using the code hash given by above command, run the CLI again to execute the function:

```sh
$ ./cli execute --from-file <path-to-binary-file> --call-args <args> --gas <gas> --function <code-hash>
```

Congratulations! You have now ran your first smart contract on Adamnite blockchain.
