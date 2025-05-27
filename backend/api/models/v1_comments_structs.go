// The models package contains the data structures used in the API.
package models

type Comment struct {
	TargetListing string `json:"target_listing"`
	CommentID     string `json:"comment_id"`
	UserIP        string `json:"user_ip"`
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	CommentText   string `json:"comment_text"`
	Timestamp     int64  `json:"timestamp"`
}

type ResponseComment struct {
	TargetListing string `json:"target_listing"`
	CommentID     string `json:"comment_id"`
	Username      string `json:"username"`
	CommentText   string `json:"comment_text"`
	Timestamp     int64  `json:"timestamp"`
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
	TempCommentDB["32707340"] = []Comment{
		{
			TargetListing: "32707340",
			CommentID:     "cmt1",
			UserIP:        "",
			UserID:        "",
			Username:      "homebuyer123",
			CommentText:   "Beautiful house! Love the backyard.",
			Timestamp:     1719900000,
		},
		{
			TargetListing: "32707340",
			CommentID:     "cmt2",
			UserIP:        "",
			UserID:        "",
			Username:      "nyhousefan",
			CommentText:   "Is the basement finished?",
			Timestamp:     1719903600,
		},
		{
			TargetListing: "32707340",
			CommentID:     "cmt5",
			UserIP:        "",
			UserID:        "",
			Username:      "longislandmom",
			CommentText:   "How old is the roof?",
			Timestamp:     1719904000,
		},
		{
			TargetListing: "32707340",
			CommentID:     "cmt6",
			UserIP:        "",
			UserID:        "",
			Username:      "commackdad",
			CommentText:   "Nice curb appeal. Any recent renovations?",
			Timestamp:     1719904200,
		},
		{
			TargetListing: "32707340",
			CommentID:     "cmt7",
			UserIP:        "",
			UserID:        "",
			Username:      "firsttimebuyer",
			CommentText:   "Is there an open house this weekend?",
			Timestamp:     1719904400,
		},
		{
			TargetListing: "32707340",
			CommentID:     "cmt8",
			UserIP:        "",
			UserID:        "",
			Username:      "petlover",
			CommentText:   "Is the yard fenced in for dogs?",
			Timestamp:     1719904600,
		},
		{
			TargetListing: "32707340",
			CommentID:     "cmt9",
			UserIP:        "",
			UserID:        "",
			Username:      "zillowfan",
			CommentText:   "Price seems fair for the area.",
			Timestamp:     1719904800,
		},
		{
			TargetListing: "32707340",
			CommentID:     "cmt10",
			UserIP:        "",
			UserID:        "",
			Username:      "investorjoe",
			CommentText:   "What are the property taxes?",
			Timestamp:     1719905000,
		},
	}

	TempCommentDB["32692760"] = []Comment{
		{
			TargetListing: "32692760",
			CommentID:     "cmt3",
			UserIP:        "",
			UserID:        "",
			Username:      "commacklocal",
			CommentText:   "Great neighborhood, lived here for years.",
			Timestamp:     1719907200,
		},
		{
			TargetListing: "32692760",
			CommentID:     "cmt4",
			UserIP:        "",
			UserID:        "",
			Username:      "zillowuser",
			CommentText:   "Does anyone know about the school district?",
			Timestamp:     1719910800,
		},
		{
			TargetListing: "32692760",
			CommentID:     "cmt11",
			UserIP:        "",
			UserID:        "",
			Username:      "familyman",
			CommentText:   "Perfect for a growing family.",
			Timestamp:     1719911000,
		},
		{
			TargetListing: "32692760",
			CommentID:     "cmt12",
			UserIP:        "",
			UserID:        "",
			Username:      "househunter",
			CommentText:   "How many bathrooms?",
			Timestamp:     1719911200,
		},
		{
			TargetListing: "32692760",
			CommentID:     "cmt13",
			UserIP:        "",
			UserID:        "",
			Username:      "retireeinny",
			CommentText:   "Quiet street, close to parks.",
			Timestamp:     1719911400,
		},
		{
			TargetListing: "32692760",
			CommentID:     "cmt14",
			UserIP:        "",
			UserID:        "",
			Username:      "commackmom",
			CommentText:   "Is there a finished basement?",
			Timestamp:     1719911600,
		},
		{
			TargetListing: "32692760",
			CommentID:     "cmt15",
			UserIP:        "",
			UserID:        "",
			Username:      "dogowner",
			CommentText:   "Any restrictions on pets?",
			Timestamp:     1719911800,
		},
		{
			TargetListing: "32692760",
			CommentID:     "cmt16",
			UserIP:        "",
			UserID:        "",
			Username:      "nyrealestate",
			CommentText:   "Looks recently updated!",
			Timestamp:     1719912000,
		},
	}
}
