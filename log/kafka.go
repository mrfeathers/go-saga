package log

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mrfeathers/go-saga/command"
	"github.com/segmentio/kafka-go"
)

// KafkaCommandLog is a struct that provides a mechanism to log saga commands sent through Kafka.
type KafkaCommandLog struct {
	writer *kafka.Writer
	reader *kafka.Reader

	msgs map[command.Command]kafka.Message
}

func NewKafkaCommandLog(writer *kafka.Writer, reader *kafka.Reader) *KafkaCommandLog {
	return &KafkaCommandLog{writer: writer, reader: reader, msgs: make(map[command.Command]kafka.Message)}
}

// Commit commits the specified command to Kafka.
func (k *KafkaCommandLog) Commit(ctx context.Context, c command.Command) error {
	msg, ok := k.msgs[c]
	if !ok {
		return fmt.Errorf("no message to commit for command %s", c.ID)
	}

	err := k.reader.CommitMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed command commiting: %v", err)
	}

	delete(k.msgs, c)

	return nil
}

// Read reads the next command from Kafka.
func (k *KafkaCommandLog) Read(ctx context.Context) (command.Command, error) {
	msg, err := k.reader.FetchMessage(ctx)
	if err != nil {
		return command.Command{}, err
	}

	var c command.Command
	err = json.Unmarshal(msg.Value, &c)
	if err != nil {
		return command.Command{}, fmt.Errorf("failed command unmarshalling: %v", err)
	}
	k.msgs[c] = msg

	return c, nil
}

// Write writes the specified command to Kafka topic with partitioning by saga id.
func (k *KafkaCommandLog) Write(ctx context.Context, c command.Command) error {
	v, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed command marshalling: %v", err)
	}
	msg := kafka.Message{
		Key:   []byte(c.SagaID),
		Value: v,
		Time:  c.CreatedAt,
	}

	err = k.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed command producing: %v", err)
	}

	return nil
}
