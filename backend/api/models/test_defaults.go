package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
)

// -- For CommentToCommentRow tests ---

// GetDefaultAPIComment returns a Comment with preset values for testing.
func GetDefaultAPIComment() APIComment {
	id, _ := uuid.NewV7()
	return APIComment{
		ListingID:    "0a0a0a",
		CommentID:    id,
		UserIP:       "9f67720a05fb8ca4781f1cb5fc60b8ab7a2b068bf2be9f0660",
		UserID:       "user",
		Username:     "name",
		CommentText:  "text",
		Timestamp:    time.Now().UnixMicro(),
		ListingTitle: aws.String("test title"),
		IPNonce:      aws.String("96dc49a63b34dcf9229c0ed5"),
		ListingType:  "apt",
	}
}
