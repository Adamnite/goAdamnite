package cmd

import (
	"fmt"

	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

// holds all the accounts we could sign with
type AccountHandler struct {
	selectedAccount *accountBeingHeld
	ourAccounts     map[common.Address]accountBeingHeld
	knownAccounts   map[common.Address]accountBeingHeld
}

type accountBeingHeld struct {
	//acts same as an account, but with some more UI items added
	account  *accounts.Account
	nickname string
}

func NewAccountHandler() *AccountHandler {
	ah := AccountHandler{
		ourAccounts:   make(map[common.Address]accountBeingHeld),
		knownAccounts: make(map[common.Address]accountBeingHeld),
	}
	return &ah
}

func (ah *AccountHandler) GetAccountCommands() *ishell.Cmd {
	accountFuncs := ishell.Cmd{
		Name: "accounts",
		Help: "accounts is used to store items",
	}
	accountFuncs.AddCmd(&ishell.Cmd{
		Name: "edit",
		Help: "edit an account",
		Func: ah.EditAccount,
	})
	accountFuncs.AddCmd(&ishell.Cmd{
		Name: "add",
		Help: "add an account",
		Func: ah.AddAccount,
	})
	accountFuncs.AddCmd(&ishell.Cmd{
		Name: "new",
		Help: "generate a new account",
		Func: ah.GenerateAccount,
	})
	return &accountFuncs
}
func (ah *AccountHandler) EditAccount(c *ishell.Context) {
	selected := ah.GetAnyAccount(c)
	editActions := []string{
		"Get Public Key",
		"Get Address",
		"Change Nickname",
		"remove account",
		"exit",
	}
	actionKey := c.MultiChoice(editActions, "Select action to take")
	switch actionKey {
	case 0: //get public key
		c.Println(crypto.B58encode(selected.account.PublicKey))
	case 1:
		c.Println(selected.account.Address.Hex())
	case 2:
		c.ShowPrompt(false)
		c.Print("New nickname (leave blank to revert to address): ")
		newName := c.ReadLine()
		if newName == "" {
			ah.SetNickname(*selected.account, selected.account.Address.Hex())
		} else {
			ah.SetNickname(*selected.account, newName)
		}
		c.ShowPrompt(true)
	case 3:
		if i := c.MultiChoice([]string{"yes", "no"}, "are you sure you want to remove this account?"); i == 0 {
			ah.RemoveAccount(selected.account)
			ah.selectedAccount = nil
		}
	default:
		return
	}
	c.ShowPrompt(true)
}

func (ah *AccountHandler) SelectAccount(c *ishell.Context) *accountBeingHeld {
	c.Println("\n") //adding empty space above helps prevent weird glitches that can happen with overlapping items
	accountSelection := []string{
		"add account",
	}
	accountOptions := []*accountBeingHeld{}
	for _, ac := range ah.ourAccounts {
		accountSelection = append(accountSelection, ac.nickname)
		accountOptions = append(accountOptions, &ac)
	}
	selected := c.MultiChoice(accountSelection, "select account(or generate a new one)")
	if selected == 0 || selected == -1 {
		//generate an account
		ah.AddAccount(c)
		return ah.selectedAccount
	}
	ah.selectedAccount = accountOptions[selected-1] //we have an option before the others
	return ah.selectedAccount
}

func (ah *AccountHandler) GetAnyAccount(c *ishell.Context) *accountBeingHeld {
	if i := c.MultiChoice([]string{"local", "other"}, "local or others account"); i != 1 {
		return ah.SelectAccount(c)
	}

	accountSelection := []string{
		"add account",
	}
	accountOptions := []*accountBeingHeld{}
	for _, ac := range ah.knownAccounts {
		accountSelection = append(accountSelection, ac.nickname)
		accountOptions = append(accountOptions, &ac)
	}
	selected := c.MultiChoice(accountSelection, "select account(or add one)")
	if selected == 0 || selected == -1 {
		//add the
		c.ShowPrompt(false)
		c.Print("Enter the b58 encoded public key: ")
		newPub := c.ReadLine()
		c.ShowPrompt(true)
		if ac, err := ah.AddKnownAccountByB58(newPub); err != nil {
			c.Print("Error adding that account: ")
			c.Println(err)
			return nil
		} else {
			foo := ah.knownAccounts[ac.Address]
			return &foo
		}
	}
	return accountOptions[selected-1] //we have an option before the others
}

// returns nil if there isn't a currently selected account
func (ah *AccountHandler) GetSelected() *accountBeingHeld {
	if ah.selectedAccount == nil {
		return nil
	}
	return ah.selectedAccount
}
func (ah *AccountHandler) GenerateAccount(c *ishell.Context) {
	c.Println("Generating the new account now")
	ac, _ := accounts.GenerateAccount()
	c.Println("it is vital that you keep this key secure")
	saveType := c.MultiChoice([]string{"saveInFile", "print b58 encoding"}, "how would you like to receive your private key.")
	switch saveType {
	case 0:
		c.Println("enter store point")
		savePoint := c.ReadLine()
		if err := ac.Store(savePoint); err != nil {
			c.Println(err)
			return
		}
	default:
		c.Println("Private key is: ", ac.GetPrivateB58())
	}
	if err := ah.AddAccountByAccount(*ac); err != nil {
		c.Println(err)
	}
}
func (ah *AccountHandler) AddAccount(c *ishell.Context) {
	keyType := c.MultiChoice(
		[]string{
			"generate new",
			"b58 encoded string",
			"b58 encoding with privacy",
			"from file",
		},
		"adding new accounts, please select the format you would like to enter it in",
	)

	switch keyType {
	case 1: //they have a b58 encoding of an account
		c.Println("enter key:")
		b58pk := c.ReadLine()
		if err := ah.AddAccountByB58(b58pk); err != nil {
			c.Println("it appears there was an error inputting that private key, please try again")
			return
		}
	case 2: //they have a b58 encoding of the private key, but want privacy
		c.Println("enter key:")
		b58pk := c.ReadPassword()
		if err := ah.AddAccountByB58(b58pk); err != nil {
			c.Println("it appears there was an error inputting that private key, please try again")
			return
		}
	case 3:
		c.Println("enter store point")
		savePoint := c.ReadLine()
		ac, err := accounts.AccountFromStorage(savePoint)
		if err != nil {
			c.Println(err)
			return
		}
		if err := ah.AddAccountByAccount(ac); err != nil {
			c.Println(err)
			return
		}
	default:
		ah.GenerateAccount(c)
	}
	//every error calls a return
	c.Println("account successfully loaded")
}
func (ah *AccountHandler) RemoveAccount(ac *accounts.Account) {
	delete(ah.ourAccounts, ac.Address)
	delete(ah.knownAccounts, ac.Address)
}
func (ah AccountHandler) GetByNickname(name string) *accounts.Account {
	for _, ac := range ah.ourAccounts {
		if ac.nickname == name {
			return ac.account
		}
	}
	for _, ac := range ah.knownAccounts {
		if ac.nickname == name {
			return ac.account
		}
	}
	return nil
}
func (ah AccountHandler) GetNickname(ac *accounts.Account) string {
	if acHeld, exists := ah.ourAccounts[ac.Address]; exists {
		return acHeld.nickname
	}
	if acHeld, exists := ah.knownAccounts[ac.Address]; exists {
		return acHeld.nickname
	}
	return ""
}

func (ah *AccountHandler) SetNickname(account accounts.Account, newName string) {
	if acHeld, exists := ah.ourAccounts[account.Address]; exists {
		acHeld.nickname = newName
		ah.ourAccounts[account.Address] = acHeld
		return
	}
	if acHeld, exists := ah.knownAccounts[account.Address]; exists {
		acHeld.nickname = newName
		ah.knownAccounts[account.Address] = acHeld
	}
}

// adding accounts

func (ah *AccountHandler) AddAccountByB58(pk string) error {
	pkb, err := crypto.B58decode(pk)
	if err != nil {
		return err
	}
	pk = "      " //clearing the string
	return ah.AddAccountByPrivateByte(pkb)
}

func (ah *AccountHandler) AddAccountByPrivateByte(pk []byte) error {
	ac, err := accounts.AccountFromPrivBytes(pk)
	if err != nil {
		return err
	}
	for i := range pk {
		pk[i] = 0 //clearing it by setting the pk bytes to 0s
	}
	return ah.AddAccountByAccount(ac)
}

func (ah *AccountHandler) AddAccountByAccount(ac accounts.Account) error {
	if _, exists := ah.ourAccounts[ac.Address]; exists {
		return fmt.Errorf("account already added within us")
	}
	if _, exists := ah.knownAccounts[ac.Address]; exists {
		return fmt.Errorf("account already added within known list")
	}
	newAccount := accountBeingHeld{
		account:  &ac,
		nickname: ac.Address.Hex(),
	}
	ah.ourAccounts[ac.Address] = newAccount
	ah.selectedAccount = &newAccount
	return nil
}

// adding other contacts
func (ah *AccountHandler) AddKnownAccountByB58(pk string) (*accounts.Account, error) {
	pkb, err := crypto.B58decode(pk)
	if err != nil {
		return nil, err
	}
	ac := accounts.AccountFromPubBytes(pkb)
	return &ac, ah.AddKnownByAccount(ac)
}

func (ah *AccountHandler) AddKnownByAccount(ac accounts.Account) error {
	if _, exists := ah.ourAccounts[ac.Address]; exists {
		return fmt.Errorf("account already added within us")
	}
	if _, exists := ah.knownAccounts[ac.Address]; exists {
		return fmt.Errorf("account already added within known list")
	}
	ah.knownAccounts[ac.Address] = accountBeingHeld{
		account:  &ac,
		nickname: ac.Address.Hex(),
	}
	return nil
}
