package hashx

import "crypto/sha256"

func HashSHA256(bytes []byte) []byte {
	h := sha256.New()
	h.Write(bytes)
	bs := h.Sum(nil)
	return bs
}
