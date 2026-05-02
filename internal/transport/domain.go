package transport

type Payload struct {
	SQL    string `json:"sql"`
	Args   []any  `json:"args"`
	IsExec bool   `json:"isExec"`
}

// response
// https://developers.cloudflare.com/d1/worker-api/prepared-statements/#run
type Response struct {
	Changes   int64 `json:"changes"`
	LastRowID int64 `json:"last_row_id"`
}
