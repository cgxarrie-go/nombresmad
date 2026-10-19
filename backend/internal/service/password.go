package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

// PasswordIterations is the PBKDF2-HMAC-SHA256 work factor for new hashes.
// Tests lower it. Stored hashes keep the iteration count they were created with.
var PasswordIterations = 600_000

const (
	passwordScheme    = "pbkdf2-sha256"
	passwordKeyLen    = 32
	passwordSaltLen   = 16
	maxHashIterations = 5_000_000
)

// HashPassword returns a one-way PBKDF2-HMAC-SHA256 hash. The password cannot be recovered from it.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("empty password")
	}
	if PasswordIterations < 1 || PasswordIterations > maxHashIterations {
		return "", errors.New("invalid password iteration count")
	}
	salt := make([]byte, passwordSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := pbkdf2SHA256([]byte(password), salt, PasswordIterations, passwordKeyLen)
	return passwordScheme + "$" + strconv.Itoa(PasswordIterations) + "$" +
		base64.RawURLEncoding.EncodeToString(salt) + "$" +
		base64.RawURLEncoding.EncodeToString(sum), nil
}

// CheckPassword reports whether password matches a HashPassword result.
func CheckPassword(hash, password string) bool {
	scheme, iterText, saltText, sumText, ok := splitHash(hash)
	if !ok || scheme != passwordScheme || password == "" {
		return false
	}
	iterations, err := strconv.Atoi(iterText)
	if err != nil || iterations < 1 || iterations > maxHashIterations {
		return false
	}
	salt, err := base64.RawURLEncoding.DecodeString(saltText)
	if err != nil || len(salt) < 8 {
		return false
	}
	want, err := base64.RawURLEncoding.DecodeString(sumText)
	if err != nil || len(want) == 0 {
		return false
	}
	got := pbkdf2SHA256([]byte(password), salt, iterations, len(want))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func splitHash(hash string) (scheme, iterations, salt, sum string, ok bool) {
	parts := strings.Split(hash, "$")
	if len(parts) != 4 {
		return "", "", "", "", false
	}
	return parts[0], parts[1], parts[2], parts[3], true
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hLen := sha256.Size
	blocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, blocks*hLen)
	for i := 1; i <= blocks; i++ {
		out = append(out, pbkdf2Block(password, salt, iterations, i)...)
	}
	return out[:keyLen]
}

func pbkdf2Block(password, salt []byte, iterations, blockNum int) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)
	var counter [4]byte
	binary.BigEndian.PutUint32(counter[:], uint32(blockNum))
	mac.Write(counter[:])
	u := mac.Sum(nil)
	result := append([]byte(nil), u...)
	for i := 1; i < iterations; i++ {
		mac.Reset()
		mac.Write(u)
		u = mac.Sum(nil)
		for j := range result {
			result[j] ^= u[j]
		}
	}
	return result
}
