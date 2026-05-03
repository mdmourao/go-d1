package god1

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloseStmt(t *testing.T) {
	s := &stmt{}
	err := s.Close()
	assert.NoError(t, err)
}

func TestNumInputStmt(t *testing.T) {
	s := &stmt{}
	n := s.NumInput()
	assert.Equal(t, -1, n)
}

func TestValuesToNamed(t *testing.T) {
	args := []driver.Value{1, "test", true}
	named := valuesToNamed(args)
	assert.Len(t, named, len(args))
	for i, v := range args {
		assert.Equal(t, i+1, named[i].Ordinal)
		assert.Equal(t, v, named[i].Value)
	}
}
