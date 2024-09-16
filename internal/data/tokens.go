package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"morgan.greenlight.nex/internal/validator"
)

type TokenScope string

const (
	ScopeActivation     TokenScope = "activation"
	ScopeAuthentication TokenScope = "authentication"
)

// Token struct to hold the data for an individual token. This includes the
// plaintext and hashed versions of the token, associated user ID, expiry time and
// scope.
type Token struct {
	Plaintext string     `json:"token"`
	Hash      []byte     `json:"-"`
	UserID    int64      `json:"-"`
	Expiry    time.Time  `json:"expiry"`
	Scope     TokenScope `json:"-"`
}

func generateToken(userId int64, ttl time.Duration, scope TokenScope) (*Token, error) {
	token := &Token{
		UserID: userId,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	token.Hash = hashToken(token.Plaintext)

	return token, nil

}

func hashToken(plainText string) []byte {
	ar := sha256.Sum256([]byte(plainText))
	return ar[:]
}
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext == "", "token", "must be provided")
	v.Check(len(tokenPlaintext) != 26, "token", "must be 26 bytes long")
}
