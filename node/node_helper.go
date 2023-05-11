package node 

import (
    "context"
    "fmt"
    "time"
	// Import other relevant directories as needed.
)

type Node struct {
    Context          context.Context
    Config           *Config
    TransactionHandler *TransactionHandler
    TransactionMempool *TransactionMempool
    Consensus        *Consensus
    CatchupProvider  *CatchupProvider
    GenesisHash      []byte
    Logger           *Logger
    CryptoWorker     *CryptoWorker
    BlockWorker      *BlockWorker
    CoreLedger       *CoreLedger
}

type Config struct {
    RootDirectory string
    SyncInterval  int
    SyncRetry     int
}

type TransactionHandler struct {
    OffDB           *OffDB
    VM              *VM
}

type TransactionMempool struct {
    Transactions    []Transaction
}

type Consensus struct {
    CurrentRound    int
    CurrentVersion  string
    NextVersion     string
    Timestamp       time.Time
    SyncTime        int
    TotalBlocks     int
    TotalAccounts   int
}

type CatchupProvider struct {
    NodeID          int
}

type Logger struct {
    Level           int
}

type CryptoWorker struct {
    Alg             string
}

type BlockWorker struct {
    RootDirectory   string
}

type CoreLedger struct {
    RootDirectory   string
}

func NewNode(config *Config) *Node {
    context := context.Background()
    offDB := NewOffDB()
    vm := NewVM()
    transactionHandler := &TransactionHandler{OffDB: offDB, VM: vm}
    transactionMempool := &TransactionMempool{Transactions: make([]Transaction, 0)}
    consensus := &Consensus{CurrentRound: 0, CurrentVersion: "", NextVersion: "", Timestamp: time.Now(), SyncTime: 0, TotalBlocks: 0, TotalAccounts: 0}
    catchupProvider := &CatchupProvider{NodeID: 0}
    genesisHash := []byte{}
    logger := &Logger{Level: 0}
    cryptoWorker := &CryptoWorker{Alg: "SHA-512"}
    blockWorker := &BlockWorker{RootDirectory: config.RootDirectory}
    coreLedger := &CoreLedger{RootDirectory: config.RootDirectory}
    return &Node{Context: context, Config: config, TransactionHandler: transactionHandler, TransactionMempool: transactionMempool, Consensus: consensus, CatchupProvider: catchupProvider, GenesisHash: genesisHash, Logger: logger, CryptoWorker: cryptoWorker, BlockWorker: blockWorker, CoreLedger: coreLedger}
}

func (n *Node) Start() {
    fmt.Println("Starting node...")
    fmt.Printf("Root directory: %s\n", n.Config.RootDirectory)
    fmt.Printf("Sync interval: %d\n", n.Config.SyncInterval)
    fmt.Printf("Sync retry: %d\n", n.Config.SyncRetry)
    fmt.Printf("Current round: %d\n", n.Consensus.CurrentRound)
    fmt.Printf("Current version: %s\n", n.Consensus.CurrentVersion)
    fmt.Printf("Next version: %s\n", n.Consensus.NextVersion)
    fmt.Printf("Timestamp: %v\n", n.Consensus.Timestamp)
    fmt.Printf("Sync time: %d\n", n.Consensus.SyncTime)
    fmt.Printf("Total blocks: %d\n", n.Consensus.TotalBlocks)
    fmt.Printf("Total accounts: %d\n", n.Consensus.TotalAccounts)
    fmt.Println("Node started.")
}

func (n *Node) Stop() {
    fmt.Println("Stopping node...")
    fmt.Println("Node stopped.")
}

func (n *Node) GetTime() time.Time {
    return time.Now()
