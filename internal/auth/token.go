// Package auth implements Laravel Sanctum compatible personal access tokens
// and bcrypt password hashing, so tokens and passwords created by the Laravel
// app keep working (and vice versa).
package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"time"

	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/support"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidToken = errors.New("invalid token")

type Service struct {
	DB          *gorm.DB
	TokenPrefix string
	BcryptCost  int
}

// GenerateTokenString mirrors Sanctum's HasApiTokens::generateTokenString:
// prefix + 40 random chars + crc32b(random chars).
func GenerateTokenString(prefix string) string {
	entropy := support.RandomString(40)
	return fmt.Sprintf("%s%s%08x", prefix, entropy, crc32.ChecksumIEEE([]byte(entropy)))
}

// HashToken is hash('sha256', $plain).
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// CreateToken stores a new token for the user and returns "{id}|{plain}".
func (s *Service) CreateToken(userID, name string) (string, error) {
	plain := GenerateTokenString(s.TokenPrefix)
	abilities := `["*"]`
	t := models.PersonalAccessToken{
		TokenableType: models.UserMorphType,
		TokenableID:   userID,
		Name:          name,
		Token:         HashToken(plain),
		Abilities:     &abilities,
	}
	if err := s.DB.Create(&t).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%d|%s", t.ID, plain), nil
}

// Authenticate resolves a bearer token to its user, like Sanctum's guard.
func (s *Service) Authenticate(bearer string) (*models.User, *models.PersonalAccessToken, error) {
	if bearer == "" {
		return nil, nil, ErrInvalidToken
	}

	var tok models.PersonalAccessToken
	if id, plain, ok := strings.Cut(bearer, "|"); ok {
		n, err := strconv.ParseUint(id, 10, 64)
		if err != nil || plain == "" {
			return nil, nil, ErrInvalidToken
		}
		if err := s.DB.First(&tok, n).Error; err != nil {
			return nil, nil, ErrInvalidToken
		}
		if subtle.ConstantTimeCompare([]byte(tok.Token), []byte(HashToken(plain))) != 1 {
			return nil, nil, ErrInvalidToken
		}
	} else if err := s.DB.Where("token = ?", HashToken(bearer)).First(&tok).Error; err != nil {
		return nil, nil, ErrInvalidToken
	}

	if tok.TokenableType != models.UserMorphType {
		return nil, nil, ErrInvalidToken
	}
	if tok.ExpiresAt != nil && tok.ExpiresAt.Std().Before(time.Now()) {
		return nil, nil, ErrInvalidToken
	}

	var user models.User
	if err := s.DB.First(&user, "id = ?", tok.TokenableID).Error; err != nil {
		return nil, nil, ErrInvalidToken
	}

	now := models.Now()
	s.DB.Model(&tok).Updates(map[string]any{"last_used_at": now, "updated_at": now})

	return &user, &tok, nil
}

// HashPassword hashes with bcrypt using the "$2y$" prefix PHP expects.
func (s *Service) HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), s.BcryptCost)
	if err != nil {
		return "", err
	}
	out := string(h)
	if strings.HasPrefix(out, "$2a$") {
		out = "$2y$" + out[4:]
	}
	return out, nil
}

// CheckPassword verifies a bcrypt hash ($2y$, $2a$ or $2b$).
func CheckPassword(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
