package launcher

import (
	"fmt"

	"github.com/adamnite/go-adamnite/accounts/keystore"
	"github.com/adamnite/go-adamnite/internal/utils"
	"github.com/urfave/cli/v2"
)

var (
	accountCommand = cli.Command{
		Name:     "account",
		Usage:    "Manage accounts",
		Category: "ACCOUNT COMMANDS",
		Description: `
		
Manage accounts, list all existing accounts, import a private key into a new account,
create a new account or update an existing account.Aliases: 

It supports interactive mode, when you are prompted for password as well as non-interactive 
mode where passwords are supplied via a given password file.

Make sure you remember the password you gave when creating a new account (with either or import).

Keys are stored under <DATADIR>/keystore.
Make sure you backup your keys regularly.`,
		Subcommands: []*cli.Command{
			{
				Name:   "new",
				Usage:  "Create a new account",
				Action: utils.MigrateFlags(accountNew),
				Flags:  []cli.Flag{},
				Description: `
				
gnite account new

Creates a new account and prints the address.`,
			},
		},
	}
)

func accountNew(ctx *cli.Context) error {
	cfg := adamConfig{Node: defaultNodeConfig()}

	if file := ctx.String(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	utils.SetNodeConfig(ctx, &cfg.Node)
	scryptN, scryptP, keydir, err := cfg.Node.AccountConfig()

	if err != nil {
		utils.Fatalf("Failed to read configuration: %v", err)
	}

	password := utils.GetPassPhraseWithList("Please give a password.", true, 0, utils.MakePasswordList(ctx))

	account, err := keystore.StoreKey(keydir, password, scryptN, scryptP)
	if err != nil {
		utils.Fatalf("Failed to create account: %v", err)
	}

	fmt.Printf("\nYour new key was generated\n\n")
	fmt.Printf("Public address of the key:   %s\n", account.Address.Hex())
	fmt.Printf("Path of the secret key file: %s\n\n", account.URL.Path)
	fmt.Printf("- You can share your public address with anyone. Others need it to interact with you.\n")
	fmt.Printf("- You must NEVER share the secret key with anyone! The key controls access to your funds!\n")
	fmt.Printf("- You must BACKUP your key file! Without the key, it's impossible to access account funds!\n")
	fmt.Printf("- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!\n\n")
	return nil
}
