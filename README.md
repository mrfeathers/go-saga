# go-saga

## Saga Pattern GoLang + Kafka implementation 

This package is a demonstration of a possible implementation of the Saga Pattern for my public speech. 
The package is structured in the following way:

- `command`: package that contains a command structure and constants
- `example`: usage example with simple Hotel Booking Saga
- `log`: package that contains kafka implementation of command log
- `docker-compose.yaml`: a docker-compose file that can be used to run the example (it has kafka)
- `saga.go`: contains a Saga structure
- `builder.go`: contains a Builder struct that allows you to define a Saga by chaining the transactions and compensations
- `sec.go`: contains the implementation of the Saga Execution Coordinator (SEC). The SEC coordinates the execution of sagas by invoking the appropriate commands and compensations, based on the outcome of each transaction.

## Usage

To use this package, import it into your Golang project and follow the example from the `example` folder.
```cli
go get github.com/mrfeathers/go-saga
```

## License

This package is released under the MIT License. See the LICENSE file for more information.
