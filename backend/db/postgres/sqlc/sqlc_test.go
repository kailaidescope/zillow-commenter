package sqlc

import (
	"errors"
	"log"
	"testing"

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
func validPostCommentParams() PostCommentParams {
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

	return PostCommentParams{
		CommentID:   *commentID,
		ListingID:   "123456",
		UserIp:      "192.168.1.1",
		UserID:      userID.String(),
		Username:    "TestUser",
		CommentText: "This is a valid comment.",
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

	params := validPostCommentParams()
	err := validate.Struct(params)
	if err != nil {
		t.Errorf("Expected valid params, got error: %v", err)
	}
}

// --- COMMENTID ---

func TestPostCommentParamsValidation_CommentID_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.CommentID = pgtype.UUID{} // Zero value, not valid

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentID, got nil")
	}
}

// FAILED
func TestPostCommentParamsValidation_CommentID_InvalidUUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
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

	params := validPostCommentParams()
	params.ListingID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_Number(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.ListingID = "abc123" // Not a number

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-numeric ListingID, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_ExcludesDot(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.ListingID = "123.456" // Contains a dot

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID containing '.', got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MinLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.ListingID = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for ListingID with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_ListingID_MaxLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
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

	params := validPostCommentParams()
	params.UserIp = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserIp, got nil")
	}
}

func TestPostCommentParamsValidation_UserIp_InvalidIP(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
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

	params := validPostCommentParams()
	params.UserID = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing UserID, got nil")
	}
}

func TestPostCommentParamsValidation_UserID_InvalidUUID(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.UserID = "not-a-uuid"

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for invalid UserID UUID, got nil")
	}
}

// --- USERNAME ---

func TestPostCommentParamsValidation_Username_Required(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.Username = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing Username, got nil")
	}
}

func TestPostCommentParamsValidation_Username_Alphanum(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.Username = "user!@#" // Not alphanum

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for non-alphanum Username, got nil")
	}
}

func TestPostCommentParamsValidation_Username_MinLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.Username = "ab" // min=3

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for Username with length < 3, got nil")
	}
}

func TestPostCommentParamsValidation_Username_MaxLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
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

	params := validPostCommentParams()
	params.CommentText = ""

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for missing CommentText, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_MinLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
	params.CommentText = "" // min=1

	err := validate.Struct(params)
	if err == nil {
		t.Error("Expected error for CommentText with length < 1, got nil")
	}
}

func TestPostCommentParamsValidation_CommentText_MaxLength(t *testing.T) {
	teardown, validate := SetupAndTeardown(t)
	defer teardown(t)

	params := validPostCommentParams()
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
