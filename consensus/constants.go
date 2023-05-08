package consensus

type consensusHandlingTypes int8

const (
	NetworkingOnly        consensusHandlingTypes = iota
	PrimaryTransactions                          //representing chamber A, or main transactions
	SecondaryTransactions                        //representing chamber B, or VM consensus
)
