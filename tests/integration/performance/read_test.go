package performance_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mdmourao/go-d1/tests/integration/internal/testenv"
	"github.com/mdmourao/go-d1/tests/integration/internal/utils"
)

// Run with:
//
//	go test -bench=BenchmarkRead -benchtime=100x -run=^$ ./performance
//	go test -bench=BenchmarkRead -benchtime=10s   -run=^$ ./performance
//	go test -bench=BenchmarkRead -benchtime=100x -cpu=10 -run=^$ ./performance   // parallel
func BenchmarkRead(b *testing.B) {
	db := testenv.OpenDB(b)
	ctx := context.Background()

	table := utils.UniqueTable("god1_read_perf")
	b.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DROP TABLE IF EXISTS "+table) })

	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		`CREATE TABLE %s (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`, table)); err != nil {
		b.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (id, name) VALUES (?, ?)", table), 1, "x"); err != nil {
		b.Fatal(err)
	}

	query := fmt.Sprintf("SELECT name FROM %s WHERE id = ?", table)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var name string
			if err := db.QueryRowContext(ctx, query, 1).Scan(&name); err != nil {
				b.Fatal(err)
			}
		}
	})
}
