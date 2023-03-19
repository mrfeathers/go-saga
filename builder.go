package saga

// Builder is used to construct a Saga.
type Builder struct {
	sagaName      string
	transactions  []Transaction
	compensations []Compensation
}

// TransactionBuilder is used to build a Transaction and add it to the Saga.
type TransactionBuilder struct {
	builder Builder
	index   int
	t       Transaction
}

// SavePoint marks the current Transaction as a SavePoint.
// SavePoint is used to roll back the Saga to this point if a failure occurs.
func (tb TransactionBuilder) SavePoint() TransactionBuilder {
	tb.t.IsSavePoint = true
	return tb
}

// WithCompensation adds a compensation function to the current Transaction and returns the Saga Builder.
// The compensation function is used to undo the effects of the current Transaction if a failure occurs later in the Saga.
// Note that Compensation must be idempotent and must not return ErrAbortSaga.
func (tb TransactionBuilder) WithCompensation(name string, f TransactionFunc) Builder {
	c := Compensation{
		Name: name,
		Func: f,
	}
	tb.t.CompensationName = c.Name

	// If there are any previous Transactions with compensation functions,
	// set the next compensation function in the chain to the current compensation function.
	if tb.index > 0 {
		for i := tb.index - 1; i >= 0; i-- {
			if tb.builder.transactions[i].CompensationName != "" {
				c.NextCompensationName = tb.builder.transactions[i].CompensationName
				break
			}
		}
	}

	// Add the Transaction and Compensation to the Saga.
	tb.builder.transactions = append(tb.builder.transactions, tb.t)
	tb.builder.compensations = append(tb.builder.compensations, c)

	return tb.builder
}

// NoCompensation adds the current Transaction to the Saga without a compensation function and returns the Saga Builder.
// Use this method for Transactions that are not reversible.
func (tb TransactionBuilder) NoCompensation() Builder {
	tb.builder.transactions = append(tb.builder.transactions, tb.t)
	tb.builder.compensations = append(tb.builder.compensations, Compensation{Name: ""})

	return tb.builder
}

// New returns a new Saga Builder with the given name.
func New(name string) Builder {
	return Builder{sagaName: name, transactions: []Transaction{}, compensations: []Compensation{}}
}

// Begin creates a new Transaction and returns a TransactionBuilder for it.
// This method should be used for the first Transaction in the Saga.
func (b Builder) Begin(name string, f TransactionFunc) TransactionBuilder {
	return TransactionBuilder{
		builder: b,
		index:   0,
		t:       Transaction{Name: name, Func: f},
	}
}

// Then creates a new Transaction and returns a TransactionBuilder for it.
// This method should be used for all Transactions in the Saga after the first one.
func (b Builder) Then(name string, f TransactionFunc) TransactionBuilder {
	t := Transaction{
		Name: name,
		Func: f,
	}

	transactionLen := len(b.transactions)
	index := transactionLen
	if transactionLen > 0 {
		b.transactions[transactionLen-1].NextTransactionName = t.Name
	}

	return TransactionBuilder{
		builder: b,
		index:   index,
		t:       t,
	}
}

// End creates a new Saga with the current transactions and compensations.
func (b Builder) End() Saga {
	if len(b.transactions) == 0 {
		return Saga{name: b.sagaName}
	}

	transactions := make(map[string]Transaction)
	for _, t := range b.transactions {
		transactions[t.Name] = t
	}

	compensations := make(map[string]Compensation)
	for _, c := range b.compensations {
		if c.Name == "" {
			continue
		}
		compensations[c.Name] = c
	}

	return Saga{
		name:             b.sagaName,
		firstTransaction: b.transactions[0].Name,
		transactions:     transactions,
		compensations:    compensations,
	}
}
