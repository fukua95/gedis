package util

import (
	"strconv"
	"strings"
)

func Atoi(b []byte) (int, error) {
	return strconv.Atoi(string(b))
}

func Itoa[T int | int64](v T) []byte {
	return []byte(strconv.Itoa(int(v)))
}

func BytesToLower(b []byte) []byte {
	return []byte(strings.ToLower(string(b)))
}
