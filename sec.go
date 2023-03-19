package saga

import (
	"context"
	"errors"
	"fmt"

	"github.com/mrfeathers/go-saga/command"
)

// SEC represents the Saga Execution Coordinator.
type SEC struct {
	errLogger     ErrLogger
	commandLogger CommandLogger
	sagas         map[string]Saga
}

// NewSEC creates a new SEC instance with the given CommandLogger and ErrLogger.
func NewSEC(commandLogger CommandLogger, errLogger ErrLogger) *SEC {
	return &SEC{commandLogger: commandLogger, errLogger: errLogger, sagas: make(map[string]Saga)}
}

type ErrLogger interface {
	Log(err error)
}

// CommandLogger defines an interface for reading, writing and committing commands.
type CommandLogger interface {
	Commit(ctx context.Context, c command.Command) error
	Read(ctx context.Context) (command.Command, error)
	Write(ctx context.Context, c command.Command) error
}

// RegisterSaga adds a Saga to the SEC.
func (s *SEC) RegisterSaga(saga Saga) {
	s.sagas[saga.name] = saga
}

// Start starts the Saga Execution Coordinator loop.
func (s *SEC) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c, err := s.commandLogger.Read(ctx)
			if err != nil {
				s.errLogger.Log(fmt.Errorf("command read failed: %w", err))
				continue
			}

			err = s.ProcessCommand(ctx, c)
			if err != nil {
				s.errLogger.Log(fmt.Errorf("process command %s failed: %w", c.ID, err))
				continue
			}

			err = s.commandLogger.Commit(ctx, c)
			if err != nil {
				s.errLogger.Log(fmt.Errorf("command %s commit failed: %w", c.ID, err))
				continue
			}
		}
	}
}

var ErrAbortSaga = errors.New("saga aborted")

// ProcessCommand processes the given command within the context of the SEC.
func (s *SEC) ProcessCommand(ctx context.Context, c command.Command) error {
	saga, ok := s.sagas[c.SagaName]
	if !ok {
		return fmt.Errorf("no saga with name %s exists", c.SagaName)
	}

	switch c.Name {
	case command.BeginSagaCommand:
		nextTransaction := saga.FirstTransaction()
		if nextTransaction == "" {
			return s.commandLogger.Write(ctx, command.EndSaga(c.SagaName, c.SagaID))
		}

		return s.commandLogger.Write(ctx, command.BeginTransaction(c.SagaName, c.SagaID, nextTransaction, c.SagaParams))
	case command.BeginTransactionCommand:
		execErr := saga.ExecuteTransaction(ctx, c.TransactionID, c.SagaParams)
		if execErr != nil {
			if errors.Is(execErr, ErrAbortSaga) {
				// abort saga, need to compensate transactions to the save point
				return s.commandLogger.Write(ctx, command.AbortSaga(c.SagaName, c.SagaID, c.TransactionID))
			}
			// abort transaction, need to repeat this transaction again
			return s.commandLogger.Write(ctx, command.AbortTransaction(c.SagaName, c.SagaID, c.TransactionID, c.SagaParams))
		}

		return s.commandLogger.Write(ctx, command.EndTransaction(c.SagaName, c.SagaID, c.TransactionID, c.SagaParams))
	case command.AbortTransactionCommand:
		return s.commandLogger.Write(ctx, command.BeginTransaction(c.SagaName, c.SagaID, c.TransactionID, c.SagaParams))
	case command.AbortSagaCommand:
		return s.commandLogger.Write(ctx, command.EndTransactionCompensate(c.SagaName, c.SagaID, c.TransactionID, saga.Compensation(c.TransactionID), c.SagaParams))
	case command.EndTransactionCommand:
		nextTransaction := saga.Next(c.TransactionID)
		if c.CompensationID != "" {
			nextTransaction = c.CompensationID
		}

		if nextTransaction == "" {
			return s.commandLogger.Write(ctx, command.EndSaga(c.SagaName, c.SagaID))
		}

		return s.commandLogger.Write(ctx, command.BeginTransaction(c.SagaName, c.SagaID, nextTransaction, c.SagaParams))
	case command.EndSagaCommand:
		return nil
	default:
		return fmt.Errorf("unknow command %d", c.Name)
	}
}
