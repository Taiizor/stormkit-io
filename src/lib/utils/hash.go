package utils

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandomToken returns a random token based on the given length. It builds the token
// from lowercase, uppercase English characters and integers.
func RandomToken(n int) string {
	src := rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// StringToID takes a string number as an argument, parses it
// and returns a types.ID value.
func StringToID(id string) types.ID {
	idInt, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return 0
	}

	return types.ID(idInt)
}
