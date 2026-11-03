package recipes

import (
	"crypto/rand"
	"encoding/hex"
)

func newUUID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return hex.EncodeToString(id[0:4]) + "-" +
		hex.EncodeToString(id[4:6]) + "-" +
		hex.EncodeToString(id[6:8]) + "-" +
		hex.EncodeToString(id[8:10]) + "-" +
		hex.EncodeToString(id[10:16]), nil
}
