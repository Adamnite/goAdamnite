package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/adamnite/go-adamnite/caesar"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type CaesarHandler struct {
	server              *caesar.CaesarNode
	thisUser            *accounts.Account
	maxMessagesOnScreen int
	chatLogs            map[common.Address][]*chatText //the chat history by mapping
	showProgress        bool
}
type chatText struct {
	fromUs bool
	text   string
	time   string
}

func (ch *CaesarHandler) addChatMsg(msg *utils.CaesarMessage) {
	if msg.From.Address == ch.thisUser.Address {
		if _, exists := ch.chatLogs[msg.To.Address]; !exists {
			ch.chatLogs[msg.To.Address] = []*chatText{}
		}
		//well, we can't decrypt our own messages... shit...
		//TODO: setup the weird math keys that let us decrypt things we've sent...
		ch.chatLogs[msg.To.Address] = append(ch.chatLogs[msg.To.Address], &chatText{
			fromUs: true,
			text:   "*******",
			time:   msg.InitialTime.Format(time.Kitchen),
		})
	} else {
		if _, exists := ch.chatLogs[msg.From.Address]; !exists {
			ch.chatLogs[msg.From.Address] = []*chatText{}
		}
		text, _ := msg.GetMessageString(*ch.thisUser)
		ch.chatLogs[msg.From.Address] = append(ch.chatLogs[msg.To.Address], &chatText{
			fromUs: false,
			text:   text,
			time:   msg.InitialTime.Format(time.Kitchen),
		})
	}
}

// get a Caesar chat handler
func NewCaesarHandler() *CaesarHandler {
	return &CaesarHandler{maxMessagesOnScreen: 10, chatLogs: make(map[common.Address][]*chatText), showProgress: true}
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
		Help: "start <seed Node Connection String> <sending private key>", //TODO: add loading of the key from storage so this actually works...
		LongHelp: "Start the messaging server up. Allows local logging of messages, as well as sending of messages\n" +
			"\t <sending private key>\t: The key that signs messages sent from this server, if left blank, will be generated.\n" +
			"\t <seed Node Connection String>\t: The connection string (EG'1.2.3.4:5678') of a node you know and trust to be running, that you can form a network from.",
		Func: ch.Start,
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
	var progBar ishell.ProgressBar
	if ch.showProgress {
		progBar = c.ProgressBar()
		progBar.Prefix("starting up the server:")
		progBar.Start()
	}
	//TODO: this assume that no account was passed to local!
	ch.thisUser, _ = accounts.GenerateAccount()
	c.Println("Hosting from :", crypto.B58encode(ch.thisUser.PublicKey))
	server := caesar.NewCaesarNode(ch.thisUser)
	if err := server.Startup(); err != nil {
		c.Println(err)
		if ch.showProgress {
			progBar.Stop()
		}
		return
	}
	if ch.showProgress {
		progBar.Progress(10)
	}
	ch.server = server //if we've made it here, the servers probably gonna be working
	if len(c.Args) >= 1 {
		//test that we have enough arguments that a connection string *could* have been passed
		if err := server.ConnectToNetworkFrom(c.Args[0]); err != nil {
			c.Println(err)
			if ch.showProgress {
				progBar.Stop()
			}
			return
		}
	}

	if ch.showProgress {
		progBar.Progress(40)
	}
	server.FillNetworking(true) //TODO: only override if we don't already have a server running

	// server should be up, and well connected now!
	if ch.showProgress {
		progBar.Final(fmt.Sprintf("\nserver is up and running!\n\tHosting connection from:%v", server.GetConnectionPoint()))
		progBar.Progress(100)
		progBar.Stop()
	}
}

func (ch *CaesarHandler) OpenChat(c *ishell.Context) {
	if !ch.isServerLive() {
		c.Println("Server is not live!")
		return
	}
	ch.server.FillNetworking(false)
	if len(c.Args) < 1 {
		c.Println("need to have someone to talk to!")
		return
	}
	pubk, err := crypto.B58decode(c.Args[0])
	if err != nil {
		c.Println("ERR")
		c.Println(err)
		return
	}
	target := accounts.AccountFromPubBytes(pubk)

	//setup the texting display

	// messagesDisplaying := ch.server.GetMessagesBetween(*ch.thisUser, target)
	breakFully := false
	ch.server.NewMessageUpdater = func(msg *utils.CaesarMessage) {
		if msg.To.Address == ch.thisUser.Address {
			//only add messages *to* us, no point in adding messages from us, we cant decrypt them!
			ch.addChatMsg(msg)
			err := ch.updateChatScreen(c, target)
			breakFully = err != nil
		}
	}
	//get all the logged messages we have
	msgs := ch.server.GetMessagesBetween(*ch.thisUser, target)
	for _, m := range msgs {
		ch.addChatMsg(m)
	}
	if err := ch.updateChatScreen(c, target); err != nil || breakFully {
		c.Println(err)
		return
	}
	if len(c.Args) >= 2 { //then they're passing the message with the text (aka, debugging)
		text := ""
		for i := 2; i < len(c.Args); i++ {
			text = text + c.Args[i]
		}
		ch.sendMessage(target, text)
		err = ch.updateChatScreen(c, target)
		if err != nil {
			log.Println(err)
		}
		return

	}
	//TODO: get messages from here
	c.ShowPrompt(false)
	//I think this will auto break?
	text := ""
	for text != "\n" || err != nil {
		text = c.ReadMultiLinesFunc(func(s string) bool {
			return s[len(s)-1] == '\n'
		})
		ch.sendMessage(target, text)
		err = ch.updateChatScreen(c, target)
		log.Println("text worked(?): ", text)
	}

}

func (ch *CaesarHandler) updateChatScreen(c *ishell.Context, target accounts.Account) error {
	messagesToDisplay := ch.chatLogs[target.Address]
	c.ClearScreen()
	c.Printf("%v \t\t\t(you)%v", crypto.B58encode(target.PublicKey), crypto.B58encode(ch.thisUser.PublicKey))
	if len(messagesToDisplay) > ch.maxMessagesOnScreen {
		messagesToDisplay = messagesToDisplay[len(messagesToDisplay)-ch.maxMessagesOnScreen:]
	}
	for _, msg := range messagesToDisplay {
		if msg.fromUs {
			//then its on the left
			c.Println(fmt.Sprintf("[%v]%v", msg.time, msg.text)) //TODO: space this so it'll look nice
		} else {
			//on the right
			//assume max characters on the screen per side to be 50(ish)
			c.Println(fmt.Sprintf("\t\t[%v]%v", msg.time, msg.text)) //TODO: space this so it'll look nice

		}
	}
	return nil
}
func (ch *CaesarHandler) sendMessage(target accounts.Account, text string) {
	if _, exists := ch.chatLogs[target.Address]; !exists {
		ch.chatLogs[target.Address] = []*chatText{}
	}
	if err := ch.server.Send(target, text); err != nil {
		return
	}
	ch.chatLogs[target.Address] = append(ch.chatLogs[target.Address], &chatText{
		fromUs: true,
		text:   text,
		time:   time.Now().UTC().Format(time.Kitchen),
	})

}
