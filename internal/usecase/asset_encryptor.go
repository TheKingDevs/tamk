package usecase

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheKingDevs/tamk/pkg/logger"
)

const (
	AssetEncryptionMagic = "TAMK_ENC_1" // Magic header for encrypted files
	SaltSize             = 16
	IVSize               = 12 // AES-GCM IV
)

// AssetEncryptor handles encryption/decryption of project assets.
type AssetEncryptor struct{}

// EncryptAssets encrypts all files in the assets directory.
func (e *AssetEncryptor) EncryptAssets(projectPath, password string) error {
	assetsDir := filepath.Join(projectPath, "src", "main", "assets")
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return nil // No assets to encrypt
	}

	key, salt := deriveEncryptionKey(password)

	encryptedCount := 0
	err := filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip already encrypted files (check magic header)
		if isEncryptedFile(path) {
			return nil
		}

		// Skip gitignore and other dotfiles
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		if err := e.encryptFile(path, key, salt); err != nil {
			logger.Warn(fmt.Sprintf("Failed to encrypt asset %s: %v", path, err))
			return nil // Non-fatal
		}

		encryptedCount++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to encrypt assets: %w", err)
	}

	if encryptedCount > 0 {
		logger.Info(fmt.Sprintf("Encrypted assets: %d", encryptedCount))
	}

	return nil
}

// DecryptAsset decrypts a single encrypted asset file.
func (e *AssetEncryptor) DecryptAsset(encryptedPath, password string) ([]byte, error) {
	data, err := os.ReadFile(encryptedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted file: %w", err)
	}

	if len(data) < len(AssetEncryptionMagic)+SaltSize+IVSize {
		return nil, fmt.Errorf("invalid encrypted file format")
	}

	// Verify magic header
	if string(data[:len(AssetEncryptionMagic)]) != AssetEncryptionMagic {
		return nil, fmt.Errorf("not an encrypted TAMK asset")
	}

	offset := len(AssetEncryptionMagic)

	// Extract salt and IV
	salt := data[offset : offset+SaltSize]
	offset += SaltSize

	iv := data[offset : offset+IVSize]
	offset += IVSize

	ciphertext := data[offset:]

	// Derive key from password and salt
	key := deriveKeyFromSalt(password, salt)

	// Decrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// GetEncryptionHash returns a hash of the encrypted assets for integrity checks.
func (e *AssetEncryptor) GetEncryptionHash(projectPath string) (string, error) {
	assetsDir := filepath.Join(projectPath, "src", "main", "assets")
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return "", nil
	}

	hasher := sha256.New()

	err := filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		hasher.Write([]byte(info.Name()))
		hasher.Write([]byte{0})
		return nil
	})

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (e *AssetEncryptor) encryptFile(path string, key, salt []byte) error {
	plaintext, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Generate random IV
	iv := make([]byte, IVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nil, iv, plaintext, nil)

	// Build output: magic + salt + iv + ciphertext
	output := make([]byte, 0, len(AssetEncryptionMagic)+SaltSize+IVSize+len(ciphertext))
	output = append(output, []byte(AssetEncryptionMagic)...)
	output = append(output, salt...)
	output = append(output, iv...)
	output = append(output, ciphertext...)

	// Write encrypted data to same path (keeps original filename)
	return os.WriteFile(path, output, 0o644)
}

func isEncryptedFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil || len(data) < len(AssetEncryptionMagic) {
		return false
	}
	return string(data[:len(AssetEncryptionMagic)]) == AssetEncryptionMagic
}

func deriveEncryptionKey(password string) (key, salt []byte) {
	salt = make([]byte, SaltSize)
	rand.Read(salt)
	key = deriveKeyFromSalt(password, salt)
	return
}

func deriveKeyFromSalt(password string, salt []byte) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hasher.Write(salt)
	return hasher.Sum(nil)[:32] // AES-256
}
