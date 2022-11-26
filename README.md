# goAdamnite Beta Testing Instructions

Thanks for offering to beta test the Adamnite Blockchain!

# Running a node

1. You should download the GoAdamnite Repository and run the following commands in a terminal or powershell: cd goAdamnite/{the platform you are using} (Please note that we support Windows, MacOs, and Ubuntu)
2. You should then run .\gnite.exe (if you are on Windows), or sudo .\gnite if you are on MacOs or Ubuntu. This will automatically run a node for you. Keep this open; you will need it to connect to the network, create an account and send transactions. Please note that you may need to run the command ```chmod +x gnite``` if you are on MacOs or Ubuntu. 

# Accounts
1. To create an account, run the following command ```.\nite.exe (sudo .\nite if you are on Linux/MacOs) account new```
3. Be sure to enter a password, and to remember the password for future use. DM Tsimafei#6578 or Toucan#5099 on Discord with your public address on Discord to get some test NITE. If you do not get a response within 8 hours, DM Arch2230.

# Sending transactions
1. First, make sure that you are in the correct directory based on your operating system. 
2. Check your balance by running ./nite-test --balance "your public address".
2. To send a transaction, run the following command (keep in mind it is all one line): ```.\nite-test.exe (.\nite.test if you are on Linux or Mac Os) --sendaddr "the address you want to send coins to" --recaddr "your public address" --amount "the amount you want to send"
3. You can also participate in consensus by running the command ```.\nite-test.exe (.\nite.test if you are on Linux or Mac Os) --sendaddr "your address" --recaddr "the account you want to stake your coins to (can be one of the public accounts listed in the beta testing channel)" --amount "the amount you want to send" --txtype true ```
4. Run nite-test -h for help and additional commands.



Disclaimer: The goAdamnite software is offerred as a beta test and proof of concept software. It does not constitute an offer to buy, use, or otherwise consume the NITE tokens for any monetary purpose. All nite tokens have no monetary value, and are only provided for the purpose stress-testing the underlying network. The core software will likely change in the coming weeks and months as feedback from testing is implemented into development.

Adamnite Labs and Adamnite Ltd make no gurantees about the future value of the NITE token or its associated network. By using this software, you also release Adamnite Labs and Adamnite Ltd from potential liablities associated with the software. Rewards for running nodes are not guranteed, and will only be given to testers who run nodes throughout the entirety of the testing process and after mainnet launch.





