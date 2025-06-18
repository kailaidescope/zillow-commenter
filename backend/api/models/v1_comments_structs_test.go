package models

import (
	"errors"
	"log"
	"math/big"
	"testing"
	"time"

	"zillow-commenter.com/m/db/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ===================================================================================================================== //
//                                                   Model Tests                                                         //
// ===================================================================================================================== //

// --- GenericRowToComment tests ---

// Stub struct to simulate a generic database row.
type fakeRow struct {
	CommentID   pgtype.UUID
	ListingID   string
	UserIp      string
	UserID      string
	Username    string
	CommentText string
	Extract     pgtype.Numeric
}

// validPgtypeUUID generates a valid pgtype.UUID for testing.
func validPgtypeUUID() (pgtype.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return pgtype.UUID{}, errors.Join(errors.New("Failed to generate valid UUID"), err)
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}, nil
}

// validPgtypeNumeric creates a valid pgtype.Numeric for testing.
func validPgtypeNumeric(val int64) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(val),
		Valid: true,
	}
}

// Arbitrary row structure that should be convertible to a Comment.
func defaultFakeRow() fakeRow {
	commentID, _ := validPgtypeUUID()
	userID, _ := uuid.NewV7()
	timestamp := time.Now()
	return fakeRow{
		CommentID:   commentID,
		ListingID:   "1234567",
		UserIp:      "127.0.0.1",
		UserID:      userID.String(),
		Username:    "tester",
		CommentText: "hello",
		Extract:     validPgtypeNumeric(timestamp.Unix()),
	}
}

// Test for arbitrary row conversion to Comment.
func TestGenericRowToComment_ValidFakeRow(t *testing.T) {
	row := defaultFakeRow()
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	if comment.TargetListing != row.ListingID || comment.Username != row.Username {
		t.Error("Unexpected comment fields: ", comment)
	}
	log.Println("Successfully converted fake SQLC row struct to Comment:\n\n", comment, "\n\nfrom row:\n\n", row)
}

// Default PostCommentRow for testing.
func defaultPostCommentRow() sqlc.PostCommentRow {
	commentID, _ := validPgtypeUUID()
	userID, _ := uuid.NewV7()
	timestamp := time.Now()
	return sqlc.PostCommentRow{
		CommentID:   commentID,
		ListingID:   "1234567",
		UserIp:      "127.0.0.1",
		UserID:      userID.String(),
		Username:    "tester",
		CommentText: "hello",
		Extract:     validPgtypeNumeric(timestamp.Unix()),
	}
}

// Test for converting PostCommentRow to Comment.
func TestGenericRowToComment_ValidPostCommentRow(t *testing.T) {
	row := defaultPostCommentRow()
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	if comment.TargetListing != row.ListingID || comment.Username != row.Username {
		t.Error("Unexpected comment fields: ", comment)
	}
	log.Println("Successfully converted PostCommentRow to Comment:\n\n", comment, "\n\nfrom row:\n\n", row)
}

// Default GetCommentsByListingIDRow for testing.
func defaultGetCommentRow() sqlc.GetCommentsByListingIDRow {
	commentID, _ := validPgtypeUUID()
	userID, _ := uuid.NewV7()
	timestamp := time.Now()
	return sqlc.GetCommentsByListingIDRow{
		CommentID:   commentID,
		ListingID:   "1234567",
		UserIp:      "127.0.0.1",
		UserID:      userID.String(),
		Username:    "tester",
		CommentText: "hello",
		Extract:     validPgtypeNumeric(timestamp.Unix()),
	}
}

// Test for converting GetCommentsByListingIDRow to Comment.
func TestGenericRowToComment_ValidGetCommentRow(t *testing.T) {
	row := defaultGetCommentRow()
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	if comment.TargetListing != row.ListingID || comment.Username != row.Username {
		t.Error("Unexpected comment fields: ", comment)
	}
	log.Println("Successfully converted GetCommentRow to Comment:\n\n", comment, "\n\nfrom row:\n\n", row)
}

func TestGenericRowToComment_InvalidType(t *testing.T) {
	_, err := GenericSQLCRowToComment(123)
	if err == nil {
		t.Error("Expected error for non-struct input")
	}
}

func TestGenericRowToComment_MissingField(t *testing.T) {
	type Incomplete struct {
		ListingID string
	}
	_, err := GenericSQLCRowToComment(Incomplete{ListingID: "foo"})
	if err == nil {
		t.Error("Expected missing CommentID field error")
	}
}

func TestGenericRowToComment_InvalidUUIDType(t *testing.T) {
	row := defaultFakeRow()
	type BadUUID struct {
		CommentID   string
		ListingID   string
		UserIp      string
		UserID      string
		Username    string
		CommentText string
		Extract     pgtype.Numeric
	}
	badRow := BadUUID{
		CommentID:   "not-a-uuid",
		ListingID:   row.ListingID,
		UserIp:      row.UserIp,
		UserID:      row.UserID,
		Username:    row.Username,
		CommentText: row.CommentText,
		Extract:     row.Extract,
	}
	convertedRow, err := GenericSQLCRowToComment(badRow)
	if err == nil {
		t.Error("Expected error for CommentID field not of type pgtype.UUID:", convertedRow)
	}
}

func TestGenericRowToComment_InvalidUUIDValue(t *testing.T) {
	row := defaultFakeRow()
	row.CommentID.Valid = false // Make it invalid
	convertedRow, err := GenericSQLCRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid UUID value:", convertedRow)
	}
}

func TestGenericRowToComment_InvalidTimestamp(t *testing.T) {
	row := defaultFakeRow()
	row.Extract.Valid = false // Make timestamp invalid
	_, err := GenericSQLCRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid timestamp")
	}
}

// --- (Comment) ToPostCommentRow tests ---

func TestComment_ToPostCommentRow_Valid(t *testing.T) {
	comment := defaultComment()
	row := comment.ToPostCommentRow()
	if row == nil {
		t.Error("Expected non-nil PostCommentRow")
	}
	if row.ListingID != comment.TargetListing {
		t.Error("Expected ListingID ", comment.TargetListing, ", got ", row.ListingID)
	}
	if row.UserIp != comment.UserIP {
		t.Error("Expected UserIp ", comment.UserIP, ", got ", row.UserIp)
	}
	if row.UserID != comment.UserID {
		t.Error("Expected UserID ", comment.UserID, ", got ", row.UserID)
	}
	if row.Username != comment.Username {
		t.Error("Expected Username ", comment.Username, ", got ", row.Username)
	}
	if row.CommentText != comment.CommentText {
		t.Error("Expected CommentText ", comment.CommentText, ", got ", row.CommentText)
	}
	if !row.CommentID.Valid {
		t.Error("Expected valid CommentID")
	}
	if row.CommentID.Bytes != [16]byte(comment.CommentID) {
		t.Error("Expected CommentID bytes ", [16]byte(comment.CommentID), ", got ", row.CommentID.Bytes)
	}
	if !row.Extract.Valid {
		t.Error("Expected valid Extract field")
	}
	if row.Extract.Int.Int64() != comment.Timestamp {
		t.Error("Expected Extract ", comment.Timestamp, ", got ", row.Extract.Int.Int64())
	}
}

func TestComment_ToPostCommentRow_UUIDBytes(t *testing.T) {
	comment := defaultComment()
	row := comment.ToPostCommentRow()
	expectedBytes := [16]byte(comment.CommentID)
	if row.CommentID.Bytes != expectedBytes {
		t.Error("Expected UUID bytes ", expectedBytes, ", got ", row.CommentID.Bytes)
	}
}

func TestComment_ToPostCommentRow_TimestampConversion(t *testing.T) {
	comment := defaultComment()
	row := comment.ToPostCommentRow()
	if !row.Extract.Valid {
		t.Error("Expected valid Extract field")
	}
	if row.Extract.Int.Int64() != comment.Timestamp {
		t.Error("Expected Extract ", comment.Timestamp, ", got ", row.Extract.Int.Int64())
	}
}

func TestComment_ToPostCommentRow_NilReceiver(t *testing.T) {
	var comment *Comment
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling ToPostCommentRow on nil receiver")
		}
	}()
	_ = comment.ToPostCommentRow()
}

// --- CommentRowToComment tests ---

func defaultCommentRow() sqlc.GetCommentsByListingIDRow {
	commentID, _ := validPgtypeUUID()
	userID, _ := uuid.NewV7()
	timestamp := time.Now()
	return sqlc.GetCommentsByListingIDRow{
		CommentID:   commentID,
		ListingID:   "listing",
		UserIp:      "ip",
		UserID:      userID.String(),
		Username:    "name",
		CommentText: "text",
		Extract:     validPgtypeNumeric(timestamp.Unix()),
	}
}

func TestCommentRowToComment_Valid(t *testing.T) {
	row := defaultCommentRow()
	comment, err := GetCommentRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	if [16]byte(comment.CommentID) != row.CommentID.Bytes {
		t.Error("Expected comment ID ", row.CommentID, ", got ", comment.CommentID)
	}
}

func TestCommentRowToComment_InvalidUUID(t *testing.T) {
	row := defaultCommentRow()
	row.CommentID = pgtype.UUID{Bytes: [16]byte{}, Valid: false}
	convertedRow, err := GetCommentRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid comment ID format:", convertedRow)
	}
}

func TestCommentRowToComment_InvalidTimestamp(t *testing.T) {
	row := defaultCommentRow()
	row.Extract = pgtype.Numeric{Int: big.NewInt(1), Valid: false}
	_, err := GetCommentRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid timestamp")
	}
}

func TestCommentRowToComment_TimestampTooOld(t *testing.T) {
	row := defaultCommentRow()
	row.Extract = validPgtypeNumeric(1000)
	_, err := GetCommentRowToComment(row)
	if err == nil {
		t.Error("Expected error for timestamp too old")
	}
}

// --- CommentRowsToComments tests ---

func TestCommentRowsToComments_Valid(t *testing.T) {
	row := defaultCommentRow()
	rows := []sqlc.GetCommentsByListingIDRow{row}
	comments, err := GetCommentRowsToComments(rows)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(comments) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(comments))
	}
}

func TestCommentRowsToComments_InvalidRow(t *testing.T) {
	row := defaultCommentRow()
	badRow := defaultCommentRow()
	badRow.Extract = pgtype.Numeric{Int: big.NewInt(1), Valid: false}
	rows := []sqlc.GetCommentsByListingIDRow{row, badRow}
	_, err := GetCommentRowsToComments(rows)
	if err == nil {
		t.Error("Expected error for invalid row in slice")
	}
}

// --- CommentToCommentRow and CommentsToCommentRows tests ---

// defaultComment returns a Comment with preset values for testing.
func defaultComment() Comment {
	id, _ := uuid.NewV7()
	return Comment{
		TargetListing: "listing",
		CommentID:     id,
		UserIP:        "ip",
		UserID:        "user",
		Username:      "name",
		CommentText:   "text",
		Timestamp:     1748389239,
	}
}

func TestCommentToCommentRow_AndBack(t *testing.T) {
	comment := defaultComment()
	row := CommentToGetCommentRow(comment)
	// Convert back to Comment
	got, err := GetCommentRowToComment(*row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if got.CommentID != comment.CommentID || got.TargetListing != comment.TargetListing {
		t.Errorf("Round-trip conversion failed: %+v vs %+v", got, comment)
	}
}

func TestCommentsToCommentRows_Empty(t *testing.T) {
	rows := CommentsToGetCommentRows([]Comment{})
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(rows))
	}
}

// --- ToResponse and ToResponseSlice tests ---

func TestComment_ToResponse(t *testing.T) {
	comment := defaultComment()
	resp := comment.ToResponse()
	if resp.TargetListing != comment.TargetListing || resp.CommentID != comment.CommentID {
		t.Errorf("ToResponse mismatch: %+v vs %+v", resp, comment)
	}
}

func TestToResponseSlice(t *testing.T) {
	comment := defaultComment()
	comments := []Comment{comment}
	resps := ToResponseSlice(comments)
	if len(resps) != 1 {
		t.Errorf("Expected 1 response, got %d", len(resps))
	}
	if resps[0].TargetListing != comment.TargetListing {
		t.Errorf("Unexpected TargetListing: %s", resps[0].TargetListing)
	}
}
