package hashx

import (
	"crypto/md5"
	"crypto/sha256"
)

func HashSHA256(bytes []byte) []byte {
	h := sha256.New()
	h.Write(bytes)
	bs := h.Sum(nil)
	return bs
}

func HashMD5(bytes []byte) []byte {
	h := md5.New()
	h.Write(bytes)
	bs := h.Sum(nil)
	return bs
}
