package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
)

var ReservedWords = map[string]bool{
	"api":       true,
	"login":     true,
	"register":  true,
	"dashboard": true,
	"swagger":   true,
	"health":    true,
}

var SlugRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

const Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	charsetLen := big.NewInt(int64(len(Charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		bytes[i] = Charset[num.Int64()]
	}
	return string(bytes), nil
}

func IsReservedSlug(slug string) bool {
	return ReservedWords[strings.ToLower(strings.TrimSpace(slug))]
}

func IsValidSlugFormat(slug string) bool {
	return SlugRegex.MatchString(slug)
}

