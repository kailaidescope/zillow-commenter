package sqlc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ===================================================================================================================== //
//                                             Setup and Teardown                                                        //
// ===================================================================================================================== //

// SetupAndTeardown initializes the validator and registers the PostCommentParamsValidation function.
//
// IMPORTANT: This function should be called in each test case to ensure the validator is set up correctly.
//
// Input:
//   - tb: A testing.TB interface that allows the function to log messages and handle test failures.
//
// Output:
//   - A function that can be deferred to perform teardown actions after the test completes.
//   - A pointer to a validator.Validate instance that can be used to validate structs.
func SetupAndTeardown(tb testing.TB) (func(tb testing.TB), *validator.Validate) {
	// Create a validator singleton
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Register the PostCommentParamsValidation function for validating PostCommentParams structs
	validate.RegisterStructValidation(PostCommentParamsValidation, PostCommentParams{})

	return func(tb testing.TB) {
		tb.Log("Teardown complete")
	}, validate
}

// ===================================================================================================================== //
//                                                Write tests below                                                      //
// ===================================================================================================================== //

// Helper to create a valid PostCommentParams
type ValidPostCommentParamsType int

const (
	ValidParamsIPv4 ValidPostCommentParamsType = iota
	ValidParamsIPv6
	ValidParamsAltIPv4
)

func validPostCommentParams(paramType ValidPostCommentParamsType) PostCommentParams {
	// Create a valid CommentID
	commentID, err := validPgtypeUUID()
	if err != nil {
		log.Fatal("Failed to create valid CommentID", err)
	}

	// Create a valid userID
	userID, err := uuid.NewV7()
	if err != nil {
		log.Fatal("Failed to create valid UUID for UserID", err)
	}

	switch paramType {
	case ValidParamsIPv6:
		return PostCommentParams{
			CommentID:   *commentID,
			ListingID:   "654321",
			UserIp:      "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			UserID:      userID.String(),
			Username:    "TestUserIPv6",
			CommentText: "This is a valid IPv6 comment.",
		}
	case ValidParamsAltIPv4:
		return PostCommentParams{
			CommentID:   *commentID,
			ListingID:   "789012",
			UserIp:      "10.0.0.1",
			UserID:      userID.String(),
			Username:    "TestUserAltIPv4",
			CommentText: "This is another valid IPv4 comment.",
		}
	default: // ValidParamsIPv4
		return PostCommentParams{
			CommentID:   *commentID,
			ListingID:   "123456",
			UserIp:      "192.168.1.1",
			UserID:      userID.String(),
			Username:    "TestUser",
			CommentText: "This is a valid comment.",
		}
	}
}

// Helper to create a valid pgtype.UUID (replace with your actual type if needed)
func validPgtypeUUID() (*pgtype.UUID, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return nil, errors.Join(errors.New("failed to generate UUID"), err)
	}

	return &pgtype.UUID{Bytes: [16]byte(newUUID), Valid: true}, nil
}

// ===================================================================================================================== //
//                                                Test Cases                                                             //
// ===================================================================================================================== //

func TestPostCommentParamsValidation_Valid(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params, got error: %v", err)
	}
}

func TestPostCommentParamsValidation_Valid_AltIPv4(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsAltIPv4)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params for AltIPv4, got error: %v", err)
	}
}

func TestPostCommentParamsValidation_Valid_IPv6(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv6)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params for IPv6, got error: %v", err)
	}
}

// --- COMMENTID ---

func TestPostCommentParamsValidation_CommentID_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentID = pgtype.UUID{} // Zero value, not valid

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentID, got nil")
	}
}

func TestPostCommentParamsValidation_CommentID_AlmostValidUUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	// Set invalid UUID bytes (not a valid UUID)
	tempId := params.CommentID.String()
	// Change the 15th character (index 14) from '7' to a different digit, changing the version code
	if len(tempId) > 14 && tempId[14] == '7' {
		tempId = tempId[:14] + "3" + tempId[15:] // Change '7' to '3' for testing
	}
	uuid, err := uuid.Parse(tempId)
	if err != nil {
		t.Error("Failed to parse modified CommentID UUID:", err)
		return
	}
	params.CommentID = pgtype.UUID{Bytes: [16]byte(uuid), Valid: true}

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid CommentID UUID, got nil")
	}
}

func TestPostCommentParamsValidation_CommentID_InvalidUUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	// Set invalid UUID bytes (not a valid UUID)
	params.CommentID = pgtype.UUID{Bytes: [16]byte{0, 0, 3}, Valid: true}

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid CommentID UUID, got nil")
	}
}

// --- LISTINGID ---

func TestPostCommentParamsValidation_ListingID_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_Number(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "abc123" // Not a number

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-numeric ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_ExcludesDot(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "123.456" // Contains a dot

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID containing '.', got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MinLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MaxLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "123456789012345678901" // 21 chars, max=20

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID with length > 20, got nil")
	}
}

// --- USERIP ---

func TestPostCommentParamsValidation_UserIp_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserIp = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserIp, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_InvalidIP(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserIp = "not_an_ip"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid UserIp, got nil")
	}
}

// --- USERID ---

func TestPostCommentParamsValidation_UserID_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_InvalidUUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = "not-a-uuid"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid UserID UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version3UUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	nonV7UUID := uuid.NewMD5(uuid.NameSpaceDNS, []byte("example.com")) // Version 3
	params.UserID = nonV7UUID.String()

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with version 3 UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version4UUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
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
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	nonV7UUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("example.com")) // Version 5
	params.UserID = nonV7UUID.String()

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with version 5 UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version6UUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
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
func newV7UUIDWithUnixSeconds(timestamp time.Time) (*uuid.UUID, error) {
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
	fabricatedUUID, err := newV7UUIDWithUnixSeconds(timestamp)
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
	if bytes.Compare(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6]) != 0 {
		t.Fatal("Fabricated UUID timestamp does not match intended timestamp")
	}

	log.Println("Generated 10-day-old UUID: ", fabricatedUUID)
}

// TestUUIDTimeGen_100Days ensures that the newV7UUIDWithUnixSeconds() function correctly fabricates
// uuids with specified timestamps.
func TestUUIDTimeGen_100Days(t *testing.T) {
	// Generate a timestamp 10 days ago
	timestamp := (time.Now().Add(time.Hour * 24 * -100))

	// Generate a new UUID with that timestamp
	fabricatedUUID, err := newV7UUIDWithUnixSeconds(timestamp)
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
	if bytes.Compare(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6]) != 0 {
		t.Fatal("Fabricated UUID timestamp does not match intended timestamp")
	}

	log.Println("Generated 100-day-old UUID: ", fabricatedUUID)
}

// Tests for illogical UUID dates

func TestPostCommentParamsValidation_UserID_UUIDTooFarInPast(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	// Far past: 1970-01-01
	pastTime := time.Unix(int64(1000), 0)
	uuidPast, err := newV7UUIDWithUnixSeconds(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (far past):", err)
	}
	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidPast.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID too far in the past, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_UUIDSlightlyInPast(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	// Slightly before May 27, 2025 (reference: 1748389238)
	pastTime := time.Unix(int64(1748389238-10000), 0)
	uuidPast, err := newV7UUIDWithUnixSeconds(pastTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (slightly past):", err)
	}
	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidPast.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID slightly in the past, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_UUIDTooFarInFuture(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	// Far future: 10 years ahead
	futureTime := time.Now().Add(10 * 365 * 24 * time.Hour)
	uuidFuture, err := newV7UUIDWithUnixSeconds(futureTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (far future):", err)
	}
	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidFuture.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID too far in the future, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_UUIDSlightlyInFuture(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	// Slightly in the future: 101 hours ahead
	futureTime := time.Now().Add(101 * time.Hour)
	uuidFuture, err := newV7UUIDWithUnixSeconds(futureTime)
	if err != nil {
		t.Fatal("Failed to generate V7 UUID for UserID (slightly future):", err)
	}
	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = uuidFuture.String()

	err = validate.Struct(params)
	if err == nil {
		t.Error("Expected error for UserID with V7 UUID slightly in the future, got nil")
	}
}

// --- USERNAME ---

func TestPostCommentParamsValidation_Username_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing Username, got nil")
	}
}

func TestPostCommentParamsValidation_Username_Alphanum(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = "user!@#" // Not alphanum

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-alphanum Username, got nil")
	}
}

func TestPostCommentParamsValidation_Username_MinLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = "ab" // min=3

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for Username with length < 3, got nil")
	}
}

func TestPostCommentParamsValidation_Username_MaxLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = "abcdefghijklmnopqrstuvwxyz" // 26 chars, max=25

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for Username with length > 25, got nil")
	}
}

// --- COMMENTTEXT ---

func TestPostCommentParamsValidation_CommentText_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentText = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentText, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_NonPrintableASCII(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	// Insert a non-printable ASCII character (e.g., ASCII 7 - bell)
	params.CommentText = "This is a valid comment.\a"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText containing non-printable ASCII, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_OnlyPrintableASCII(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	// All printable ASCII characters from 32 (space) to 126 (~)
	printable := ""
	for i := 32; i <= 126; i++ {
		printable += string(rune(i))
	}
	params.CommentText = printable

	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid CommentText with only printable ASCII, got error: %v", err)
	}
}

func TestPostCommentParamsValidation_CommentText_MinLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentText = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_MaxLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentText = makeStringOfLength(301) // max=300

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText with length > 300, got nil")
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
