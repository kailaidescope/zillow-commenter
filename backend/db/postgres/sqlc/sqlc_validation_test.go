package sqlc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ===================================================================================================================== //
//                                             Setup and Teardown                                                        //
// ===================================================================================================================== //

// ValidationSetupAndTeardown initializes the validator and registers the PostCommentParamsValidation function.
//
// IMPORTANT: This function should be called in each test case to ensure the validator is set up correctly.
//
// Input:
//   - tb: A testing.TB interface that allows the function to log messages and handle test failures.
//
// Output:
//   - A function that can be deferred to perform teardown actions after the test completes.
//   - A pointer to a validator.Validate instance that can be used to validate structs.
func ValidationSetupAndTeardown(tb testing.TB) (func(tb testing.TB), *validator.Validate) {
	// Create a validator singleton
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Register custom validations for structs and fields
	RegisterValidators(validate)

	return func(tb testing.TB) {
		tb.Log("Teardown complete")
	}, validate
}

// ===================================================================================================================== //
//                                                Write tests below                                                      //
// ===================================================================================================================== //

//

// ===================================================================================================================== //
//                                             Validation Test Helpers                                                   //
// ===================================================================================================================== //

// ===================================================================================================================== //
//                                             Validation Tests                                                          //
// ===================================================================================================================== //

func TestPostCommentParamsValidation_Valid(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params, got error: %v", err)
	}
}

// --- COMMENTID ---

func TestPostCommentParamsValidation_CommentID_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.CommentID = pgtype.UUID{} // Zero value, not valid

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentID, got nil")
	}
}

func TestPostCommentParamsValidation_CommentID_AlmostValidUUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// Set valid UUID bytes to be edited
	tempId := params.CommentID.String()
	// Change the version code (at index 14) from '7' to a different digit, changing the
	if len(tempId) > 14 && tempId[14] == '7' {
		tempId = tempId[:14] + "3" + tempId[15:] // Change '7' to '3' for testing
	}
	invalidVersionUUID, err := uuid.Parse(tempId)
	if err != nil {
		t.Error("Failed to parse modified CommentID UUID:", err)
		return
	}
	// Set the modified UUID back to the params
	params.CommentID = pgtype.UUID{Bytes: [16]byte(invalidVersionUUID), Valid: true}

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid CommentID UUID, got nil")
	}
}

func TestPostCommentParamsValidation_CommentID_InvalidUUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// Set invalid UUID bytes (not a valid UUID)
	params.CommentID = pgtype.UUID{Bytes: [16]byte{0, 0, 3}, Valid: true}

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid CommentID UUID, got nil")
	}
}

// --- LISTINGID ---

func getRealZillowListingIDs() []string {
	// This function returns a list of real Zillow Listing IDs.
	return []string{"32698227", "32692760", "32692378"}
}

/* func TestZillowIdExistenceValidator_ListingID_Exists(t *testing.T) {
	err := playwright.Install()
	if err != nil {
		t.Fatal("Failed to install Playwright: ", err)
	}

	for _, listingID := range getRealZillowListingIDs() {
		exists, err := validateZillowListingExistence(listingID)
		if err != nil {
			t.Errorf("Failed to validate Zillow Listing ID '%s': %v", listingID, err)
			continue
		}
		if !exists {
			t.Errorf("Expected Zillow Listing ID '%s' to exist, but it does not", listingID)
		} else {
			t.Logf("Zillow Listing ID '%s' exists as expected", listingID)
		}
	}
} */

func TestPostCommentParamsValidation_ListingID_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_Number(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingID = "abc123" // Not a number

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-numeric ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_ExcludesDot(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingID = "123.456" // Contains a dot

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID containing '.', got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingID = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingID = "123456789012345678901" // 21 chars, max=20

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID with length > 20, got nil")
	}
}

// --- USERIP ---

func TestPostCommentParamsValidation_UserIp_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserIp = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserIp, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_InvalidIP(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserIp = "not_an_encrypted_ip"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid UserIp, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_NotHexadecimal(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// Insert a non-hex character
	params.UserIp = makeStringOfLength(40) + "Z"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserIp with non-hexadecimal character, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_MinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// 39 chars, min=40
	params.UserIp = makeStringOfLength(39)

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserIp with length < 40, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// 131 chars, max=130
	params.UserIp = makeStringOfLength(131)

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserIp with length > 130, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_ValidHexMinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// 40 chars, all hex
	params.UserIp = ""
	for i := 0; i < 40; i++ {
		params.UserIp += "a"
	}

	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid UserIp with min length 40, got error: %v", err)
	}
}

func TestPostCommentParamsValidation_UserIp_ValidHexMaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// 130 chars, all hex
	params.UserIp = ""
	for i := 0; i < 130; i++ {
		params.UserIp += "b"
	}

	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid UserIp with max length 130, got error: %v", err)
	}
}

// Test valid IPv4 address
func TestPostCommentParamsValidation_Valid_AltIPv4(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsAltIPv4)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params for AltIPv4, got error: %v", err)
	}
}

// Test valid IPv6 Address
func TestPostCommentParamsValidation_Valid_IPv6(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv6)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params for IPv6, got error: %v", err)
	}
}

// --- USERID ---

func TestPostCommentParamsValidation_UserID_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_InvalidUUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = "not-a-uuid"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid UserID UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version3UUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	nonV7UUID := uuid.NewMD5(uuid.NameSpaceDNS, []byte("example.com")) // Version 3
	params.UserID = nonV7UUID.String()

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with version 3 UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version4UUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	nonV7UUID, err := uuid.NewRandom() // Version 4
	if err != nil {
		t.Fatal("Failed to generate random UUID for UserID:", err)
	}
	params.UserID = nonV7UUID.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with version 4 UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version5UUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	nonV7UUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("example.com")) // Version 5
	params.UserID = nonV7UUID.String()

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with version 5 UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version6UUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	nonV7UUID, err := uuid.NewV6() // Version 6
	if err != nil {
		t.Fatal("Failed to generate random UUID for UserID:", err)
	}
	params.UserID = nonV7UUID.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with version 6 UUID, got nil")
	}
}

// Helper to create a V7 UUID with a custom timestamp (in seconds since epoch)
//
// Input:
//   - timestamp: a time object representing the time to set the UUID's time segment to
//
// Ouput:
//   - *uuid.UUID: a pointer to a uuid with the specified time segment, nil if error occurred
//   - error: non-nil when an error occurs during processing
func newV7UUIDWithUnixTimestamp(timestamp time.Time) (*uuid.UUID, error) {
	// Create new V7 UUID
	tempUUID, err := uuid.NewV7()
	if err != nil {
		return nil, errors.Join(errors.New("failed to create temporary uuid for time setting"), err)
	}
	//log.Println("temp bytes = ", tempUUID[0:6], ", uuid = ", tempUUID)

	// Create a buffer to read the int64 into for later copying
	shiftBuffer := bytes.NewBuffer([]byte{})
	shiftBuffer.Reset()
	err = binary.Write(shiftBuffer, binary.BigEndian, timestamp.UnixMilli())
	if err != nil {
		return nil, errors.Join(errors.New("failed to write timestamp to temporary buffer"), err)
	}
	//log.Println("Buffer = ", shiftBuffer.Bytes(), ", temp bytes = ", tempUUID[0:6], ", uuid = ", tempUUID)

	//log.Println("temp buffer: ", tempBuffer)

	// Replace the time segement of the uuid (first 6 bytes) with the fabricated time segment
	tempUUID, err = uuid.FromBytes(bytes.Replace(tempUUID[0:16], tempUUID[0:6], shiftBuffer.Bytes()[2:8], 1))
	if err != nil {
		return nil, errors.Join(errors.New("failed to replace timestamp in original uuid"), err)
	}

	//log.Println("Buffer[2:8] = ", shiftBuffer.Bytes()[2:8], ", temp bytes = ", tempUUID[0:6], ", uuid = ", tempUUID)

	return &tempUUID, nil
}

// TestUUIDTimeGen_10Days ensures that the newV7UUIDWithUnixSeconds() function correctly fabricates
// uuids with specified timestamps.
func TestUUIDTimeGen_10Days(t *testing.T) {
	// Generate a timestamp 10 days ago
	timestamp := (time.Now().Add(time.Hour * 24 * -10))

	// Generate a new UUID with that timestamp
	fabricatedUUID, err := newV7UUIDWithUnixTimestamp(timestamp)
	if err != nil {
		t.Fatal("Failed to generate UUID with specified time", err)
	}

	// Write timestamp into a buffer for comparison
	timeBuffer := bytes.NewBuffer([]byte{})
	timeBuffer.Reset()
	err = binary.Write(timeBuffer, binary.BigEndian, timestamp.UnixMilli())
	if err != nil {
		t.Fatal("Failed to parse time to bytes to check against uuid")
	}

	//log.Println(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6])

	// Check the timestampe bytes against the UUID's bytes
	if !bytes.Equal(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6]) {
		t.Fatal("Fabricated UUID timestamp does not match intended timestamp")
	}

	if getUUIDTimestamp(*fabricatedUUID).Unix() != timestamp.Unix() {
		t.Log("Fabricated UUID timestamp: ", getUUIDTimestamp(*fabricatedUUID), getUUIDTimestamp(*fabricatedUUID).Unix(), ", intended timestamp: ", timestamp, timestamp.Unix())
		t.Fatal("Fabricated UUID timestamp does not match intended timestamp (getUUIDTimestamp mismatch)")
	}

	log.Println("Generated 10-day-old UUID: ", fabricatedUUID)
}

// TestUUIDTimeGen_100Days ensures that the newV7UUIDWithUnixSeconds() function correctly fabricates
// uuids with specified timestamps.
func TestUUIDTimeGen_100Days(t *testing.T) {
	// Generate a timestamp 10 days ago
	timestamp := (time.Now().Add(time.Hour * 24 * -100))

	// Generate a new UUID with that timestamp
	fabricatedUUID, err := newV7UUIDWithUnixTimestamp(timestamp)
	if err != nil {
		t.Fatal("Failed to generate UUID with specified time", err)
	}

	// Write timestamp into a buffer for comparison
	timeBuffer := bytes.NewBuffer([]byte{})
	timeBuffer.Reset()
	err = binary.Write(timeBuffer, binary.BigEndian, timestamp.UnixMilli())
	if err != nil {
		t.Fatal("Failed to parse time to bytes to check against uuid")
	}

	//log.Println(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6])

	// Check the timestampe bytes against the UUID's bytes
	if !bytes.Equal(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6]) {
		t.Fatal("Fabricated UUID timestamp does not match intended timestamp")
	}

	if getUUIDTimestamp(*fabricatedUUID).Unix() != timestamp.Unix() {
		t.Log("Fabricated UUID timestamp: ", getUUIDTimestamp(*fabricatedUUID), getUUIDTimestamp(*fabricatedUUID).Unix(), ", intended timestamp: ", timestamp, timestamp.Unix())
		t.Fatal("Fabricated UUID timestamp does not match intended timestamp (getUUIDTimestamp mismatch)")
	}

	log.Println("Generated 100-day-old UUID: ", fabricatedUUID)
}

// Test for illogical UUID dates in the past
func TestPostCommentParamsValidation_UserID_UUIDTooFarInPast(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	// Far past: 1970-01-01
	pastTime := time.Unix(int64(1000), 0)
	uuidPast, err := newV7UUIDWithUnixTimestamp(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (far past):", err)
	}
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidPast.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID too far in the past, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_UUIDSlightlyTooFarPast(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	// Slightly before Tue May 27 2025 23:53:20 GMT+0000 (unix 1748390000)
	pastTime := time.Unix(int64(1748390000-10), 0)
	uuidPast, err := newV7UUIDWithUnixTimestamp(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (slightly past):", err)
	}
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidPast.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID slightly in the past, got nil")
	}
}

// Test for valid UUID date
func TestPostCommentParamsValidation_UserID_UUIDJustAfterValidationStart(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	// Slightly after Tue May 27 2025 23:53:20 GMT+0000 (unix 1748390000)
	pastTime := time.Unix(int64(1748390000+1000), 0)
	uuidPast, err := newV7UUIDWithUnixTimestamp(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (far past):", err)
	}
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidPast.String()

	err = validate.Struct(params)
	if err != nil {
		t.Error("UUID timestamp should be accepted, but was denied")
	}
}

// Test more illogical UUID dates, in the future
func TestPostCommentParamsValidation_UserID_UUIDTooFarInFuture(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	// Far future: 10 years ahead
	futureTime := time.Now().Add(10 * 365 * 24 * time.Hour)
	uuidFuture, err := newV7UUIDWithUnixTimestamp(futureTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (far future):", err)
	}
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidFuture.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID too far in the future, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_UUIDSlightlyInFuture(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	// Slightly in the future: 11 hours ahead
	futureTime := time.Now().Add(11 * time.Hour)
	uuidFuture, err := newV7UUIDWithUnixTimestamp(futureTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for slightly in future:", err)
	}
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidFuture.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID slightly in the future, got nil")
	}
}

// --- USERNAME ---

func TestPostCommentParamsValidation_Username_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.Username = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing Username, got nil")
	}
}

func TestPostCommentParamsValidation_Username_Alphanum(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	// Test non-alphanumeric string
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.Username = "user!@#" // Not alphanum

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-alphanum Username, got nil")
	}

	// Test random alphanumeric string
	params.Username = ""

	// Get random alphas
	for i := 0; i < 15; i++ {
		charNum := rand.Intn(26)
		shifter := 65 + rand.Intn(2)*32 // 65 is ascii character 'A', 97 is ascii character 'a'
		params.Username += string(rune(charNum + shifter))
	}

	// Get random numbers
	for i := 0; i < 10; i++ {
		charNum := rand.Intn(10)
		shifter := 48 // 48 is ascii character '0'
		params.Username += string(rune(charNum + shifter))
	}

	err = validate.Struct(params)
	if err != nil {
		t.Error("Expected no error on alphanumeric username:", params.Username, ", but instead got:", err)
	}
}

func TestPostCommentParamsValidation_Username_OtherChars(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.Username = " .,_-"

	err := validate.Struct(params)
	if err != nil {
		t.Error("Expected no error for username: ", params.Username, ", of miscellaneous characters, but instead got:", err)
	}
}

func TestPostCommentParamsValidation_Username_Characters(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	allowedChars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 .,_-")
	for i := 0; i < 50; i++ {
		params := GetValidPostCommentParams(ValidParamsIPv4)
		randomUsername := make([]rune, 25)
		for j := 0; j < 25; j++ {
			idx := rand.Intn(len(allowedChars))
			randomUsername[j] = allowedChars[idx]
		}
		params.Username = string(randomUsername)

		err := validate.Struct(params)
		if err != nil {
			t.Errorf("[%d] Expected no error for randomized allowed username '%s', but got: %v", i, params.Username, err)
		}
	}
}

// getInvalidUsernameCharacter returns a random ASCII rune not in allowedChars
func getInvalidUsernameCharacter(allowedChars map[rune]struct{}) rune {
	for {
		candidate := rune(rand.Intn(127)) // ASCII 0-126
		if _, ok := allowedChars[candidate]; !ok {
			return candidate
		}
	}
}

func TestPostCommentParamsValidation_Username_InvalidCharacters(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	allowedCharsSlice := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 .,_-")
	allowedChars := make(map[rune]struct{}, len(allowedCharsSlice))
	for _, c := range allowedCharsSlice {
		allowedChars[c] = struct{}{}
	}

	for i := 0; i < 50; i++ {
		params := GetValidPostCommentParams(ValidParamsIPv4)
		randomUsername := make([]rune, 25)
		for j := 0; j < 25; j++ {
			idx := rand.Intn(len(allowedCharsSlice))
			randomUsername[j] = allowedCharsSlice[idx]
		}
		// Insert an invalid character at a random position
		invalidIdx := rand.Intn(25)
		invalidChar := getInvalidUsernameCharacter(allowedChars)
		randomUsername[invalidIdx] = invalidChar
		params.Username = string(randomUsername)

		err := validate.Struct(params)
		if err == nil {
			t.Errorf("[%d] Expected error for username with invalid character '%s' (ASCII %d) at pos %d: '%s', but got nil", i, string(invalidChar), invalidChar, invalidIdx, params.Username)
		}
	}
}

// Test every invalid ASCII character individually in the username
func TestPostCommentParamsValidation_Username_EachInvalidCharacter(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	allowedCharsSlice := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 .,_-")
	allowedChars := make(map[rune]struct{}, len(allowedCharsSlice))
	for _, c := range allowedCharsSlice {
		allowedChars[c] = struct{}{}
	}

	for ascii := 0; ascii < 127; ascii++ {
		candidate := rune(ascii)
		if _, ok := allowedChars[candidate]; ok {
			continue // skip allowed chars
		}
		params := GetValidPostCommentParams(ValidParamsIPv4)
		randomUsername := make([]rune, 25)
		for j := 0; j < 25; j++ {
			idx := rand.Intn(len(allowedCharsSlice))
			randomUsername[j] = allowedCharsSlice[idx]
		}
		// Insert the invalid character at a random position
		invalidIdx := rand.Intn(25)
		randomUsername[invalidIdx] = candidate
		params.Username = string(randomUsername)

		err := validate.Struct(params)
		if err == nil {
			t.Errorf("Expected error for username with invalid character '%s' (ASCII %d) at pos %d: '%s', but got nil", string(candidate), ascii, invalidIdx, params.Username)
		}
	}
}

func TestPostCommentParamsValidation_Username_MinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.Username = "ab" // min=3

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for Username with length < 3, got nil")
	}
}

func TestPostCommentParamsValidation_Username_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.Username = "abcdefghijklmnopqrstuvwxyz" // 26 chars, max=25

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for Username with length > 25, got nil")
	}
}

// --- COMMENTTEXT ---

func TestPostCommentParamsValidation_CommentText_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.CommentText = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentText, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_NonPrintableASCII(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// Insert a non-printable ASCII character (e.g., ASCII 7 - bell)
	params.CommentText = "This is a valid comment.\a"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText containing non-printable ASCII, got nil")
	}
}

// Attempt to validate CommentText with all non-printable ASCII characters, prints
// error indicating which character code caused the failure.
func TestPostCommentParamsValidation_CommentText_AllNonPrintableASCII(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	for i := 0; i < 32; i++ {
		if i == 10 {
			params := GetValidPostCommentParams(ValidParamsIPv4)
			params.CommentText = "Valid text" + string(rune(i))
			err := validate.Struct(params)
			if err != nil {
				t.Error("Expected no error for CommentText containing newline (ASCII code ", i, "), but got:", err)
			}
		} else {
			params := GetValidPostCommentParams(ValidParamsIPv4)
			params.CommentText = "Valid text" + string(rune(i))
			err := validate.Struct(params)
			if err == nil {
				t.Error("Expected error for CommentText containing non-printable ASCII (code ", i, "), got nil")
			}
		}
	}
	// DEL character (ASCII 127)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.CommentText = "Valid text" + string(rune(127))
	err := validate.Struct(params)
	if err == nil {
		t.Errorf("Expected error for CommentText containing non-printable ASCII (code 127), got nil")
	}
}

// Tests validating a comment with only printable ASCII characters.
func TestPostCommentParamsValidation_CommentText_OnlyPrintableASCII(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// All printable ASCII characters from 32 (space) to 126 (~)
	printable := ""
	for i := 32; i <= 126; i++ {
		printable += string(rune(i))
	}
	params.CommentText = printable + string(rune(10))

	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid CommentText with only printable ASCII, got error: %v", err)
	}
}

func TestPostCommentParamsValidation_CommentText_MinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.CommentText = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.CommentText = makeStringOfLength(301) // max=300

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText with length > 300, got nil")
	}
}

// --- LISTINGTITLE ---

func TestPostCommentParamsValidation_ListingTitle_InvalidPGType(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingTitle.String = "Regular title"
	params.ListingTitle.Valid = false

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid Valid field in ListingTitle, got nil")
	}
}

func TestPostCommentParamsValidation_ListingTitle_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingTitle.String = ""
	params.ListingTitle.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing ListingTitle, got nil")
	}
}

func TestPostCommentParamsValidation_ListingTitle_MinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingTitle.String = ""
	params.ListingTitle.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingTitle with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_ListingTitle_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingTitle.String = makeStringOfLength(201) // max=200
	params.ListingTitle.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingTitle with length > 200, got nil")
	}
}

func TestPostCommentParamsValidation_ListingTitle_AllowedCharacters(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingTitle.String = "123 Main St., Apt #5 - New York"
	params.ListingTitle.Valid = true

	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid ListingTitle with allowed characters, got error: %v", err)
	}

	for val := 32; val < 127; val++ {
		params := GetValidPostCommentParams(ValidParamsIPv4)
		params.ListingTitle.String = string(rune(val))
		params.ListingTitle.Valid = true

		err := validate.Struct(params)
		if err != nil {
			t.Error("Expected no error for ListingTitle with allowed character #", val, " (", string(rune(val)), "), got", err)
		}
	}
}

func TestPostCommentParamsValidation_ListingTitle_DisallowedCharacters(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	for val := 0; val < 32; val++ {
		params := GetValidPostCommentParams(ValidParamsIPv4)
		params.ListingTitle.String = string(rune(val))
		params.ListingTitle.Valid = true

		err := validate.Struct(params)
		if err == nil {
			t.Error("Expected error for ListingTitle with disallowed character #", val, " (", string(rune(val)), "), got nil")
		}
	}

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingTitle.String = string(rune(127))
	params.ListingTitle.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingTitle with disallowed character #", 127, " (", string(rune(127)), "), got nil")
	}
}

// --- IP NONCE ---

func TestPostCommentParamsValidation_IpNonce_InvalidPGType(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.IpNonce.String = makeStringOfLength(24)
	params.IpNonce.Valid = false

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid Valid field in IpNonce, got nil")
	}
}

func TestPostCommentParamsValidation_IpNonce_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.IpNonce.String = ""
	params.IpNonce.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing IpNonce, got nil")
	}
}

func TestPostCommentParamsValidation_IpNonce_NotHexadecimal(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// Insert a non-hex character
	params.IpNonce.String = makeStringOfLength(23) + "Z"
	params.IpNonce.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for IpNonce with non-hexadecimal character, got nil")
	}
}

func TestPostCommentParamsValidation_IpNonce_WrongLength_Short(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// 23 chars, should be 24
	params.IpNonce.String = makeStringOfLength(23)
	params.IpNonce.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for IpNonce with length < 24, got nil")
	}
}

func TestPostCommentParamsValidation_IpNonce_WrongLength_Long(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// 25 chars, should be 24
	params.IpNonce.String = makeStringOfLength(25)
	params.IpNonce.Valid = true

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for IpNonce with length > 24, got nil")
	}
}

func TestPostCommentParamsValidation_IpNonce_ValidHexLength24(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := GetValidPostCommentParams(ValidParamsIPv4)
	// 24 chars, all hex
	params.IpNonce.String = ""
	for i := 0; i < 24; i++ {
		params.IpNonce.String += "a"
	}
	params.IpNonce.Valid = true

	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid IpNonce with length 24 and hex, got error: %v", err)
	}
}

// Helper to create a string of a given length
func makeStringOfLength(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}

/* func TestAscii(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	//validate.Var("Input", "ascii")

	t.Log("\n\nStarting validation test for printascii:\n")

	err := validate.Var("54 Bethany Drive, Commack, NY 11725", "printascii")
	if err != nil {
		t.Error("Failed to validate ")
	}

	for i := 0; i <= 127; i++ {
		err := validate.Var(string(rune(i)), "printascii")
		if err != nil {
			t.Log("  -  Error validating (ascii", i, ")")
		}
	}
} */

// ===================================================================================================================== //
//                                        Custom UUID Validator Tests                                                    //
// ===================================================================================================================== //

func TestCustomUUIDValidator_ValidV7UUID(t *testing.T) {
	u, err := uuid.NewV7()
	if err != nil {
		t.Fatal("Failed to generate V7 UUID:", err)
	}
	if err := customUUIDValidator(u); err != nil {
		t.Error("Expected valid V7 UUID, got error:", err)
	}
}

func TestCustomUUIDValidator_InvalidVersion(t *testing.T) {
	uuidV3 := uuid.NewMD5(uuid.NameSpaceDNS, []byte("example.com")) // Version 3
	if err := customUUIDValidator(uuidV3); err == nil {
		t.Error("Expected error for non-V7 UUID version, got nil")
	}
	uuidV4, err := uuid.NewRandom() // Version 4
	if err != nil {
		t.Fatal("Failed to generate V4 UUID:", err)
	}
	if err := customUUIDValidator(uuidV4); err == nil {
		t.Error("Expected error for V4 UUID version, got nil")
	}
	uuidV6, err := uuid.NewV6() // Version 6
	if err != nil {
		t.Fatal("Failed to generate V6 UUID:", err)
	}
	if err := customUUIDValidator(uuidV6); err == nil {
		t.Error("Expected error for V6 UUID version, got nil")
	}
}

func TestCustomUUIDValidator_TooFarInPast(t *testing.T) {
	// Far past: 1970-01-01
	pastTime := time.Unix(1000, 0)
	u, err := newV7UUIDWithUnixTimestamp(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for far past:", err)
	}
	if err := customUUIDValidator(*u); err == nil {
		t.Error("Expected error for UUID too far in the past, got nil")
	}
}

func TestCustomUUIDValidator_JustBeforeAllowedPast(t *testing.T) {
	// Just before allowed past: slightly before May 27 2025 23:53:20 GMT+0000
	pastTime := time.Unix(1748390000-10, 0)
	u, err := newV7UUIDWithUnixTimestamp(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for just before allowed past:", err)
	}
	if err := customUUIDValidator(*u); err == nil {
		t.Error("Expected error for UUID just before allowed past, got nil")
	}
}

func TestCustomUUIDValidator_JustAfterAllowedPast(t *testing.T) {
	// Just after allowed past: slightly after May 27 2025 23:53:20 GMT+0000
	pastTime := time.Unix(1748390000+1000, 0)
	u, err := newV7UUIDWithUnixTimestamp(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for just after allowed past:", err)
	}
	if err := customUUIDValidator(*u); err != nil {
		t.Errorf("Expected valid UUID just after allowed past, got error: %v", err)
	}
}

func TestCustomUUIDValidator_TooFarInFuture(t *testing.T) {
	// Far future: 10 years ahead
	futureTime := time.Now().Add(10 * 365 * 24 * time.Hour)
	u, err := newV7UUIDWithUnixTimestamp(futureTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for far future:", err)
	}
	if err := customUUIDValidator(*u); err == nil {
		t.Error("Expected error for UUID too far in the future, got nil")
	}
}

func TestCustomUUIDValidator_SlightlyInFuture(t *testing.T) {
	// Slightly in the future: 11 hours ahead
	futureTime := time.Now().Add(11 * time.Hour)
	u, err := newV7UUIDWithUnixTimestamp(futureTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for slightly in future:", err)
	}
	if err := customUUIDValidator(*u); err == nil {
		t.Error("Expected error for UUID slightly in the future, got nil")
	}
}

// ===================================================================================================================== //
//                                         Custom IP Validator Tests                                                     //
// ===================================================================================================================== //

func TestValidateIP_ValidIPv4(t *testing.T) {
	_, validate := ValidationSetupAndTeardown(t)
	ip := "192.168.1.1"
	err := ValidateIP(validate, ip)
	if err != nil {
		t.Errorf("Expected valid IPv4, got error: %v", err)
	}
}

func TestValidateIP_ValidIPv6(t *testing.T) {
	_, validate := ValidationSetupAndTeardown(t)
	ip := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	err := ValidateIP(validate, ip)
	if err != nil {
		t.Errorf("Expected valid IPv6, got error: %v", err)
	}
}

func TestValidateIP_InvalidIP(t *testing.T) {
	_, validate := ValidationSetupAndTeardown(t)
	ip := "not_an_ip"
	err := ValidateIP(validate, ip)
	if err == nil {
		t.Error("Expected error for invalid IP, got nil")
	}
}

func TestValidateIP_EmptyString(t *testing.T) {
	_, validate := ValidationSetupAndTeardown(t)
	ip := ""
	err := ValidateIP(validate, ip)
	if err == nil {
		t.Error("Expected error for empty IP, got nil")
	}
}

func TestValidateIP_IPv4WithPort(t *testing.T) {
	_, validate := ValidationSetupAndTeardown(t)
	ip := "192.168.1.1:8080"
	err := ValidateIP(validate, ip)
	if err == nil {
		t.Error("Expected error for IPv4 with port, got nil")
	}
}

func TestValidateIP_IPv6WithPort(t *testing.T) {
	_, validate := ValidationSetupAndTeardown(t)
	ip := "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:443"
	err := ValidateIP(validate, ip)
	if err == nil {
		t.Error("Expected error for IPv6 with port, got nil")
	}
}
