package redisloadtest

import (
	"crypto/rand"
	"fmt"
	"github.com/google/uuid"
	"math/big"
	"time"
)

const (
	// 90 days
	ExpiresSeconds = 90 * 24 * 60 * 60 * time.Second
	letters        = "abcdefghijklmnopqrstuvwxyz"
	//letters = "01"
)

func GenerateKey(appbundleLength int) (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate uid: %w", err)
	}
	appBundle, err := RandomString(appbundleLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	return uid.String() + appBundle, nil
}

func RandomString(n int) (string, error) {
	result := make([]byte, n)
	for i := range result {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}
		result[i] = letters[randomIndex.Int64()]
	}
	return string(result), nil
}
