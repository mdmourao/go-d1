package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Tx

var _ driver.Tx = Tx{}

type Tx struct{}

func (Tx) Commit() error   { return ErrTransactionsNotSupported }
func (Tx) Rollback() error { return ErrTransactionsNotSupported }
