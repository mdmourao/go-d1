package utils

import (
	"fmt"
	"time"
)

func UniqueTable(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
