package main

import (
	"context"
	"fmt"
	"github.com/mrfeathers/go-saga"
	"github.com/mrfeathers/go-saga/command"
	cmdlog "github.com/mrfeathers/go-saga/log"
	"github.com/segmentio/kafka-go"
	"log"
)

type ErrLogger struct{}

func (l ErrLogger) Log(err error) {
	log.Println(err)
}

func main() {
	ctx := context.Background()
	// Run docker compose with kafka to connect locally
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:29092", "saga-commands", 0)
	if err != nil {
		panic(err.Error())
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:29092"},
		Topic:   "saga-commands",
		GroupID: "command-log",
	})
	w := &kafka.Writer{
		Addr:  kafka.TCP("localhost:29092"),
		Topic: "saga-commands",
	}
	defer func() {
		r.Close()
		w.Close()
		conn.Close()
	}()

	kafkaCommandLog := cmdlog.NewKafkaCommandLog(w, r)

	// create a Saga definition
	hotelBookingSaga := saga.New("HotelBooking").
		Begin("BookHotel", func(ctx context.Context, params any) error {
			fmt.Println("T1: hotel booked")
			return nil
		}).
		WithCompensation("CancelHotelBooking", func(ctx context.Context, params any) error {
			fmt.Println("C1: hotel canceled")
			return nil
		}).
		Then("PayHotelBooking", func(ctx context.Context, params any) error {
			fmt.Println("T2: hotel payed")
			return nil
		}).
		WithCompensation("RefundHotelBooking", func(ctx context.Context, params any) error {
			fmt.Println("C2: hotel refunded")
			return nil
		}).
		Then("BookTransfer", func(ctx context.Context, params any) error {
			fmt.Println("T3: transfer booked")
			return nil
		}).
		WithCompensation("CancelTransferBooking", func(ctx context.Context, params any) error {
			fmt.Println("C3: transfer canceled")
			return nil
		}).
		Then("PayTransferBooking", func(ctx context.Context, params any) error {
			fmt.Println("T4: transfer payed with error")
			return fmt.Errorf("ooops: %w", saga.ErrAbortSaga)
		}).
		WithCompensation("RefundTransferBooking", func(ctx context.Context, params any) error {
			fmt.Println("C4: transfer refunded")
			return nil
		}).
		End()

	// create SEC
	sec := saga.NewSEC(kafkaCommandLog, ErrLogger{})
	// register your Saga in SEC
	sec.RegisterSaga(hotelBookingSaga)

	// write BeginCommand to command log
	err = kafkaCommandLog.Write(ctx, command.BeginSaga(hotelBookingSaga.Name(), nil))
	if err != nil {
		panic(err)
	}

	// run SEC
	sec.Start(ctx)
}
