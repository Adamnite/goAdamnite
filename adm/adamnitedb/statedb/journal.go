package statedb

import (
	"math/big"

	"github.com/adamnite/go-adamnite/utils"
)

type journalEntry interface {
	// revert undoes the changes introduced by this journal entry.
	revert(*StateDB)

	// dirtied returns the adamnite address modified by this journal entry.
	dirtied() *utils.Address
}

type journal struct {
	entries []journalEntry
	dirties map[utils.Address]int
}

func newJournal() *journal {
	return &journal{
		dirties: make(map[utils.Address]int),
	}
}

func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if addr := entry.dirtied(); addr != nil {
		j.dirties[*addr]++
	}
}

func (j *journal) revert(statedb *StateDB, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		j.entries[i].revert(statedb)

		if addr := j.entries[i].dirtied(); addr != nil {
			if j.dirties[*addr]--; j.dirties[*addr] == 0 {
				delete(j.dirties, *addr)
			}
		}
	}
	j.entries = j.entries[:snapshot]
}

func (j *journal) dirty(addr utils.Address) {
	j.dirties[addr]++
}

func (j *journal) length() int {
	return len(j.entries)
}

type createObjectChange struct {
	account *utils.Address
}

func (change createObjectChange) revert(s *StateDB) {
	delete(s.stateObjects, *change.account)
	delete(s.stateObjectsDirty, *change.account)
}

func (change createObjectChange) dirtied() *utils.Address {
	return change.account
}

type resetObjectChange struct {
	prev         *stateObject
	prevdestruct bool
}

func (change resetObjectChange) revert(s *StateDB) {
	s.setStateObject(change.prev)
}

func (change resetObjectChange) dirtied() *utils.Address {
	return nil
}

type suicideChange struct {
	account     *utils.Address
	prev        bool // whether account had already suicided
	prevbalance *big.Int
}

func (sch suicideChange) revert(s *StateDB) {
	obj := s.getStateObject(*sch.account)
	if obj != nil {
		obj.suicided = sch.prev
		obj.setBalance(sch.prevbalance)
	}
}

func (sch suicideChange) dirtied() *utils.Address {
	return sch.account
}

type balanceChange struct {
	account *utils.Address
	prev    *big.Int
}

func (bch balanceChange) revert(s *StateDB) {
	s.getStateObject(*bch.account).setBalance(bch.prev)
}

func (bch balanceChange) dirtied() *utils.Address {
	return bch.account
}

type nonceChange struct {
	account *utils.Address
	prev    uint64
}

func (nch nonceChange) revert(s *StateDB) {
	s.getStateObject(*nch.account).setNonce(nch.prev)
}

func (nch nonceChange) dirtied() *utils.Address {
	return nch.account
}
