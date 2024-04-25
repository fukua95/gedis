package util

import (
	"strconv"
	"strings"
)

func Atoi(b []byte) (int, error) {
	return strconv.Atoi(string(b))
}

func Itoa(v int) []byte {
	return []byte(strconv.Itoa(v))
}

func BytesToLower(b []byte) []byte {
	return []byte(strings.ToLower(string(b)))
}
