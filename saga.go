package saga

import (
	"context"
	"fmt"
)

// Saga represents a Saga which is a collection of Transactions and Compensations.
type Saga struct {
	name             string
	firstTransaction string
	transactions     map[string]Transaction
	compensations    map[string]Compensation
}

// Transaction represents a step in the Saga.
type Transaction struct {
	Name                string
	NextTransactionName string

	CompensationName string
	IsSavePoint      bool
	Func             TransactionFunc
}

// Compensation represents a Compensation step for a Transaction in the Saga.
type Compensation struct {
	Name                 string
	NextCompensationName string
	Func                 TransactionFunc
}

// TransactionFunc is the function signature for a function that can be executed as a Transaction or a Compensation.
type TransactionFunc func(ctx context.Context, params any) error

// Name returns the name of the Saga.
func (s Saga) Name() string {
	return s.name
}

// FirstTransaction returns the ID of the first Transaction in the Saga.
func (s Saga) FirstTransaction() string {
	return s.firstTransaction
}

// Next returns the ID of the next step in the Saga after the given transaction or compensation.
func (s Saga) Next(transactionID string) string {
	if t, ok := s.transactions[transactionID]; ok {
		return t.NextTransactionName
	}

	if c, ok := s.compensations[transactionID]; ok {
		if t, ok := s.transactions[c.NextCompensationName]; ok && t.IsSavePoint {
			return t.NextTransactionName
		}
		return c.NextCompensationName
	}

	return ""
}

// Compensation returns the ID of the Compensation for the given Transaction ID.
func (s Saga) Compensation(transactionID string) string {
	if t, ok := s.transactions[transactionID]; ok {
		return t.CompensationName
	}
	return ""
}

// ExecuteTransaction executes the Transaction or Compensation with the given ID, passing the provided data to the TransactionFunc.
func (s Saga) ExecuteTransaction(ctx context.Context, transactionID string, data any) error {
	if t, ok := s.transactions[transactionID]; ok {
		return t.Func(ctx, data)
	}

	if c, ok := s.compensations[transactionID]; ok {
		return c.Func(ctx, data)
	}

	return fmt.Errorf("no transaction or compensation with id %s", transactionID)
}
