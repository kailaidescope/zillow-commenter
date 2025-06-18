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
	"github.com/microcosm-cc/bluemonday"
)

// ===================================================================================================================== //
//                                             Setup and Teardown                                                        //
// ===================================================================================================================== //

// SanitizationSetupAndTeardown initializes the sanitizer.
//
// IMPORTANT: This function should be called in each test case to ensure the sanitizer is set up correctly.
//
// Input:
//   - tb: A testing.TB interface that allows the function to log messages and handle test failures.
//
// Output:
//   - A function that can be deferred to perform teardown actions after the test completes.
//   - A pointer to a bluemonday.Policy instance that can be used to sanitize strings.
func SanitizationSetupAndTeardown(tb testing.TB) (func(tb testing.TB), *bluemonday.Policy) {
	// Create a sanitization policy
	sanitizationPolicy := bluemonday.StrictPolicy()

	return func(tb testing.TB) {
		tb.Log("Teardown complete")
	}, sanitizationPolicy
}

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

	// Register the PostCommentParamsValidation function for validating PostCommentParams structs
	validate.RegisterStructValidation(PostCommentParamsValidation, PostCommentParams{})

	return func(tb testing.TB) {
		tb.Log("Teardown complete")
	}, validate
}

// ===================================================================================================================== //
//                                                Write tests below                                                      //
// ===================================================================================================================== //

//

// ===================================================================================================================== //
//                                               Sanitization Tests                                                      //
// ===================================================================================================================== //

func TestSanitize_ListingID(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "<b>123456</b>"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.ListingID != "123456" {
		t.Errorf("Expected sanitized ListingID to be '123456', got '%s'", sanitized.ListingID)
	}
}

func TestSanitize_UserIp(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.UserIp = "<script>alert('x')</script>192.168.1.1"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserIp != "192.168.1.1" {
		t.Errorf("Expected sanitized UserIp to be '192.168.1.1', got '%s'", sanitized.UserIp)
	}
}

func TestSanitize_UserID(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = "<i>" + params.UserID + "</i>"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserID != params.UserID[3:len(params.UserID)-4] {
		t.Errorf("Expected sanitized UserID to be '%s', got '%s'", params.UserID[3:len(params.UserID)-4], sanitized.UserID)
	}
}

func TestSanitize_Username(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = "<b>TestUser</b>"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.Username != "TestUser" {
		t.Errorf("Expected sanitized Username to be 'TestUser', got '%s'", sanitized.Username)
	}
}

func TestSanitize_CommentText(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentText = "<script>alert('xss')</script>This is a comment."
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.CommentText != "This is a comment." {
		t.Errorf("Expected sanitized CommentText to be 'This is a comment.', got '%s'", sanitized.CommentText)
	}
}

func TestSanitize_ListingID_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = `<img src="x" onerror="alert('XSS')">123456<script>alert(1)</script>`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.ListingID != "123456" {
		t.Errorf("Expected sanitized ListingID to be '123456', got '%s'", sanitized.ListingID)
	}
}

func TestSanitize_UserIp_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.UserIp = `192.168.1.1"><svg/onload=alert(2)>`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserIp != `192.168.1.1&#34;&gt;` {
		t.Errorf("Expected sanitized UserIp to be '192.168.1.1\"', got '%s'", sanitized.UserIp)
	}
}

func TestSanitize_UserID_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	expectedUserID := params.UserID // Store the original UserID
	params.UserID = `<iframe src="javascript:alert('XSS')"></iframe>` + params.UserID + `<script>alert(3)</script>`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserID != expectedUserID {
		t.Errorf("Expected sanitized UserID to be '%s', got '%s'", params.UserID, sanitized.UserID)
	}
}

func TestSanitize_Username_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = `"><img src=x onerror=alert(4)>TestUser<script>alert(5)</script>`
	sanitized := params.Sanitize(*sanitizer)
	// Should return the username with the characters encoded as HTML character entities
	if sanitized.Username != `&#34;&gt;TestUser` {
		t.Errorf("Expected sanitized Username to be '\">TestUser', got '%s'", sanitized.Username)
	}
}

func TestSanitize_CommentText_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentText = `<script>alert('xss')</script>This is a comment.<img src="x" onerror="alert('XSS')">`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.CommentText != "This is a comment." {
		t.Errorf("Expected sanitized CommentText to be 'This is a comment.', got '%s'", sanitized.CommentText)
	}
}

// ===================================================================================================================== //
//                                         Unit Tests for String Sanitizers                                             //
// ===================================================================================================================== //

func TestRemoveLinks(t *testing.T) {
	replacementText := "[link removed]"

	cases := []struct {
		input    string
		expected string
	}{
		{"Check this out: http://example.com", "Check this out: " + replacementText},
		{"Visit https://secure.com for info", "Visit " + replacementText + " for info"},
		{"Go to www.website.org now!", "Go to " + replacementText + " now!"},
		{"No links here", "No links here"},
		{"Multiple links: http://a.com and https://b.com", "Multiple links: " + replacementText + " and " + replacementText},
		{"Text before http://foo.com and after", "Text before " + replacementText + " and after"},
		{"https://abc.com?query=1", replacementText},
		{"www.abc.com/page.html", replacementText},
		{"Mixed: www.abc.com, http://def.com, and text", "Mixed: " + replacementText + ", " + replacementText + ", and text"},
		{"ftp://notalink.com", "ftp://notalink.com"}, // Should not match
		{"http://", replacementText},
		{"www.", "www."},
		{"https://sub.domain.com/path", replacementText},
		{"Check www.site.com and http://site.com", "Check " + replacementText + " and " + replacementText},
		{"Just text", "Just text"},
		{"http://example.com/path?query=1#fragment", replacementText},
		{"www.example.com:8080", replacementText},
		{"http://example.com.", replacementText + "."},
	}

	for _, c := range cases {
		result := removeLinks(c.input)
		if result != c.expected {
			t.Error("removeLinks failed:", "input='"+c.input+"'", "expected='"+c.expected+"'", "got='"+result+"'")
		} else {
			//t.Logf("removeLinks passed: input='%s', expected='%s', got='%s'", c.input, c.expected, result)
		}
	}
}

func TestRemoveEmails(t *testing.T) {
	replacementText := "[email removed]"

	cases := []struct {
		input    string
		expected string
	}{
		{"Contact me at test@example.com", "Contact me at " + replacementText},
		{"Emails: foo@bar.com, bar@foo.org", "Emails: " + replacementText + ", " + replacementText},
		{"No email here", "No email here"},
		{"Edge case: a@b.c", "Edge case: a@b.c"}, // Should not match, as TLD is only 1 char
		{"Send to john.doe@company.co.uk", "Send to " + replacementText},
		{"Multiple: a@b.com b@c.net c@d.org", "Multiple: " + replacementText + " " + replacementText + " " + replacementText},
		{"test@sub.domain.com", replacementText},
		{"user+tag@domain.com", replacementText},
		{"user_name@domain.io", replacementText},
		{"user@domain", "user@domain"},     // Invalid, should not match
		{"user@domain.c", "user@domain.c"}, // TLD too short
		{"user@domain.comm", replacementText},
		{"user@domain.com.", replacementText + "."},
		{"user@domain.com!", replacementText + "!"},
		{"user@domain.com?subject=hi", replacementText + "?subject=hi"},
		{"user@domain.com;user2@domain.com", replacementText + ";" + replacementText},
	}

	for _, c := range cases {
		result := removeEmails(c.input)
		if result != c.expected {
			t.Error("removeEmails failed:", "input='"+c.input+"'", "expected='"+c.expected+"'", "got='"+result+"'")
		} else {
			//t.Logf("removeEmails passed: input='%s', expected='%s', got='%s'", c.input, c.expected, result)
		}
	}
}

func TestRemovePhoneNumbers(t *testing.T) {
	replacementText := "[phone number removed]"

	cases := []struct {
		input    string
		expected string
	}{
		{"Call me at 555-123-4567", "Call me at " + replacementText},
		{"My number is (555) 123-4567.", "My number is " + replacementText + "."},
		{"+1 555 123 4567 is my office.", replacementText + " is my office."},
		{"No phone here", "No phone here"},
		{"Multiple: 555.123.4567 and 5551234567", "Multiple: " + replacementText + " and " + replacementText},
		{"5551234567", replacementText},
		{"(555)123-4567", replacementText},
		{"555 123 4567", replacementText},
		{"555.123.4567", replacementText},
		{"+44 20 7946 0958", replacementText},
		{"123-4567", "123-4567"}, // Not a full phone number, should not match
		{"555-1234", "555-1234"}, // Not a full phone number, should not match
		{"Phone: 555-123-4567, Alt: (555) 123-4567", "Phone: " + replacementText + ", Alt: " + replacementText},
		{"5551234567 ext. 89", replacementText + " ext. 89"},
		{"Text 555-123-4567 text", "Text " + replacementText + " text"},
		{"(555)1234567", replacementText},
		{"555123-4567", replacementText},
	}

	for _, c := range cases {
		result := removePhoneNumbers(c.input)
		if result != c.expected {
			t.Error("removePhoneNumbers failed:", "input='"+c.input+"'", "expected='"+c.expected+"'", "got='"+result+"'")
		} else {
			//t.Logf("removePhoneNumbers passed: input='%s', expected='%s', got='%s'", c.input, c.expected, result)
		}
	}
}

// ===================================================================================================================== //
//                                             Validation Test Helpers                                                   //
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
//                                             Validation Tests                                                          //
// ===================================================================================================================== //

func TestPostCommentParamsValidation_Valid(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params, got error: %v", err)
	}
}

// --- COMMENTID ---

func TestPostCommentParamsValidation_CommentID_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentID = pgtype.UUID{} // Zero value, not valid

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentID, got nil")
	}
}

func TestPostCommentParamsValidation_CommentID_AlmostValidUUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
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
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_Number(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "abc123" // Not a number

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-numeric ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_ExcludesDot(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "123.456" // Contains a dot

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID containing '.', got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.ListingID = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
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
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserIp = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserIp, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_InvalidIP(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserIp = "not_an_ip"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid UserIp, got nil")
	}
}

// Test valid IPv4 address
func TestPostCommentParamsValidation_Valid_AltIPv4(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsAltIPv4)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params for AltIPv4, got error: %v", err)
	}
}

// Test valid IPv6 Address
func TestPostCommentParamsValidation_Valid_IPv6(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv6)
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params for IPv6, got error: %v", err)
	}
}

// --- USERID ---

func TestPostCommentParamsValidation_UserID_Required(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_InvalidUUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.UserID = "not-a-uuid"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid UserID UUID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_Version3UUID(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
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
	teardown, validate := ValidationSetupAndTeardown(t)
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
	teardown, validate := ValidationSetupAndTeardown(t)
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
	teardown, validate := ValidationSetupAndTeardown(t)
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
	if bytes.Compare(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6]) != 0 {
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
	if bytes.Compare(timeBuffer.Bytes()[2:8], fabricatedUUID[0:6]) != 0 {
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
	params := validPostCommentParams(ValidParamsIPv4)
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
	params := validPostCommentParams(ValidParamsIPv4)
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
	params := validPostCommentParams(ValidParamsIPv4)
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
	params := validPostCommentParams(ValidParamsIPv4)
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
	params := validPostCommentParams(ValidParamsIPv4)
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

	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing Username, got nil")
	}
}

func TestPostCommentParamsValidation_Username_Alphanum(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = "user!@#" // Not alphanum

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-alphanum Username, got nil")
	}
}

func TestPostCommentParamsValidation_Username_MinLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.Username = "ab" // min=3

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for Username with length < 3, got nil")
	}
}

func TestPostCommentParamsValidation_Username_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
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
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentText = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentText, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_NonPrintableASCII(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
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
		params := validPostCommentParams(ValidParamsIPv4)
		params.CommentText = "Valid text" + string(rune(i))
		err := validate.Struct(params)
		if err == nil {
			t.Error("Expected error for CommentText containing non-printable ASCII (code ", i, "), got nil")
		}
	}
	// DEL character (ASCII 127)
	params := validPostCommentParams(ValidParamsIPv4)
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
	teardown, validate := ValidationSetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams(ValidParamsIPv4)
	params.CommentText = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_MaxLength(t *testing.T) {
	teardown, validate := ValidationSetupAndTeardown(t)
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
