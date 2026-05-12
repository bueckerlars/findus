package secrets

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const jwtSecretFile = ".jwt_secret"

// LoadOrCreateJWTSecret returns a secret of at least 32 bytes from env, file, or newly generated.
func LoadOrCreateJWTSecret(dataDir, fromEnv string) ([]byte, error) {
	if s := strings.TrimSpace(fromEnv); s != "" {
		if len(s) < 32 {
			return nil, fmt.Errorf("FINDUS_JWT_SECRET must be at least 32 characters")
		}
		return []byte(s), nil
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	p := filepath.Join(dataDir, jwtSecretFile)
	if b, err := os.ReadFile(p); err == nil {
		s := strings.TrimSpace(string(b))
		if len(s) >= 32 {
			return []byte(s), nil
		}
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	secret := hex.EncodeToString(buf)
	if err := os.WriteFile(p, []byte(secret+"\n"), 0o600); err != nil {
		return nil, err
	}
	return []byte(secret), nil
}
