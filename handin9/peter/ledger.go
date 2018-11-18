package main

import (
	"fmt"
	"reflect"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	Lock     sync.Mutex
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) SignedTransaction(t *SignedTransaction) bool {
	l.Lock.Lock()
	defer l.Lock.Unlock()

	initializeAccount := func(pkStr string) {
		if _, alreadyInitialized := l.Accounts[pkStr]; !alreadyInitialized {
			l.Accounts[pkStr] = 0
		}
	}

	if t.isValid() && l.isValid(t) {

		// Register accounts, if they aren't there
		initializeAccount(t.From)
		initializeAccount(t.To)

		amount := t.Amount - 1

		l.Accounts[t.From] -= amount
		l.Accounts[t.To] += amount
		return true
	}

	return false
}

func (l *Ledger) isValid(t *SignedTransaction) bool {
	fromAmount := 0

	if val, ok := l.Accounts[t.From]; ok {
		fromAmount = val
	}

	if fromAmount-t.Amount < 0 {
		return false
	}

	return true
}

func (l *Ledger) InitializePremiumAccount(pkStr string) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	l.Accounts[pkStr] = 1000000
}

func (l *Ledger) PrintStatus() {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	var keys = reflect.ValueOf(l.Accounts).MapKeys()
	fmt.Println("There are", len(keys), "keys")
	for i := 0; i < len(keys); i++ {
		var key = keys[i]
		var str = key.String()
		fmt.Println("Account", i, "has", l.Accounts[str], "dineros. Key:", str[0:50]+"...")
	}
}
