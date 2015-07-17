package transactionpool

import (
	"github.com/NebulousLabs/Sia/build"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
)

func (tp *TransactionPool) sendNotification() {
	for _, subscriber := range tp.notifySubscribers {
		subscriber <- struct{}{}
	}
}

func (tp *TransactionPool) updateSubscribersTransactions() {
	var txns []types.Transaction
	var cc modules.ConsensusChange
	for _, tSet := range tp.transactionSets {
		txns = append(txns, tSet...)
	}
	for _, tSetDiff := range tp.transactionSetDiffs {
		cc = cc.Append(tSetDiff)
	}
	for _, subscriber := range tp.subscribers {
		subscriber.ReceiveUpdatedUnconfirmedTransactions(txns, cc)
	}
	tp.sendNotification()
}

func (tp *TransactionPool) updateSubscribersConsensus(cc modules.ConsensusChange) {
	for _, subscriber := range tp.subscribers {
		subscriber.ReceiveConsensusSetUpdate(cc)
	}
	tp.sendNotification()
}

func (tp *TransactionPool) TransactionPoolNotify() <-chan struct{} {
	c := make(chan struct{}, modules.NotifyBuffer)
	id := tp.mu.Lock()
	c <- struct{}{}
	tp.notifySubscribers = append(tp.notifySubscribers, c)
	tp.mu.Unlock(id)
	return c
}

func (tp *TransactionPool) TransactionPoolSubscribe(subscriber modules.TransactionPoolSubscriber) {
	id := tp.mu.Lock()
	tp.subscribers = append(tp.subscribers, subscriber)
	println("newcomer")
	for i := 0; i <= tp.consensusChangeIndex; i++ {
		cc, err := tp.consensusSet.ConsensusChange(i)
		println(i)
		println(tp.consensusChangeIndex)
		if err != nil && build.DEBUG {
			panic(err)
		}
		subscriber.ReceiveConsensusSetUpdate(cc)
		println("sent an update successfully")
	}
	tp.mu.Unlock(id)
}
