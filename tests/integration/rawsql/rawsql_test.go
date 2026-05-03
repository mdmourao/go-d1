package rawsql_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mdmourao/go-d1/tests/integration/internal/testenv"
	"github.com/mdmourao/go-d1/tests/integration/internal/utils"
)

func TestRawSQL_CRUD(t *testing.T) {
	db := testenv.OpenDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	table := utils.UniqueTable("god1_rawsql_users")
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DROP TABLE IF EXISTS "+table)
	})

	createSQL := fmt.Sprintf(
		`CREATE TABLE %s (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			name  TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			active INTEGER NOT NULL DEFAULT 1
		)`, table)
	_, err := db.ExecContext(ctx, createSQL)
	require.NoError(t, err, "create table")

	res, err := db.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (name, email, active) VALUES (?, ?, ?)", table),
		"Gopher", "gopher@example.com", true)
	require.NoError(t, err, "insert")
	if id, err := res.LastInsertId(); err == nil {
		require.Greater(t, id, int64(0), "expected positive last insert id")
	}
	if n, err := res.RowsAffected(); err == nil {
		require.Equal(t, int64(1), n)
	}

	_, err = db.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (name, email) VALUES (?, ?), (?, ?)", table),
		"Duke", "duke@example.com",
		"Tux", "tux@example.com")
	require.NoError(t, err, "bulk insert")

	var name string
	err = db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT name FROM %s WHERE email = ?", table),
		"gopher@example.com").Scan(&name)
	require.NoError(t, err, "select single")
	require.Equal(t, "Gopher", name)

	rows, err := db.QueryContext(ctx,
		fmt.Sprintf("SELECT id, name, email, active FROM %s ORDER BY id", table))
	require.NoError(t, err, "select all")
	defer rows.Close()

	var seen int
	for rows.Next() {
		var id int64
		var n, e string
		var active bool
		require.NoError(t, rows.Scan(&id, &n, &e, &active))
		seen++
	}
	require.NoError(t, rows.Err())
	require.Equal(t, 3, seen)

	res, err = db.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET active = ? WHERE name = ?", table),
		false, "Tux")
	require.NoError(t, err, "update")
	if n, err := res.RowsAffected(); err == nil {
		require.Equal(t, int64(1), n)
	}

	res, err = db.ExecContext(ctx,
		fmt.Sprintf("DELETE FROM %s WHERE name = ?", table),
		"Tux")
	require.NoError(t, err, "delete")
	if n, err := res.RowsAffected(); err == nil {
		require.Equal(t, int64(1), n)
	}

	var count int
	require.NoError(t, db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count))
	require.Equal(t, 2, count)
}
