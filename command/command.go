package command

import (
	"github.com/google/uuid"
	"time"
)

// Name represents the type of command.
type Name int

const (
	BeginSagaCommand Name = iota + 1
	EndSagaCommand
	BeginTransactionCommand
	EndTransactionCommand
	AbortTransactionCommand
	AbortSagaCommand
)

// Command represents a Saga command.
type Command struct {
	ID   string `json:"id"`
	Name Name   `json:"name"`

	SagaID     string `json:"sagaID"`
	SagaName   string `json:"sagaName"`
	SagaParams any    `json:"sagaParams"`

	TransactionID  string `json:"transactionID"`
	CompensationID string `json:"compensationID"`

	CreatedAt time.Time `json:"createdAt"`
}

// BeginSaga returns a command to begin a new saga.
func BeginSaga(sagaName string, params any) Command {
	return Command{
		ID:         uuid.NewString(),
		Name:       BeginSagaCommand,
		SagaID:     uuid.NewString(),
		SagaName:   sagaName,
		SagaParams: params,
		CreatedAt:  time.Now(),
	}
}

// EndSaga returns a command to end an existing saga.
func EndSaga(sagaName, sagaID string) Command {
	return Command{
		ID:        uuid.NewString(),
		SagaName:  sagaName,
		Name:      EndSagaCommand,
		SagaID:    sagaID,
		CreatedAt: time.Now(),
	}
}

// AbortSaga returns a command to abort an existing saga.
func AbortSaga(sagaName, sagaID string, transactionID string) Command {
	return Command{
		ID:            uuid.NewString(),
		Name:          AbortSagaCommand,
		SagaID:        sagaID,
		SagaName:      sagaName,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}
}

// BeginTransaction returns a command to begin a new transaction (or compensation transaction).
func BeginTransaction(sagaName, sagaID string, transactionID string, params any) Command {
	return Command{
		ID:            uuid.NewString(),
		Name:          BeginTransactionCommand,
		SagaID:        sagaID,
		SagaName:      sagaName,
		SagaParams:    params,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}
}

// EndTransaction returns a command to end an existing transaction.
func EndTransaction(sagaName, sagaID, transactionID string, params any) Command {
	return Command{
		ID:            uuid.NewString(),
		Name:          EndTransactionCommand,
		SagaID:        sagaID,
		SagaName:      sagaName,
		SagaParams:    params,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}
}

// EndTransactionCompensate returns a command to end an existing transaction and compensate it. Use for AbortSagaCommand.
func EndTransactionCompensate(sagaName, sagaID string, transactionID, compensationID string, params any) Command {
	return Command{
		ID:             uuid.NewString(),
		Name:           EndTransactionCommand,
		SagaID:         sagaID,
		SagaName:       sagaName,
		SagaParams:     params,
		TransactionID:  transactionID,
		CompensationID: compensationID,
		CreatedAt:      time.Now(),
	}
}

// AbortTransaction returns a command to abort a transaction.
func AbortTransaction(sagaName, sagaID string, transactionID string, params any) Command {
	return Command{
		ID:            uuid.NewString(),
		Name:          AbortTransactionCommand,
		SagaID:        sagaID,
		SagaName:      sagaName,
		SagaParams:    params,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}
}
