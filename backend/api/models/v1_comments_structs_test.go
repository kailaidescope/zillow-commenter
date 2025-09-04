package models

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"zillow-commenter.com/m/db/postgres/sqlc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/jackc/pgx/v5/pgtype"
)

// ===================================================================================================================== //
//                                               Test Helper Structs                                                     //
// ===================================================================================================================== //

// ===================================================================================================================== //
//                                                   Model Tests                                                         //
// ===================================================================================================================== //

// --- GenericRowToComment tests ---

func checkGenericRowToCommentConversion(genericRow interface{}, comment *APIComment, t *testing.T) {
	reflectedRow := reflect.ValueOf(genericRow)
	if reflectedRow.Kind() == reflect.Ptr {
		reflectedRow = reflectedRow.Elem()
	}
	if reflectedRow.Kind() != reflect.Struct {
		t.Error("Input row to test generic row conversion is not a struct")
	}

	if comment == nil {
		t.Error("Expected non-nil Comment struct")
	}
	if reflectedRow.FieldByName("ListingID").Interface().(string) != comment.ListingID {
		t.Error("Expected ListingID ", reflectedRow.FieldByName("ListingID").Interface().(string), ", got ", comment.ListingID)
	}
	if reflectedRow.FieldByName("UserIp").Interface().(string) != comment.UserIP {
		t.Error("Expected UserIp ", reflectedRow.FieldByName("UserIp").Interface().(string), ", got ", comment.UserIP)
	}
	if reflectedRow.FieldByName("UserID").Interface().(string) != comment.UserID {
		t.Error("Expected UserID ", reflectedRow.FieldByName("UserID").Interface().(string), ", got ", comment.UserID)
	}
	if reflectedRow.FieldByName("Username").Interface().(string) != comment.Username {
		t.Error("Expected Username ", reflectedRow.FieldByName("Username").Interface().(string), ", got ", comment.Username)
	}
	if reflectedRow.FieldByName("CommentText").Interface().(string) != comment.CommentText {
		t.Error("Expected CommentText ", reflectedRow.FieldByName("CommentText").Interface().(string), ", got ", comment.CommentText)
	}
	if reflectedRow.FieldByName("CommentID").Interface().(pgtype.UUID).Bytes != [16]byte(comment.CommentID) {
		t.Error("Expected CommentID bytes ", reflectedRow.FieldByName("CommentID").Interface().(pgtype.UUID).Bytes, ", got ", [16]byte(comment.CommentID))
	}
	if (comment.ListingTitle == nil) == reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).Valid {
		t.Error("Expected comment.ListingTitle == nil to be ", comment.ListingTitle != nil, ", since row.ListingTitle.Valid == nil is ", reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).Valid, ", but comment.ListingTitle == nil is ", comment.ListingTitle == nil)
	}
	if comment.ListingTitle != nil && *comment.ListingTitle != reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).String {
		t.Error("Expected comment and row to have same listing title, but got comment.ListingTitle=", *comment.ListingTitle, " and row.ListingTitle.String=", reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).String)
	}
	if (comment.IPNonce == nil) == reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).Valid {
		t.Error("Expected comment.IPNonce == nil to be ", comment.IPNonce != nil, ", since row.IpNonce.Valid is ", reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).Valid, ", but comment.IPNonce == nil is ", comment.IPNonce == nil)
	}
	if comment.IPNonce != nil && *comment.IPNonce != reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).String {
		t.Error("Expected comment and row to have same ip nonce, but got comment.IPNonce=", *comment.IPNonce, " and row.IpNonce.String=", reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).String)
	}
}

// Test for arbitrary row conversion to Comment.
func TestGenericRowToComment_ValidFakeRow(t *testing.T) {
	row := sqlc.GetDefaultFakeRow()
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	if comment.ListingID != row.ListingID || comment.Username != row.Username {
		t.Error("Unexpected comment fields: ", comment)
	}
	if comment.ListingTitle == nil || *comment.ListingTitle != row.ListingTitle.String {
		t.Errorf("Expected ListingTitle '%s', got '%v'", row.ListingTitle.String, comment.ListingTitle)
	}
	checkGenericRowToCommentConversion(row, comment, t)
	//log.Println("Successfully converted fake SQLC row struct to Comment:\n\n", comment, "\n\nfrom row:\n\n", row)
}

// Test for converting PostCommentRow to Comment.
func TestGenericRowToComment_ValidPostCommentRow(t *testing.T) {
	row := sqlc.GetDefaultPostCommentRow()
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	if comment.ListingID != row.ListingID || comment.Username != row.Username {
		t.Error("Unexpected comment fields: ", comment)
	}
	if comment.ListingTitle == nil || *comment.ListingTitle != row.ListingTitle.String {
		t.Errorf("Expected ListingTitle '%s', got '%v'", row.ListingTitle.String, comment.ListingTitle)
	}
	checkGenericRowToCommentConversion(row, comment, t)
	//log.Println("Successfully converted PostCommentRow to Comment:\n\n", comment, "\n\nfrom row:\n\n", row)
}

// Test for converting GetCommentsByListingIDRow to Comment.
func TestGenericRowToComment_ValidGetCommentRow(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	if comment.ListingID != row.ListingID || comment.Username != row.Username {
		t.Error("Unexpected comment fields: ", comment)
	}
	if comment.ListingTitle == nil || *comment.ListingTitle != row.ListingTitle.String {
		t.Errorf("Expected ListingTitle '%s', got '%v'", row.ListingTitle.String, comment.ListingTitle)
	}
	checkGenericRowToCommentConversion(row, comment, t)
	//log.Println("Successfully converted GetCommentRow to Comment:\n\n", comment, "\n\nfrom row:\n\n", row)
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
	row := sqlc.GetDefaultFakeRow()
	type BadUUID struct {
		CommentID    string
		ListingID    string
		UserIp       string
		UserID       string
		Username     string
		CommentText  string
		Extract      pgtype.Numeric
		ListingTitle pgtype.Text
		IpNonce      pgtype.Text
	}
	badRow := BadUUID{
		CommentID:    "not-a-uuid",
		ListingID:    row.ListingID,
		UserIp:       row.UserIp,
		UserID:       row.UserID,
		Username:     row.Username,
		CommentText:  row.CommentText,
		Extract:      row.Extract,
		ListingTitle: row.ListingTitle,
		IpNonce:      row.IpNonce,
	}
	convertedRow, err := GenericSQLCRowToComment(badRow)
	if err == nil {
		t.Error("Expected error for CommentID field not of type pgtype.UUID:", convertedRow)
	}
}

func TestGenericRowToComment_InvalidUUIDValue(t *testing.T) {
	row := sqlc.GetDefaultFakeRow()
	row.CommentID.Valid = false // Make it invalid
	convertedRow, err := GenericSQLCRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid UUID value:", convertedRow)
	}
}

func TestGenericRowToComment_InvalidTimestamp(t *testing.T) {
	row := sqlc.GetDefaultFakeRow()
	row.Extract.Valid = false // Make timestamp invalid
	_, err := GenericSQLCRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid timestamp")
	}
}

// Checks that the timestamp is converted into the Unix second format.
func TestGenericRowToComment_TimestampFormat(t *testing.T) {
	// Create a row with a valid timestamp
	row := sqlc.GetDefaultFakeRow()

	// Convert the row to a Comment
	convertedRow, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Error("Expected no error, got ", err)
	}

	// Check that the timestamp is in Unix seconds format by comparing it to the current time.
	currentTime := time.Now().UnixMicro()
	if convertedRow.Timestamp < currentTime-1000 || convertedRow.Timestamp > currentTime+1000 {
		t.Error("Expected timestamp to be close to current time, ", currentTime, ", but got ", convertedRow.Timestamp)
	}
}

// Test for ListingTitle with Valid=false (should be nil in Comment)
func TestGenericRowToComment_ListingTitleInvalid(t *testing.T) {
	row := sqlc.GetDefaultFakeRow()
	row.ListingTitle = pgtype.Text{String: "should be nil", Valid: false}
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if comment.ListingTitle != nil {
		t.Errorf("Expected ListingTitle to be nil, got %v", comment.ListingTitle)
	}
}

// Test for missing ListingTitle field (should error)
func TestGenericRowToComment_MissingListingTitle(t *testing.T) {
	validFakeRow := sqlc.GetDefaultFakeRow()
	type Incomplete struct {
		CommentID   pgtype.UUID
		ListingID   string
		UserIp      string
		UserID      string
		Username    string
		CommentText string
		Extract     pgtype.Numeric
		IpNonce     pgtype.Text
	}
	row := Incomplete{
		CommentID:   validFakeRow.CommentID,
		ListingID:   validFakeRow.ListingID,
		UserIp:      validFakeRow.UserIp,
		UserID:      validFakeRow.UserID,
		Username:    validFakeRow.Username,
		CommentText: validFakeRow.CommentText,
		Extract:     validFakeRow.Extract,
		IpNonce:     validFakeRow.IpNonce,
	}
	_, err := GenericSQLCRowToComment(row)
	if err == nil || err.Error() != "missing ListingTitle field" {
		t.Error("Expected missing ListingTitle field error, got", err)
	}
}

// Test for IPNonce with Valid=false (should be nil in Comment)
func TestGenericRowToComment_IPNonceInvalid(t *testing.T) {
	row := sqlc.GetDefaultFakeRow()
	row.IpNonce = pgtype.Text{String: "should be nil", Valid: false}
	comment, err := GenericSQLCRowToComment(row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if comment.IPNonce != nil {
		t.Errorf("Expected IPNonce to be nil, got %v", comment.IPNonce)
	}
}

// Test for missing IpNonce field (should error)
func TestGenericRowToComment_MissingIpNonce(t *testing.T) {
	validFakeRow := sqlc.GetDefaultFakeRow()
	type Incomplete struct {
		CommentID    pgtype.UUID
		ListingID    string
		UserIp       string
		UserID       string
		Username     string
		CommentText  string
		Extract      pgtype.Numeric
		ListingTitle pgtype.Text
	}
	row := Incomplete{
		CommentID:    validFakeRow.CommentID,
		ListingID:    validFakeRow.ListingID,
		UserIp:       validFakeRow.UserIp,
		UserID:       validFakeRow.UserID,
		Username:     validFakeRow.Username,
		CommentText:  validFakeRow.CommentText,
		Extract:      validFakeRow.Extract,
		ListingTitle: validFakeRow.ListingTitle,
	}
	_, err := GenericSQLCRowToComment(row)
	if err == nil || err.Error() != "missing IpNonce field" {
		t.Error("Expected missing IpNonce field error, got", err)
	}
}

// --- (Comment) ToPostCommentRow tests ---

func checkCommentToPostParamsConversion(comment APIComment, postParams *sqlc.PostCommentParams, t *testing.T) {
	if postParams == nil {
		t.Error("Expected non-nil PostCommentRow")
	}
	if postParams.ListingID != comment.ListingID {
		t.Error("Expected ListingID ", comment.ListingID, ", got ", postParams.ListingID)
	}
	if postParams.UserIp != comment.UserIP {
		t.Error("Expected UserIp ", comment.UserIP, ", got ", postParams.UserIp)
	}
	if postParams.UserID != comment.UserID {
		t.Error("Expected UserID ", comment.UserID, ", got ", postParams.UserID)
	}
	if postParams.Username != comment.Username {
		t.Error("Expected Username ", comment.Username, ", got ", postParams.Username)
	}
	if postParams.CommentText != comment.CommentText {
		t.Error("Expected CommentText ", comment.CommentText, ", got ", postParams.CommentText)
	}
	if !postParams.CommentID.Valid {
		t.Error("Expected valid CommentID")
	}
	if postParams.CommentID.Bytes != [16]byte(comment.CommentID) {
		t.Error("Expected CommentID bytes ", [16]byte(comment.CommentID), ", got ", postParams.CommentID.Bytes)
	}
	if (comment.ListingTitle == nil) == postParams.ListingTitle.Valid {
		t.Error("Expected row.ListingTitle.Valid to be ", !postParams.ListingTitle.Valid, ", since comment.ListingTitle == nil is ", comment.ListingTitle == nil, ", but Valid=", postParams.ListingTitle.Valid)
	}
	if comment.ListingTitle != nil && *comment.ListingTitle != postParams.ListingTitle.String {
		t.Error("Expected comment and row to have same listing title, but got comment.ListingTitle=", *comment.ListingTitle, " and row.ListingTitle.String=", postParams.ListingTitle.String)
	}
	if (comment.IPNonce == nil) == postParams.IpNonce.Valid {
		t.Error("Expected row.IpNonce.Valid to be ", !postParams.IpNonce.Valid, ", since comment.IPNonce == nil is ", comment.IPNonce == nil, ", but Valid=", postParams.IpNonce.Valid)
	}
	if comment.IPNonce != nil && *comment.IPNonce != postParams.IpNonce.String {
		t.Error("Expected comment and row to have same ip nonce, but got comment.IPNonce=", *comment.IPNonce, " and row.IpNonce.String=", postParams.IpNonce.String)
	}
}

func checkCommentToGenericRowConversion(comment APIComment, genericRow interface{}, t *testing.T) {
	if genericRow == nil {
		t.Error("Expected non-nil converted row struct")
	}

	reflectedRow := reflect.ValueOf(genericRow)
	if reflectedRow.Kind() == reflect.Ptr {
		reflectedRow = reflectedRow.Elem()
	}
	if reflectedRow.Kind() != reflect.Struct {
		t.Error("Input row to test generic row conversion is not a struct")
	}

	if reflectedRow.FieldByName("ListingID").Interface().(string) != comment.ListingID {
		t.Error("Expected ListingID ", comment.ListingID, ", got ", reflectedRow.FieldByName("ListingID").Interface().(string))
	}
	if reflectedRow.FieldByName("UserIp").Interface().(string) != comment.UserIP {
		t.Error("Expected UserIp ", comment.UserIP, ", got ", reflectedRow.FieldByName("UserIp").Interface().(string))
	}
	if reflectedRow.FieldByName("UserID").Interface().(string) != comment.UserID {
		t.Error("Expected UserID ", comment.UserID, ", got ", reflectedRow.FieldByName("UserID").Interface().(string))
	}
	if reflectedRow.FieldByName("Username").Interface().(string) != comment.Username {
		t.Error("Expected Username ", comment.Username, ", got ", reflectedRow.FieldByName("Username").Interface().(string))
	}
	if reflectedRow.FieldByName("CommentText").Interface().(string) != comment.CommentText {
		t.Error("Expected CommentText ", comment.CommentText, ", got ", reflectedRow.FieldByName("CommentText").Interface().(string))
	}
	if !reflectedRow.FieldByName("CommentID").Interface().(pgtype.UUID).Valid {
		t.Error("Expected valid CommentID")
	}
	if reflectedRow.FieldByName("CommentID").Interface().(pgtype.UUID).Bytes != [16]byte(comment.CommentID) {
		t.Error("Expected CommentID bytes ", [16]byte(comment.CommentID), ", got ", reflectedRow.FieldByName("CommentID").Interface().(pgtype.UUID).Bytes)
	}
	if (comment.ListingTitle == nil) == reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).Valid {
		t.Error("Expected row.ListingTitle.Valid to be ", !reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).Valid, ", since comment.ListingTitle == nil is ", comment.ListingTitle == nil, ", but Valid=", reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).Valid)
	}
	if comment.ListingTitle != nil && *comment.ListingTitle != reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).String {
		t.Error("Expected comment and row to have same listing title, but got comment.ListingTitle=", *comment.ListingTitle, " and row.ListingTitle.String=", reflectedRow.FieldByName("ListingTitle").Interface().(pgtype.Text).String)
	}
	if (comment.IPNonce == nil) == reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).Valid {
		t.Error("Expected row.IpNonce.Valid to be ", !reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).Valid, ", since comment.IPNonce == nil is ", comment.IPNonce == nil, ", but Valid=", reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).Valid)
	}
	if comment.IPNonce != nil && *comment.IPNonce != reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).String {
		t.Error("Expected comment and row to have same ip nonce, but got comment.IPNonce=", *comment.IPNonce, " and row.IpNonce.String=", reflectedRow.FieldByName("IpNonce").Interface().(pgtype.Text).String)
	}
}

func TestComment_ToPostCommentParams_Valid(t *testing.T) {
	comment := GetDefaultAPIComment()
	row := comment.ToPostCommentParams()

	// Test row conversion
	checkCommentToGenericRowConversion(comment, row, t)
}

func TestComment_ToPostCommentParams_Valid_NilListingTitle(t *testing.T) {
	comment := GetDefaultAPIComment()
	comment.ListingTitle = nil
	row := comment.ToPostCommentParams()

	// Test row conversion
	checkCommentToGenericRowConversion(comment, row, t)
}

func TestComment_ToPostCommentParams_Valid_NilIPNonce(t *testing.T) {
	comment := GetDefaultAPIComment()
	comment.IPNonce = nil
	row := comment.ToPostCommentParams()

	// Test row conversion
	checkCommentToGenericRowConversion(comment, row, t)
}

func TestComment_ToPostCommentParams_UUIDBytes(t *testing.T) {
	comment := GetDefaultAPIComment()
	row := comment.ToPostCommentParams()
	expectedBytes := [16]byte(comment.CommentID)
	if row.CommentID.Bytes != expectedBytes {
		t.Error("Expected UUID bytes ", expectedBytes, ", got ", row.CommentID.Bytes)
	}
}

func TestComment_ToPostCommentParams_NilReceiver(t *testing.T) {
	var comment *APIComment
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling ToPostCommentRow on nil receiver")
		}
	}()
	_ = comment.ToPostCommentParams()
}

// --- GetCommentRowToComment tests ---

func TestCommentRowToComment_Valid(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	comment, err := GetCommentRowToComment(row)
	if err != nil {
		t.Fatal("Expected no error, got ", err)
	}
	checkGenericRowToCommentConversion(row, comment, t)
}

func TestCommentRowToComment_InvalidUUID(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	row.CommentID = pgtype.UUID{Bytes: [16]byte{}, Valid: false}
	convertedRow, err := GetCommentRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid comment ID format:", convertedRow)
	}
}

func TestCommentRowToComment_InvalidTimestamp(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	row.Extract = pgtype.Numeric{Int: big.NewInt(1), Valid: false}
	_, err := GetCommentRowToComment(row)
	if err == nil {
		t.Error("Expected error for invalid timestamp")
	}
}

func TestCommentRowToComment_TimestampTooOld(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	row.Extract = sqlc.GetValidPGTypeNumeric(1000)
	_, err := GetCommentRowToComment(row)
	if err == nil {
		t.Error("Expected error for timestamp too old")
	}
}

// Checks that the timestamp is converted into the Unix second format.
func TestCommentRowToComment_TimestampFormat(t *testing.T) {
	// Create a row with a valid timestamp
	row := sqlc.GetDefaultGetCommentRow()

	// Convert the row to a Comment
	convertedRow, err := GetCommentRowToComment(row)
	if err != nil {
		t.Error("Expected no error, got ", err)
	}

	// Check that the timestamp is in Unix seconds format by comparing it to the current time.
	currentTime := time.Now().UnixMicro()
	if convertedRow.Timestamp < currentTime-1000 || convertedRow.Timestamp > currentTime+1000 {
		t.Error("Expected timestamp to be close to current time, ", currentTime, ", but got ", convertedRow.Timestamp)
	}
}

func TestGetCommentRowToComment_ValidListingTitle(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	comment, err := GetCommentRowToComment(row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if comment.ListingTitle == nil || *comment.ListingTitle != row.ListingTitle.String {
		t.Errorf("Expected ListingTitle '%s', got '%v'", row.ListingTitle.String, comment.ListingTitle)
	}
}

func TestGetCommentRowToComment_InvalidListingTitle(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	row.ListingTitle = pgtype.Text{String: "", Valid: false}
	comment, err := GetCommentRowToComment(row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if comment.ListingTitle != nil {
		t.Errorf("Expected ListingTitle to be nil, since input pgtype is invalid, got '%v'", comment.ListingTitle)
	}
}

func TestGetCommentRowToComment_ValidIPNonce(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	comment, err := GetCommentRowToComment(row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if comment.IPNonce == nil || *comment.IPNonce != row.IpNonce.String {
		t.Errorf("Expected IPNonce '%s', got '%v'", row.IpNonce.String, comment.IPNonce)
	}
}

func TestGetCommentRowToComment_InvalidIPNonce(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	row.IpNonce = pgtype.Text{String: "", Valid: false}
	comment, err := GetCommentRowToComment(row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if comment.IPNonce != nil {
		t.Errorf("Expected IPNonce to be nil, since input pgtype is invalid, got '%v'", comment.IPNonce)
	}
}

// --- GetCommentRowsToComments tests ---

func TestGetCommentRowsToComments_Valid(t *testing.T) {
	rows := []sqlc.GetCommentsByListingIDRow{sqlc.GetDefaultGetCommentRow(), sqlc.GetDefaultGetCommentRow(), sqlc.GetDefaultGetCommentRow(), sqlc.GetDefaultGetCommentRow()}
	comments, err := GetCommentRowsToComments(rows)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(comments) != len(rows) {
		t.Errorf("Expected %d comment, got %d", len(rows), len(comments))
	}
	for i, comment := range comments {
		t.Log("Testing comment #", i)

		// Check each conversion
		checkGenericRowToCommentConversion(rows[i], &comment, t)
	}
}

func TestGetCommentRowsToComments_InvalidRow(t *testing.T) {
	badRow := sqlc.GetDefaultGetCommentRow()
	badRow.Extract = pgtype.Numeric{Int: big.NewInt(1), Valid: false}
	rows := []sqlc.GetCommentsByListingIDRow{sqlc.GetDefaultGetCommentRow(), sqlc.GetDefaultGetCommentRow(), badRow, sqlc.GetDefaultGetCommentRow()}
	_, err := GetCommentRowsToComments(rows)
	if err == nil {
		t.Error("Expected error for invalid row in slice")
	}
}

func TestGetCommentRowsToComments_ListingTitle(t *testing.T) {
	row := sqlc.GetDefaultGetCommentRow()
	rows := []sqlc.GetCommentsByListingIDRow{row, sqlc.GetDefaultGetCommentRow(), sqlc.GetDefaultGetCommentRow()}
	comments, err := GenericSQLCRowsToComments(rows)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(comments) != len(rows) {
		t.Fatalf("Expected %d comment, got %d", len(rows), len(comments))
	}
	if comments[0].ListingTitle == nil || *comments[0].ListingTitle != row.ListingTitle.String {
		t.Errorf("Expected ListingTitle '%s', got '%v'", row.ListingTitle.String, comments[0].ListingTitle)
	}
	for i, comment := range comments {
		t.Log("Testing comment #", i)

		// Check each conversion
		checkGenericRowToCommentConversion(rows[i], &comment, t)
	}
}

// --- CommentToGetCommentRow and CommentsToGetCommentRows tests ---

func TestCommentToGetCommentRow_AndBack(t *testing.T) {
	comment := GetDefaultAPIComment()
	row := CommentToGetCommentRow(comment)
	checkCommentToGenericRowConversion(comment, row, t)
	// Convert back to Comment
	got, err := GetCommentRowToComment(*row)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if got.CommentID != comment.CommentID || got.ListingID != comment.ListingID {
		t.Errorf("Round-trip conversion failed: %+v vs %+v", got, comment)
	}
	checkGenericRowToCommentConversion(row, &comment, t)
}

func TestCommentToGetCommentRow_ListingTitle(t *testing.T) {
	comment := GetDefaultAPIComment()
	row := CommentToGetCommentRow(comment)
	if row.ListingTitle.String != *comment.ListingTitle {
		t.Errorf("Expected ListingTitle '%s', got '%s'", *comment.ListingTitle, row.ListingTitle.String)
	}
	if !row.ListingTitle.Valid {
		t.Error("Expected ListingTitle to be valid")

		checkCommentToGenericRowConversion(comment, row, t)
	}
}

func TestCommentToGetCommentRow_NilListingTitle(t *testing.T) {
	comment := GetDefaultAPIComment()
	comment.ListingTitle = nil
	row := CommentToGetCommentRow(comment)
	if row.ListingTitle.Valid {
		t.Error("Expected ListingTitle to be invalid when Comment.ListingTitle is nil")
	}
	checkCommentToGenericRowConversion(comment, row, t)
}

func TestCommentToGetCommentRow_IPNonce(t *testing.T) {
	comment := GetDefaultAPIComment()
	row := CommentToGetCommentRow(comment)
	if row.IpNonce.String != *comment.IPNonce {
		t.Errorf("Expected IPNonce '%s', got '%s'", *comment.IPNonce, row.IpNonce.String)
	}
	if !row.IpNonce.Valid {
		t.Error("Expected IPNonce to be valid")

		checkCommentToGenericRowConversion(comment, row, t)
	}
}

func TestCommentToGetCommentRow_NilIPNonce(t *testing.T) {
	comment := GetDefaultAPIComment()
	comment.IPNonce = nil
	row := CommentToGetCommentRow(comment)
	if row.IpNonce.Valid {
		t.Error("Expected IPNonce to be invalid when Comment.IPNonce is nil")
	}
	checkCommentToGenericRowConversion(comment, row, t)
}

func TestCommentsToGetCommentRows_Empty(t *testing.T) {
	rows := CommentsToGetCommentRows([]APIComment{})
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(rows))
	}
}

func TestCommentsToGetCommentRows_ListingTitle(t *testing.T) {
	comment := GetDefaultAPIComment()
	comments := []APIComment{comment, comment}
	rows := CommentsToGetCommentRows(comments)
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}
	if rows[1].ListingTitle.String != *comment.ListingTitle {
		t.Errorf("Expected ListingTitle '%s', got '%s'", *comment.ListingTitle, rows[1].ListingTitle.String)
	}
	if !rows[1].ListingTitle.Valid {
		t.Error("Expected ListingTitle to be valid")
	}
	for i, row := range rows {
		t.Log("Testing row #", i)

		// Check each conversion
		checkCommentToGenericRowConversion(comments[i], row, t)
	}
}

func TestCommentsToGetCommentRows_IPNonce(t *testing.T) {
	comment := GetDefaultAPIComment()
	comments := []APIComment{comment, comment}
	rows := CommentsToGetCommentRows(comments)
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}
	if rows[1].IpNonce.String != *comment.IPNonce {
		t.Errorf("Expected IPNonce '%s', got '%s'", *comment.IPNonce, rows[1].IpNonce.String)
	}
	if !rows[1].IpNonce.Valid {
		t.Error("Expected IPNonce to be valid")
	}
	for i, row := range rows {
		t.Log("Testing row #", i)

		// Check each conversion
		checkCommentToGenericRowConversion(comments[i], row, t)
	}
}

// --- ToResponse and ToResponseSlice tests ---

func TestComment_ToResponse(t *testing.T) {
	comment := GetDefaultAPIComment()
	resp := comment.ToResponse()
	if resp.TargetListing != comment.ListingID || resp.CommentID != comment.CommentID {
		t.Errorf("ToResponse mismatch: %+v vs %+v", resp, comment)
	}
}

func TestToResponseSlice(t *testing.T) {
	comment := GetDefaultAPIComment()
	comments := []APIComment{comment}
	resps := ToResponseSlice(comments)
	if len(resps) != 1 {
		t.Errorf("Expected 1 response, got %d", len(resps))
	}
	if resps[0].TargetListing != comment.ListingID {
		t.Errorf("Unexpected TargetListing: %s", resps[0].TargetListing)
	}
}

// ================================================================================================================= //
//                                          Helper function Tests                                                    //
// ================================================================================================================= //

// --- convertPGListingTitleToAPI and convertAPIListingTitleToPG tests ---

func TestConvertPGTextToAPI_Valid(t *testing.T) {
	pgText := pgtype.Text{String: "Sample Title", Valid: true}
	apiString := convertPGTextToString(pgText)
	if apiString == nil {
		t.Fatal("Expected non-nil API string for valid pgtype.Text")
	}
	if *apiString != pgText.String {
		t.Errorf("Expected API string '%s', got '%s'", pgText.String, *apiString)
	}
}

func TestConvertPGTextToAPI_Invalid(t *testing.T) {
	pgText := pgtype.Text{String: "Should be nil", Valid: false}
	apiString := convertPGTextToString(pgText)
	if apiString != nil {
		t.Errorf("Expected nil API string for invalid pgtype.Text, got '%v'", *apiString)
	}
}

func TestConvertAPIStringToPG_NonNil(t *testing.T) {
	apiString := aws.String("API Title")
	pgText := convertStringToPGText(apiString)
	if !pgText.Valid {
		t.Error("Expected pgtype.Text.Valid to be true for non-nil API string")
	}
	if pgText.String != *apiString {
		t.Errorf("Expected pgtype.Text.String '%s', got '%s'", *apiString, pgText.String)
	}
}

func TestConvertAPIStringToPG_Nil(t *testing.T) {
	var apiString *string = nil
	pgText := convertStringToPGText(apiString)
	if pgText.Valid {
		t.Error("Expected pgtype.Text.Valid to be false for nil API string")
	}
	if pgText.String != "" {
		t.Errorf("Expected pgtype.Text.String to be empty for nil API string, got '%s'", pgText.String)
	}
}
