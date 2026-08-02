package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/chishkin-afk/intask/backend/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager manages JWT token generation and validation.
//
// It uses RSA key pairs for signing and verifying tokens.
// The private key is used to generate tokens, while the public key
// is used to validate their authenticity.
type JWTManager struct {
	cfg     *config.Config
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

// New creates a new JWTManager instance.
//
// It loads RSA private and public keys from paths specified in the configuration.
// Returns an error if any of the keys cannot be loaded or parsed.
func New(cfg *config.Config) (*JWTManager, error) {
	private, err := loadPrivate(cfg.JWT.PrivatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	public, err := loadPublic(cfg.JWT.PublicPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &JWTManager{
		cfg:     cfg,
		private: private,
		public:  public,
	}, nil
}

// Generate creates a signed JWT token for the specified user.
//
// The generated token contains user ID and standard JWT claims,
// including issuer, subject, issued time, and expiration time.
//
// The token is signed using the RSA private key.
func (jm *JWTManager) Generate(userID uuid.UUID) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, customClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "intask",
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.cfg.JWT.TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	signedToken, err := token.SignedString(jm.private)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// Validate verifies a JWT token and extracts the user ID from its claims.
//
// The token signature is validated using the RSA public key.
// Returns ErrInvalidToken if the token is invalid or cannot be verified.
func (jm *JWTManager) Validate(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, errs.ErrInvalidToken
		}

		return jm.public, nil
	})
	if err != nil {
		return uuid.Nil, errs.ErrInvalidToken
	}

	if !token.Valid {
		return uuid.Nil, errs.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*customClaims); ok {
		return claims.UserID, nil
	}

	return uuid.Nil, errs.ErrInvalidToken
}

func loadPrivate(path string) (*rsa.PrivateKey, error) {
	path = filepath.Clean(path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPrivateKeyFromPEM(bytes)
}

func loadPublic(path string) (*rsa.PublicKey, error) {
	path = filepath.Clean(path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(bytes)
}
