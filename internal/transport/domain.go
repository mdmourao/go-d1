package transport

type Payload struct {
	SQL    string `json:"sql"`
	Args   []any  `json:"args"`
	IsExec bool   `json:"is_exec"`
}

// response
// https://developers.cloudflare.com/d1/worker-api/prepared-statements/#run

/*
{
  "success": true,
  "meta": {
    "served_by": "v3-prod",
    "served_by_region": "WEUR",
    "served_by_colo": "LIS",
    "served_by_primary": true,
    "timings": {
      "sql_duration_ms": 0.3846
    },
    "duration": 0.3846,
    "changes": 2,
    "last_row_id": 0,
    "changed_db": true,
    "size_after": 20480,
    "rows_read": 14,
    "rows_written": 2,
    "total_attempts": 1
  },
  "results": []
}
*/

type Response struct {
	Changes   int64 `json:"changes"`
	LastRowID int64 `json:"last_row_id"`
}
