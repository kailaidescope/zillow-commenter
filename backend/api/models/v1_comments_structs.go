// The models package contains the data structures used in the API.
//
// Notes:
//   - The timestamps (generated by NeonSQL's postgres database using EXTRACT) are in microseconds since the epoch. They are stored in a pgtype.Numeric type, which wraps a big.Int. Big ints wrap an int64 value. Assuming that the conversion is lossless, the timestamp is valid for all times within the next 292,000 years. That's a lot, so we don't need to worry about it for now. If you are finding this comment in 292,000 years, please remember my species—humanity. We made a lot of mistakes, but we tried, very hard, to be good people.
package models

import (
	"errors"
	"math/big"
	"reflect"

	"zillow-commenter.com/m/db/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Comment struct {
	TargetListing string    `json:"listing_id"`
	CommentID     uuid.UUID `json:"comment_id"`
	UserIP        string    `json:"user_ip"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	CommentText   string    `json:"comment_text"`
	Timestamp     int64     `json:"timestamp"`
}

type ResponseComment struct {
	TargetListing string    `json:"listing_id"`
	CommentID     uuid.UUID `json:"comment_id"`
	Username      string    `json:"username"`
	CommentText   string    `json:"comment_text"`
	Timestamp     int64     `json:"timestamp"`
}

// GenericRowToComment converts any struct with the required fields to a Comment object.
// The input must be a struct with fields: CommentID (pgtype.UUID), ListingID (string), UserIp (string),
// UserID (string), Username (string), CommentText (string), Extract (pgtype.Numeric).
//
// Input:
//   - row: an interface{} that is expected to be a struct with the required fields.
//
// Output:
//   - *Comment: a pointer to a Comment struct containing the comment data.
//   - error: an error if the conversion fails, otherwise nil.
func GenericRowToComment(row interface{}) (*Comment, error) {
	v := reflect.ValueOf(row)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("input is not a struct")
	}

	getField := func(name string) (reflect.Value, bool) {
		f := v.FieldByName(name)
		return f, f.IsValid()
	}

	// Extract CommentID
	commentIDField, ok := getField("CommentID")
	if !ok {
		return nil, errors.New("missing CommentID field")
	}
	uuidBytes, ok := commentIDField.Interface().(pgtype.UUID)
	if !ok {
		return nil, errors.New("CommentID field is not of type pgtype.UUID")
	}
	if !uuidBytes.Valid {
		return nil, errors.New("CommentID field is not valid")
	}

	commentUUID, err := uuid.ParseBytes(uuidBytes.Bytes[:])
	if err != nil {
		return nil, errors.Join(err, errors.New("invalid comment ID format"))
	}

	// Extract ListingID
	listingIDField, ok := getField("ListingID")
	if !ok {
		return nil, errors.New("missing ListingID field")
	}
	listingID := listingIDField.String()

	// Extract UserIp
	userIPField, ok := getField("UserIp")
	if !ok {
		return nil, errors.New("missing UserIp field")
	}
	userIP := userIPField.String()

	// Extract UserID
	userIDField, ok := getField("UserID")
	if !ok {
		return nil, errors.New("missing UserID field")
	}
	userID := userIDField.String()

	// Extract Username
	usernameField, ok := getField("Username")
	if !ok {
		return nil, errors.New("missing Username field")
	}
	username := usernameField.String()

	// Extract CommentText
	commentTextField, ok := getField("CommentText")
	if !ok {
		return nil, errors.New("missing CommentText field")
	}
	commentText := commentTextField.String()

	// Extract Timestamp
	extractField, ok := getField("Extract")
	if !ok {
		return nil, errors.New("missing Extract field")
	}
	extract := extractField.Interface().(pgtype.Numeric)
	if !extract.Valid {
		return nil, errors.New("timestamp is not valid")
	}
	int8Value, err := extract.Int64Value()
	if err != nil {
		return nil, errors.Join(err, errors.New("error converting timestamp to int8"))
	}
	if !int8Value.Valid || int8Value.Int64 < 1748389238 {
		return nil, errors.New("timestamp is not valid")
	}
	timestamp := int8Value.Int64

	return &Comment{
		TargetListing: listingID,
		CommentID:     commentUUID,
		UserIP:        userIP,
		UserID:        userID,
		Username:      username,
		CommentText:   commentText,
		Timestamp:     timestamp,
	}, nil
}

// CommentRowToComment converts a postgres database row from GetCommentsByListingID to a Comment struct used by the API.
//
// Input:
//   - row: a sqlc.GetCommentsByListingIDRow struct containing the comment data from the database.
//
// Output:
//   - Comment: a Comment struct containing the comment data.
//   - error: an error if the conversion fails, otherwise nil.
func CommentRowToComment(row sqlc.GetCommentsByListingIDRow) (*Comment, error) {
	// Convert postgres types to Go types.

	// Convert the comment ID from pgtype.UUID to uuid.UUID.
	commentUUID, err := uuid.FromBytes(row.CommentID.Bytes[:])
	if err != nil {
		// If the conversion fails, return an error indicating the format is invalid.
		return nil, errors.Join(errors.New("invalid comment ID format"), err)
	}

	// Convert the timestamp from pgtype.Numeric to int64.
	if !row.Extract.Valid {
		return nil, errors.New("timestamp is not valid")
	}
	// Ensure the timestamp is a valid int64.
	timestamp := row.Extract.Int.Int64()
	// Check if the timestamp is valid (greater than a reference time [May 27th, 2025]).
	if !row.Extract.Valid || timestamp < 1748389238 {
		return nil, errors.New("timestamp is not valid")
	}

	// Convert a database row to a Comment struct.
	return &Comment{
		TargetListing: row.ListingID,
		CommentID:     commentUUID,
		UserIP:        row.UserIp,
		UserID:        row.UserID,
		Username:      row.Username,
		CommentText:   row.CommentText,
		Timestamp:     timestamp,
	}, nil
}

// CommentRowsToComments converts a slice of sqlc.GetCommentsByListingIDRow to a slice of Comment structs.
func CommentRowsToComments(rows []sqlc.GetCommentsByListingIDRow) ([]Comment, error) {
	var comments []Comment
	for _, row := range rows {
		comment, err := CommentRowToComment(row)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to convert comment row to Comment struct"))
		}
		comments = append(comments, *comment)
	}
	return comments, nil
}

// CommentToCommentRow converts a Comment struct used by the API to a sqlc.GetCommentsByListingIDRow struct used by postgres.
//
// Input:
//   - comment: a Comment struct containing the comment data.
//
// Output:
//   - sqlc.GetCommentsByListingIDRow: a sqlc.GetCommentsByListingIDRow struct containing the comment data.
//   - error: an error if the conversion fails, otherwise nil.
func CommentToCommentRow(comment Comment) *sqlc.GetCommentsByListingIDRow {
	// Convert go types to postgres types.

	// Convert the timestamp to pgtype.Numeric.
	extract := pgtype.Numeric{
		Int:   big.NewInt(comment.Timestamp),
		Valid: true,
	}

	// Create a GetCommentsByListingIDRow struct from the Comment struct.
	return &sqlc.GetCommentsByListingIDRow{
		CommentID:   pgtype.UUID{Bytes: [16]byte(comment.CommentID), Valid: true},
		ListingID:   comment.TargetListing,
		UserIp:      comment.UserIP,
		UserID:      comment.UserID,
		Username:    comment.Username,
		CommentText: comment.CommentText,
		Extract:     extract,
	}
}

// CommentsToCommentRows converts a slice of Comment structs to a slice of sqlc.GetCommentsByListingIDRow structs.
func CommentsToCommentRows(comments []Comment) []sqlc.GetCommentsByListingIDRow {
	var commentRows []sqlc.GetCommentsByListingIDRow
	for _, comment := range comments {
		commentRow := CommentToCommentRow(comment)
		commentRows = append(commentRows, *commentRow)
	}
	return commentRows
}

// ToResponse converts a Comment to a ResponseComment.
// This is used to format the comment data for API responses, excluding sensitive information like UserIP and UserID.
func (c Comment) ToResponse() ResponseComment {
	return ResponseComment{
		TargetListing: c.TargetListing,
		CommentID:     c.CommentID,
		Username:      c.Username,
		CommentText:   c.CommentText,
		Timestamp:     c.Timestamp,
	}
}

// ToResponseSlice converts a slice of Comment to a slice of ResponseComment.
func ToResponseSlice(comments []Comment) []ResponseComment {
	var response []ResponseComment
	for _, comment := range comments {
		response = append(response, comment.ToResponse())
	}

	return response
}

var TempCommentDB = map[string][]Comment{}

// TempCommentDB is a temporary in-memory database for comments.
// The key is the listing ID, and the value is a slice of comments for that listing.
// This is used for demonstration purposes and should be replaced with a proper database in production.

func InitTempCommentDB() {
	// Reference times
	now := int64(1748366686) // today
	oneDay := int64(86400)

	// Helper to generate a new V7 UUID or panic if error
	newV7 := func() uuid.UUID {
		id, err := uuid.NewV7()
		if err != nil {
			panic(err)
		}
		return id
	}

	TempCommentDB["32707340"] = []Comment{
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "oldtimer1",
			CommentText:   "I remember when this house was first built!",
			Timestamp:     now - 10*oneDay, // 10 days ago
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "historybuff",
			CommentText:   "This property has a lot of history.",
			Timestamp:     now - 8*oneDay, // 8 days ago
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "homebuyer123",
			CommentText:   "Beautiful house! Love the backyard.",
			Timestamp:     now - 6*oneDay, // 6 days ago
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "nyhousefan",
			CommentText:   "Is the basement finished?",
			Timestamp:     now - 5*oneDay, // 5 days ago
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "longislandmom",
			CommentText:   "How old is the roof?",
			Timestamp:     now - 4*oneDay, // 4 days ago
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "commackdad",
			CommentText:   "Nice curb appeal. Any recent renovations?",
			Timestamp:     now - 3*oneDay, // 3 days ago
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "firsttimebuyer",
			CommentText:   "Is there an open house this weekend?",
			Timestamp:     now - 2*oneDay, // 2 days ago
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "petlover",
			CommentText:   "Is the yard fenced in for dogs?",
			Timestamp:     now - oneDay, // yesterday
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "zillowfan",
			CommentText:   "Price seems fair for the area.",
			Timestamp:     now, // today
		},
		{
			TargetListing: "32707340",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "investorjoe",
			CommentText:   "What are the property taxes?",
			Timestamp:     now, // today
		},
	}

	TempCommentDB["32692760"] = []Comment{
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "veteranresident",
			CommentText:   "Moved here 15 years ago, still love it.",
			Timestamp:     now - 12*oneDay, // 12 days ago
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "oldschool",
			CommentText:   "Neighborhood has changed a lot over the years.",
			Timestamp:     now - 9*oneDay, // 9 days ago
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "commacklocal",
			CommentText:   "Great neighborhood, lived here for years.",
			Timestamp:     now - 7*oneDay, // 7 days ago
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "zillowuser",
			CommentText:   "Does anyone know about the school district?",
			Timestamp:     now - 5*oneDay, // 5 days ago
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "familyman",
			CommentText:   "Perfect for a growing family.",
			Timestamp:     now - 3*oneDay, // 3 days ago
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "househunter",
			CommentText:   "How many bathrooms?",
			Timestamp:     now - 2*oneDay, // 2 days ago
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "retireeinny",
			CommentText:   "Quiet street, close to parks.",
			Timestamp:     now - oneDay, // yesterday
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "commackmom",
			CommentText:   "Is there a finished basement?",
			Timestamp:     now, // today
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "dogowner",
			CommentText:   "Any restrictions on pets?",
			Timestamp:     now, // today
		},
		{
			TargetListing: "32692760",
			CommentID:     newV7(),
			UserIP:        "",
			UserID:        "",
			Username:      "nyrealestate",
			CommentText:   "Looks recently updated!",
			Timestamp:     now, // today
		},
	}
}
