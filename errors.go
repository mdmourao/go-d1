package god1

import "errors"

var (
	ErrTransactionsNotSupported           = errors.New("transactions not supported")
	ErrBlobNotSupported                   = errors.New("BLOB parameters are not supported")
	ErrNamedParametersNotSupported        = errors.New("named parameters are not supported")
	ErrInt64OutOfJSSafeRange              = errors.New("int64 parameter exceeds JavaScript safe integer range (\u00b12^53-1)")
	ErrDriverInvalidDSN                   = errors.New("invalid DSN")
	ErrDriverMissingDSN                   = errors.New("missing DSN")
	ErrDriverInvalidDSNScheme             = errors.New("invalid DSN: scheme must be http or https")
	ErrDriverInvalidDSNHost               = errors.New("invalid DSN: missing host")
	ErrColumns                            = errors.New("columns not valid or missing")
	ErrRows                               = errors.New("rows not valid or missing")
	ErrParseRows                          = errors.New("error parsing rows")
	ErrInvalidRequestTimeout              = errors.New("invalid request timeout: must be greater than zero")
	ErrInvalidCloudflareAccessCredentials = errors.New("invalid Cloudflare Access credentials: client ID and secret must be provided")
	ErrInvalidSQLiteVersion               = errors.New("invalid SQLite version: must be a non-empty string")
)
