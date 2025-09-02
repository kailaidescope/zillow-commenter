// The encryption package provides functions for easy encryption using AES without having to worry about the details.
//
// Based on a [blog post](https://www.twilio.com/en-us/blog/developers/community/encrypt-and-decrypt-data-in-go-with-aes-256) by twilio
package encryption

import (
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// ================================================================================================================= //
//                                             Helper Functions                                                      //
// ================================================================================================================= //

// SetupAndTeardownWithRealKey initializes a cipher with the zillowette encryption key and returns a cleanup function.
func SetupAndTeardownWithRealKey(tb testing.TB) (func(tb testing.TB), cipher.AEAD, error) {
	// Retrieve AES key from environment variables
	os.Chdir("..")
	godotenv.Load()
	aesKey := os.Getenv("ZILLOWETTE_AES_KEY")

	return SetupAndTeardown(aesKey, tb)
}

// SetupAndTeardown initializes a cipher with the specified key and returns a cleanup function.
func SetupAndTeardown(aesKey string, tb testing.TB) (func(tb testing.TB), cipher.AEAD, error) {

	// Generate new cipher
	aesCipher, err := GetNewAESCipher(aesKey)
	if err != nil {
		return nil, nil, errors.Join(errors.New("failed to generate aes cipher object for testing"), err)
	}

	// Return teardown function
	return func(tb testing.TB) {
		tb.Log("Connection to server closed")
	}, aesCipher, nil
}

// Returns random letter in the alphabet, upper or lower case
func getRandomAlpha() rune {
	letterIndex := rand.Intn(52)

	return rune(65 + (int(letterIndex/26) * 32) + (letterIndex % 26))
}

// ================================================================================================================= //
//                                             Generation Tests                                                      //
// ================================================================================================================= //

func TestGetNewAESCipher_InvalidKeySize(t *testing.T) {
	invalidKeys := []string{
		"",                                  // empty
		"shortkey",                          // too short
		"123456789012345678901234567890",    // 30 bytes
		"123456789012345678901234567890123", // 33 bytes
	}

	for _, key := range invalidKeys {
		_, err := GetNewAESCipher(key)
		if err == nil {
			t.Errorf("Expected error for key of length %d, got nil", len(key))
		}
	}
}

func TestGetNewAESCipher_ValidKeySize(t *testing.T) {
	validKey := "12345678901234567890123456789012" // 32 bytes
	aesgcm, err := GetNewAESCipher(validKey)
	if err != nil {
		t.Fatalf("Expected no error for valid key, got: %v", err)
	}
	if aesgcm == nil {
		t.Fatalf("Expected cipher.AEAD object, got nil")
	}
}

func TestGetNewAESCipher_ValidKey_JustLetters(t *testing.T) {
	key := "abcdefghijklmnopqrstuvwxyzABCDEF" // 32 chars, just letters
	aesgcm, err := GetNewAESCipher(key)
	if err != nil {
		t.Errorf("Expected no error for just letters key, got: %v", err)
	}
	if aesgcm == nil {
		t.Errorf("Expected cipher.AEAD object, got nil")
	}
}

func TestGetNewAESCipher_ValidKey_Alphanumeric(t *testing.T) {
	key := "abc123ABC456def789GHI012jkl345MN" // 32 chars, alphanumeric
	aesgcm, err := GetNewAESCipher(key)
	if err != nil {
		t.Errorf("Expected no error for alphanumeric key, got: %v", err)
	}
	if aesgcm == nil {
		t.Errorf("Expected cipher.AEAD object, got nil")
	}
}

func TestGetNewAESCipher_ValidKey_PrintableASCII(t *testing.T) {
	key := "!\\\"#$%&'()*+,-./0123456789:;<=>?" // 32 chars, printable ASCII
	aesgcm, err := GetNewAESCipher(key)
	if err != nil {
		t.Errorf("Expected no error for printable ASCII key, got: %v", err)
	}
	if aesgcm == nil {
		t.Errorf("Expected cipher.AEAD object, got nil")
	}
}

func TestGetNewAESCipher_ValidKey_AllASCII(t *testing.T) {
	// ASCII 0-31 are control chars, but still valid as bytes
	key := string([]byte{
		0, 1, 2, 3, 4, 5, 6, 7,
		8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23,
		24, 25, 26, 27, 28, 29, 30, 31,
	})
	aesgcm, err := GetNewAESCipher(key)
	if err != nil {
		t.Errorf("Expected no error for all ASCII key, got: %v", err)
	}
	if aesgcm == nil {
		t.Errorf("Expected cipher.AEAD object, got nil")
	}
}

// ================================================================================================================= //
//                                       Encryption/Decryption Tests                                                 //
// ================================================================================================================= //

func TestEncryptStringAESGCM_Valid_Alpha(t *testing.T) {
	cleanup, aesgcm, err := SetupAndTeardownWithRealKey(t)
	if err != nil {
		t.Fatalf("SetupAndTeardownWithRealKey failed: %v", err)
	}
	defer cleanup(t)

	for i := 0; i < 50; i++ {
		length := rand.Intn(371) + 30 // random length between 30 and 400
		b := make([]rune, length)
		for j := range b {
			b[j] = getRandomAlpha()
		}
		testString := string(b)
		testEncryptDecryptRoundTrip(t, aesgcm, testString)
	}
}

func TestEncryptStringAESGCM_Valid_Alphanumeric(t *testing.T) {
	cleanup, aesgcm, err := SetupAndTeardownWithRealKey(t)
	if err != nil {
		t.Fatalf("SetupAndTeardownWithRealKey failed: %v", err)
	}
	defer cleanup(t)

	alphanum := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < 50; i++ {
		length := rand.Intn(371) + 30
		b := make([]byte, length)
		for j := range b {
			b[j] = alphanum[rand.Intn(len(alphanum))]
		}
		testString := string(b)
		testEncryptDecryptRoundTrip(t, aesgcm, testString)
	}
}

func TestEncryptStringAESGCM_Valid_PrintableASCII(t *testing.T) {
	cleanup, aesgcm, err := SetupAndTeardownWithRealKey(t)
	if err != nil {
		t.Fatalf("SetupAndTeardownWithRealKey failed: %v", err)
	}
	defer cleanup(t)

	// Printable ASCII: 32 (space) to 126 (~)
	for i := 0; i < 50; i++ {
		length := rand.Intn(371) + 30
		b := make([]byte, length)
		for j := range b {
			b[j] = byte(rand.Intn(95) + 32)
		}
		testString := string(b)
		testEncryptDecryptRoundTrip(t, aesgcm, testString)
	}
}

func TestEncryptStringAESGCM_Valid_AllASCII(t *testing.T) {
	cleanup, aesgcm, err := SetupAndTeardownWithRealKey(t)
	if err != nil {
		t.Fatalf("SetupAndTeardownWithRealKey failed: %v", err)
	}
	defer cleanup(t)

	// All ASCII: 0 to 127
	for i := 0; i < 50; i++ {
		length := rand.Intn(371) + 30
		b := make([]byte, length)
		for j := range b {
			b[j] = byte(rand.Intn(128))
		}
		testString := string(b)
		testEncryptDecryptRoundTrip(t, aesgcm, testString)
	}
}

// Helper function to test encryption and decryption round-trip
func testEncryptDecryptRoundTrip(t *testing.T, aesgcm cipher.AEAD, input string) {
	encryptedPackage, err := EncryptStringAESGCM(aesgcm, input)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	decrypted, err := DecryptStringAESGCM(aesgcm, encryptedPackage)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	if decrypted != input {
		t.Errorf("Round-trip failed: got %q, want %q", decrypted, input)
	}
}

func TestEncryptDecryptRoundTrip_InvalidNonce(t *testing.T) {
	cleanup, aesgcm, err := SetupAndTeardownWithRealKey(t)
	if err != nil {
		t.Fatalf("SetupAndTeardownWithRealKey failed: %v", err)
	}
	defer cleanup(t)

	// Encrypt a test string
	input := "This is a test string for nonce tampering"
	encryptedPackage, err := EncryptStringAESGCM(aesgcm, input)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Convert nonce from hex to bytes
	nonceBytes, err := hex.DecodeString(encryptedPackage.NonceHexString)
	if err != nil {
		t.Fatal("Failed to convert nonce to bytes for tampering")
	}

	// Tamper with the nonce (first N bytes)
	tampered := make([]byte, aesgcm.NonceSize())
	copy(tampered, nonceBytes)
	tampered[0] ^= 0xFF // Flip bits in the first byte of nonce

	// Place tampered nonce in package
	encryptedPackage.NonceHexString = hex.EncodeToString(tampered)

	// Attempt decryption
	_, err = DecryptStringAESGCM(aesgcm, encryptedPackage)
	if err == nil {
		t.Errorf("Expected decryption to fail with tampered nonce, but it succeeded")
	}
}

func TestBasicEncryption(t *testing.T) {
	cleanup, aesgcm, err := SetupAndTeardownWithRealKey(t)
	if err != nil {
		t.Fatalf("SetupAndTeardownWithRealKey failed: %v", err)
	}
	defer cleanup(t)

	input := "192.168.1.1"
	encryptedPackage, err := EncryptStringAESGCM(aesgcm, input)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	log.Println("Encrypted hex string (", len(encryptedPackage.EncryptedHexString), "):", encryptedPackage.EncryptedHexString)
	log.Println("Nonce hex string (", len(encryptedPackage.NonceHexString), "):", encryptedPackage.NonceHexString)
	decrypted, err := DecryptStringAESGCM(aesgcm, encryptedPackage)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	log.Println("Decrypted string (", len(decrypted), "):", decrypted)
	if decrypted != input {
		t.Errorf("Round-trip failed: got %q, want %q", decrypted, input)
	}
	log.Println("Hi there :3 Enjoy some free encryption!")
}
