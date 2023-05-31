package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/abiosoft/ishell/v2"
	"github.com/adamnite/go-adamnite/caesar"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/fatih/color"
)

type CaesarHandler struct {
	server              *caesar.CaesarNode
	accounts            *AccountHandler
	thisUser            *accountBeingHeld
	maxMessagesOnScreen int
	chatLogs            map[common.Address][]*chatText //the chat history by mapping
	HoldingFocus        bool
}
type chatText struct {
	fromUs bool
	text   string
	time   string
}

func (ch *CaesarHandler) addChatMsg(msg *utils.CaesarMessage) {
	if msg.From.Address == ch.thisUser.account.Address {
		if _, exists := ch.chatLogs[msg.To.Address]; !exists {
			ch.chatLogs[msg.To.Address] = []*chatText{}
		}
		//well, we can't decrypt our own messages... shit...
		//TODO: setup the weird math keys that let us decrypt things we've sent...
		ch.chatLogs[msg.To.Address] = append(ch.chatLogs[msg.To.Address], &chatText{
			fromUs: true,
			text:   "*******",
			time:   msg.GetTime().Format(time.Kitchen),
		})
	} else {
		text, _ := msg.GetMessageString(*ch.thisUser.account)
		newMsg := chatText{
			fromUs: false,
			text:   text,
			time:   msg.GetTime().Format(time.Kitchen),
		}
		if _, exists := ch.chatLogs[msg.From.Address]; !exists {
			ch.chatLogs[msg.From.Address] = []*chatText{&newMsg}
		} else {
			ch.chatLogs[msg.From.Address] = append(ch.chatLogs[msg.From.Address], &newMsg)
		}

	}
}

// get a Caesar chat handler
func NewCaesarHandler(accounts *AccountHandler) *CaesarHandler {
	return &CaesarHandler{
		accounts:            accounts,
		maxMessagesOnScreen: 10,
		chatLogs:            make(map[common.Address][]*chatText),
		HoldingFocus:        false,
	}
}
func (ch CaesarHandler) isServerLive() bool {
	return ch.server != nil
}

func (ch *CaesarHandler) GetCaesarCommands() *ishell.Cmd {
	caesarFuncs := ishell.Cmd{
		Name: "caesar",
		Help: "Caesar messaging platform, built on top of Adamnite!",
	}
	caesarFuncs.AddCmd(&ishell.Cmd{
		Name: "start",
		Help: "start <seed Node Connection String>", //TODO: add loading of the key from storage so this actually works...
		LongHelp: "Start the messaging server up. Allows local logging of messages, as well as sending of messages\n" +
			"\t <seed Node Connection String>\t: The connection string (EG'1.2.3.4:5678') of a node you know and trust to be running, that you can form a network from.",
		Func: ch.Start,
	})
	caesarFuncs.AddCmd(&ishell.Cmd{
		Name: "stop",
		Help: "stop. Safely shutdown the Caesar server",
		Func: ch.Stop,
	})
	caesarFuncs.AddCmd(&ishell.Cmd{
		Name: "talk",
		Help: "talk <target> allows direct communication with someone",
		Func: ch.OpenChat,
	})
	return &caesarFuncs
}

func (ch *CaesarHandler) Start(c *ishell.Context) {
	if ch.isServerLive() {
		c.Println("Server is already live")
		return
	}
	//TODO: have this check if they have a consensus node running, and if so, if they want to use that for the messaging too

	//TODO: this assume that no account was passed to local!
	c.Println("\n\n")
	c.Println("account to host from")
	ch.thisUser = ch.accounts.SelectAccount(c)

	progBar := c.ProgressBar()
	progBar.Prefix("starting up the server:")
	progBar.Start()
	c.Println("Hosting from :", ch.thisUser.nickname)
	server := caesar.NewCaesarNode(ch.thisUser.account)
	if err := server.Startup(); err != nil {
		c.Println(err)
		progBar.Stop()
		return
	}
	progBar.Progress(10)

	ch.server = server //if we've made it here, the servers probably gonna be working
	if len(c.Args) >= 1 {
		//test that we have enough arguments that a connection string *could* have been passed
		if err := server.ConnectToNetworkFrom(c.Args[0]); err != nil {
			c.Println(err)
			progBar.Stop()
			return
		}
	}
	progBar.Progress(40)

	server.FillNetworking(true) //TODO: only override if we don't already have a server running
	// server should be up, and well connected now!
	progBar.Final(fmt.Sprintf("\nserver is up and running!\n\tHosting connection from:%v", server.GetConnectionPoint()))
	progBar.Progress(100)
	progBar.Stop()
}
func (ch *CaesarHandler) Stop(c *ishell.Context) {
	if !ch.isServerLive() {
		c.Println("Caesar Server is shut down")
		return
	}
	c.Println("Shutting down Caesar server")
	ch.server.Close()
	ch.server = nil
	c.Println("Caesar Server is shut down")
}

func (ch *CaesarHandler) OpenChat(c *ishell.Context) {
	if !ch.isServerLive() {
		c.Println("Server is not live!")
		return
	}
	ch.server.FillNetworking(false)
	var target *accountBeingHeld
	if len(c.Args) == 0 {
		c.Println("\nselect someone you know to talk to then!\n\n\n")
		target = ch.accounts.GetAnyAccount(c)
	} else {
		pubk, _ := crypto.B58decode(c.Args[0])
		if existing := ch.accounts.GetByPubkey(pubk); existing == nil {
			if _, err := ch.accounts.AddKnownAccountByB58(c.Args[0]); err != nil {
				c.Print("Err: ")
				c.Println(err)
				return
			}
		}
		target = ch.accounts.GetByPubkey(pubk)
	}
	if target == nil {
		c.Println("appears something went wrong. Please try again!")
		return
	}

	//setup the texting display
	c.Println("\n\n\n")
	ch.server.NewMessageUpdater = func(msg *utils.CaesarMessage) {
		if msg.To.Address == ch.thisUser.account.Address {
			//only add messages *to* us, no point in adding messages from us, we cant decrypt them!
			ch.addChatMsg(msg)
			ch.updateChatScreen(c, target)
		}
	}
	//get all the logged messages we have
	msgs := ch.server.GetMessagesBetween(*ch.thisUser.account, *target.account) //TODO: fix this. Right now if you run this again in the same CLI instance, it will double the messages
	for _, m := range msgs {
		ch.addChatMsg(m)
	}
	err := ch.updateChatScreen(c, target)
	if err != nil {
		c.Println(err)
		return
	}
	if len(c.Args) >= 2 { //then they're passing the message with the text (aka, debugging)
		text := ""
		for i := 1; i < len(c.Args); i++ {
			text = text + c.Args[i]
		}
		ch.sendMessage(*target.account, text)
		err = ch.updateChatScreen(c, target)
		if err != nil {
			log.Println(err)
		}
		return

	}

	c.SetPrompt("|msg|")
	c.Println("\n\n\n")
	ch.HoldingFocus = true
	text := " "
	for text != "" && err == nil {
		text = c.ReadMultiLinesFunc(func(s string) bool {
			if len(s) == 0 {
				return false
			}
			return s[len(s)-1] == '\n'
		})
		ch.sendMessage(*target.account, text)
		c.Println("\n\n\n")
		err = ch.updateChatScreen(c, target)
		c.Println("")
	}
	if err != nil {
		c.Println(err)
	}
	c.SetPrompt(">adm>")
	ch.HoldingFocus = false
}

func (ch *CaesarHandler) updateChatScreen(c *ishell.Context, target *accountBeingHeld) error {
	c.Println("\n\n\n\n\n\n")
	userMsgColor := color.New(color.BgGreen)
	otherMsgColor := color.New(color.BgBlue)
	messagesToDisplay := ch.chatLogs[target.account.Address]
	c.ClearScreen()
	c.Printf("%v(them) \t\t(you)%v\n", otherMsgColor.Sprint(target.nickname), userMsgColor.Sprint(ch.thisUser.nickname))
	if len(messagesToDisplay) > ch.maxMessagesOnScreen {
		messagesToDisplay = messagesToDisplay[len(messagesToDisplay)-ch.maxMessagesOnScreen:]
	}

	for _, msg := range messagesToDisplay {
		if msg.fromUs {
			//then its on the left
			c.Println(userMsgColor.Sprintf("[%v]%v", msg.time, msg.text)) //TODO: space this so it'll look nice
		} else {
			//on the right
			//assume max characters on the screen per side to be 50(ish)
			c.Println(otherMsgColor.Sprintf("[%v]%v", msg.time, msg.text)) //TODO: space this so it'll look nice

		}
	}
	return nil
}
func (ch *CaesarHandler) sendMessage(target *accounts.Account, text string) {
	//TODO: check the account is real, otherwise this will break
	if err := ch.server.Send(target, text); err != nil {
		log.Println(err)
		return
	}
	newMsg := chatText{
		fromUs: true,
		text:   text,
		time:   time.Now().UTC().Format(time.Kitchen),
	}
	if _, exists := ch.chatLogs[target.Address]; !exists {
		ch.chatLogs[target.Address] = []*chatText{&newMsg}
	} else {
		ch.chatLogs[target.Address] = append(ch.chatLogs[target.Address], &newMsg)
	}

}
