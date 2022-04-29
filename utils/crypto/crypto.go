package crypto

import (
	"github.com/google/uuid"
)

func SecureRandom() string {
	return uuid.New().String()
}
