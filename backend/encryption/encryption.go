package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
)

type EncryptedPackage struct {
	EncryptedHexString string
	NonceHexString     string
}

func GetNewAESCipher(encryptionKey string) (cipher.AEAD, error) {
	aesKeyBytes := []byte(encryptionKey)
	if len(aesKeyBytes) != 32 {
		return nil, errors.New("invalid AES key size: must be exactly 32 bytes, is " + strconv.Itoa(len(aesKeyBytes)) + " instead")
	}
	// Create cipher block
	aesCipherBlock, err := aes.NewCipher(aesKeyBytes)
	if err != nil {
		return nil, errors.Join(errors.New("could not create AES cipher"), err)
	}
	// Wrap cipher block in Galois Counter Mode, a type of [Authenticated Encryption with Associated Data](https://en.wikipedia.org/wiki/Authenticated_encryption)
	aesgcm, err := cipher.NewGCM(aesCipherBlock)
	if err != nil {
		return nil, errors.Join(errors.New("could not wrap AES block in Galois Counter Mode"), err)
	}
	// IMPORTANT: ONLY generate nonce (number only used once) values when they will be immediately used and discarded by the GCM.
	// They should NOT be stored or generated here; it would create a security vulnerability

	return aesgcm, nil
}

// EncryptStringAESGCM handles the complexities of AES GCM encryption for you, returning an encrypted string of bytes in hex format
//
// Input:
//   - aesgcm: the authenticated encryption object
//   - textToEncrypt: a string you would like to encrypt
//
// Output:
//   - *EncryptedOutput: the encrypted result of your text input, including a nonce (number used only once)
//   - error: non-nil if an error occurred while encrypting your data. String will be empty in this case
func EncryptStringAESGCM(aesgcm cipher.AEAD, textToEncrypt string) (*EncryptedPackage, error) {
	bytesToEncrypt := []byte(textToEncrypt)

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Join(errors.New("failed to create nonce for AES GCM encryption"), err)
	}

	// Encrypt data
	encryptedBytes := aesgcm.Seal(nil, nonce, bytesToEncrypt, nil)

	// Convert raw encrypted bytes to hexadecimal
	encryptedText := hex.EncodeToString(encryptedBytes)

	// Package encrypted data with nonce (also converted to hex)
	encryptedOutput := &EncryptedPackage{
		EncryptedHexString: encryptedText,
		NonceHexString:     hex.EncodeToString(nonce),
	}

	return encryptedOutput, nil
}

// DecryptStringAESGCM handles the complexities of AES GCM decryption for you, returning a decrypted string
//
// Input:
//   - aesgcm: the authenticated encryption object
//   - encryptedPackage: a package of an encrypted hex string to decode and a nonce
//
// Output:
//   - string: the decrypted result of your hexadecimal input, stored in plaintext format
//   - error: non-nil if an error occurred while decrypting your data. String will be empty in this case
func DecryptStringAESGCM(aesgcm cipher.AEAD, encryptedPacakge *EncryptedPackage) (string, error) {
	// Get raw encrypted bytes from hexadecimal format
	bytesToDecrypt, err := hex.DecodeString(encryptedPacakge.EncryptedHexString)
	if err != nil {
		return "", errors.Join(errors.New("couldn't convert text to bytes, text must be in hexadecimal format"), err)
	}

	// Recreate nonce bytes from package hexadecimal
	nonce, err := hex.DecodeString(encryptedPacakge.NonceHexString)
	if err != nil {
		return "", errors.Join(errors.New("couldn't recreate nonce from hex string during decryption"), err)
	}

	// Decrypt data
	decryptedBytes, err := aesgcm.Open(nil, nonce, bytesToDecrypt, nil)
	if err != nil {
		return "", errors.Join(errors.New("couldn't open decrypted bytes, check that key and nonce are correct"), err)
	}

	// Convert raw decrypted bytes to plaintext
	decryptedText := string(decryptedBytes)

	return decryptedText, nil
}
