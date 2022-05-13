package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/internal/utils"
	"github.com/adamnite/go-adamnite/log15"
	"github.com/adamnite/go-adamnite/p2p/discover"
	"github.com/adamnite/go-adamnite/p2p/enode"
	"github.com/adamnite/go-adamnite/p2p/nat"
	"github.com/adamnite/go-adamnite/p2p/netutil"
)

func main() {
	var (
		listenAddress = flag.String("addr", ":3880", "listen address")
		verbosity     = flag.Int("verbosity", int(log15.LvlInfo), "log verbosity (0-5)")
		vmodule       = flag.String("vmodule", "", "log verbosity pattern")
		natdesc       = flag.String("nat", "none", "port mapping mechanism (any|none|upnp|pmp|extip:<IP>)")
		genKey        = flag.String("genkey", "", "generate a node key")
		writeAddr     = flag.Bool("writeaddress", false, "write out the node's public key and quit")
		netrestrict   = flag.String("netrestrict", "", "restrict network communication to the given IP networks (CIDR masks)")
		nodeKeyFile   = flag.String("nodekey", "", "private key filename")

		nodeKey *ecdsa.PrivateKey
		err     error
	)

	flag.Parse()

	glogger := log15.NewGlogHandler(log15.StreamHandler(os.Stderr, log15.TerminalFormat()))
	glogger.Verbosity(log15.Lvl(*verbosity))
	glogger.Vmodule(*vmodule)
	log15.Root().SetHandler(glogger)

	natm, err := nat.Parse(*natdesc)
	if err != nil {
		utils.Fatalf("-nat: %v", err)
	}

	switch {
	case *genKey != "":
		nodeKey, err = crypto.GenerateKey()
		if err != nil {
			utils.Fatalf("could not generate key: %v", err)
		}
		if err = crypto.SaveECDSA(*genKey, nodeKey); err != nil {
			utils.Fatalf("%v", err)
		}
		if !*writeAddr {
			return
		}
	case *nodeKeyFile == "":
		utils.Fatalf("Use -nodekey to specify a private key")
	case *nodeKeyFile != "":
		if nodeKey, err = crypto.LoadECDSA(*nodeKeyFile); err != nil {
			utils.Fatalf("-nodekey: %v", err)
		}
	}

	if *writeAddr {
		fmt.Printf("%x\n", crypto.FromECDSAPub(&nodeKey.PublicKey)[1:])
		os.Exit(0)
	}

	var restrictList *netutil.Netlist
	if *netrestrict != "" {
		restrictList, err = netutil.ParseNetlist(*netrestrict)
		if err != nil {
			utils.Fatalf("-netrestrict: %v", err)
		}
	}

	addr, err := net.ResolveUDPAddr("udp", *listenAddress)
	if err != nil {
		utils.Fatalf("-ResolveUDPAddr: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		utils.Fatalf("-ListenUDP: %v", err)
	}

	realaddr := conn.LocalAddr().(*net.UDPAddr)
	if natm != nil {
		if !realaddr.IP.IsLoopback() {
			go nat.Map(natm, nil, "udp", realaddr.Port, realaddr.Port, "adamnite discovery")
		}
		if ext, err := natm.ExternalIP(); err == nil {
			realaddr = &net.UDPAddr{IP: ext, Port: realaddr.Port}
		}
	}

	printNotice(&nodeKey.PublicKey, *realaddr)
	db, err := enode.OpenDB("")
	localNode := enode.NewLocalNode(db, nodeKey)

	cfg := discover.Config{
		PrivateKey:  nodeKey,
		NetRestrict: restrictList,
	}

	if _, err := discover.ListenV5(conn, localNode, cfg); err != nil {
		utils.Fatalf("%v", err)
	}

	select {}
}

func printNotice(nodeKey *ecdsa.PublicKey, addr net.UDPAddr) {
	if addr.IP.IsUnspecified() {
		addr.IP = net.IP{127, 0, 0, 1}
	}
	n := enode.NewV4(nodeKey, addr.IP, 0, addr.Port)
	fmt.Println(n.URLv4())
	fmt.Println("Note: you're using cmd/bootnode, a developer tool.")
	fmt.Println("We recommend using a regular node as bootstrap node for production deployments.")
}
