package god1

import "database/sql/driver"

// https://pkg.go.dev/database/sql/driver#Tx

var _ driver.Tx = Tx{}

type Tx struct{}

// TODO nil
func (Tx) Commit() error   { return nil }
func (Tx) Rollback() error { return nil }
