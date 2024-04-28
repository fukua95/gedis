package server

import (
	"encoding/base64"
	"log"
)

var (
	// the contents of an empty rdb file in base64
	emptyRdbBase64 = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
)

func EmptyRdb() []byte {
	rdb, err := base64.StdEncoding.DecodeString(emptyRdbBase64)
	if err != nil {
		log.Fatalln(err)
	}
	return rdb
}
