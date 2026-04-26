package transport

type Payload struct {
	SQL  string `json:"sql"`
	Args []any  `json:"args"`
}

// response
// https://developers.cloudflare.com/d1/worker-api/prepared-statements/#run
