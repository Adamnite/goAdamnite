# goAdamnite Beta Testing Instructions

Thanks for offering to beta test the Adamnite Blockchain!

# Running a node

1. You should download the GoAdamnite Repository and run the following commands: cd goAdamnite/cmd/gnite
2. You should run the file appropriate to your operating system: nite.exe is an executable for windows, while nite is an executable for Linux distributions.
3. Running ./nite.exe (.\nite for Linux) will run a node to support the P2P network.
4. Feel free to close this shell for now.

# Accounts
1. To create an account, first run the following command .\nite.exe (.\nite for Linux) --datadir account1 init test.json
2. Then, run the account creation command: .\nite.exe (.\nite if you are on Linux) account new
3. Be sure to enter a password, and to remember the password for future use. DM Arch2230#4689 with your public address on Discord to get some test NITE.
4. Move the private key that was generated (it is likely stored in the AppStore directory) to the keystore directory in the account1 folder.
4. Run the following command to unlock your acccount: .\nite.exe (.\nite if you are on Linux) --datadir account1 --port 30312 --nat extip:"your IPV4 address"  --bootnodes 'gnite://c868aa9d1d79714d82b13baad504877ac7d0404999782f2b915b5588b9322de8ef137f2d225f34431985894f65ea5634332f178c32b51d23e09842e2d078bec9@38.17.51.24:0?discport=30301' --allow-insecure-unlock --unlock "your account's public address"
5. Keep this terminal open; you will need it to send transactions.

# Sending transactions
1. First, make sure you are in the nite-test directory. The nite-test.exe file is an executable for Windows, while the nite-test file is for Linux distributions. Create a new shell for this purpose.
2. To send a transaction, run the following command (keep in mind it is all one line): .\nite-test.exe (.\nite.test if you are on Linux) --sendaddr "the address you want to send coins to" --recaddr "your public address" --amount "the amount you want to send" --keyfile "the directory where you saved your keyfile in the account creation step --password "your password from the previous step"
3. Run nite -h for help and additional commands.


Disclaimer: The goAdamnite software is offerred as a beta test and proof of concept software. It does not constitute an offer to buy, use, or otherwise consume the NITE tokens for any monetary purpose. All nite tokens have no monetary value, and are only provided for the purpose stress-testing the underlying network. The core software will likely change in the coming weeks and months as feedback from testing is implemented into development.

Adamnite Labs and Adamnite Ltd make no gurantees about the future value of the NITE token or its associated network. By using this software, you also release Adamnite Labs and Adamnite Ltd from potential liablities associated with the software. Rewards for running nodes are not guranteed, and will only be given to testers who run nodes throughout the entirety of the testing process and after mainnet launch.





