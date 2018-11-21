package main

import (
	"fmt"
	"reflect"
)

type Ledger struct {
	Accounts map[string]int
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) SignedTransaction(t *SignedTransaction) bool {

	if t.isValid() && l.isValid(t) {

		// Register accounts, if they aren't there
		l.initializeAccount(t.From)
		l.initializeAccount(t.To)

		amount := t.Amount - 1

		l.Accounts[t.From] -= amount
		l.Accounts[t.To] += amount
		return true
	}

	return false
}

func (l *Ledger) initializeAccount(pkStr string) {
	if _, alreadyInitialized := l.Accounts[pkStr]; !alreadyInitialized {
		l.Accounts[pkStr] = 0
	}
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

func (l *Ledger) AddAmount(account string, amount int) {
	l.initializeAccount(account)
	l.Accounts[account] += amount
}

func (l *Ledger) PrintStatus() {
	var keys = reflect.ValueOf(l.Accounts).MapKeys()
	fmt.Println("There are", len(keys), "keys")
	for i := 0; i < len(keys); i++ {
		var key = keys[i]
		var str = key.String()
		fmt.Println("Account", i, "has", l.Accounts[str], "AU(s). Key:", str[0:50]+"...")
	}
}

func (l *Ledger) Match(ledger *Ledger) bool {
	var keys = reflect.ValueOf(l.Accounts).MapKeys()

	for _, key := range keys {
		keyStr := key.String()

		if l.Accounts[keyStr] != ledger.Accounts[keyStr] {
			return false
		}
	}

	return true
}
