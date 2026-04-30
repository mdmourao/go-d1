package god1

import "errors"

var (
	ErrTransactionsNotSupported = errors.New("transactions not supported")
	ErrDriverInvalidDSN         = errors.New("invalid DSN")
	ErrDriverMissingDSN         = errors.New("missing DSN")
	ErrDriverInvalidDSNScheme   = errors.New("invalid DSN: scheme must be http or https")
	ErrDriverInvalidDSNHost     = errors.New("invalid DSN: missing host")
)
