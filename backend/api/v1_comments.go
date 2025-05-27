package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"slices"
	"time"

	"zillow-commenter.com/m/db/postgres/sqlc"

	"github.com/gin-gonic/gin"
	"zillow-commenter.com/m/api/models"
)

// GetListingComments returns a list of comments for a specific zilllow listing.
//
// GET api/v1/comments/:listing_id
//
// Input:
//   - listing_id: The zillow listing ID for which to retrieve comments.
//
// Output:
//   - 200: A JSON array of comments for the specified listing. Comment structure defined in models package.
//   - 404: If the listing does not exist.
//   - 500: Internal server error if something goes wrong.
func (server *Server) GetListingComments(c *gin.Context) {
	// Get information from the request context
	listingID := c.Param("listing_id")
	userIP := c.ClientIP()
	timestamp := time.Now().Unix()

	log.Println("GetListingComments called with listing_id:", listingID, "\nfrom IP:", userIP, "\nat timestamp:", timestamp)

	// Check if the listing exists in the temporary comment database
	comments, err := server.getComments(listingID)
	if err != nil {
		log.Println("Listing not found:", listingID)
		// Create empty response if listing does not exist
		responseComments := []models.ResponseComment{}
		c.JSON(200, responseComments)
		return
	}

	// Prepare the response comments
	responseComments := models.ToResponseSlice(comments)

	// Return the comments as a JSON response
	c.JSON(http.StatusOK, responseComments)
}

// PostListingComment creates a new comment for a specific zillow listing.
//
// POST api/v1/comments
//
// Input:
//
//	Post form containing the following fields:
//	- listing_id: The zillow listing ID to which the comment is related.
//	- user_id: The ID of the user making the comment.
//	- username: The username of the user making the comment.
//	- comment_text: The text of the comment.
//
// Output:
//   - 201: A JSON object representing the created comment.
//   - 400: If the input data is invalid.
//   - 500: Internal server error if something goes wrong.
func (server *Server) PostListingComment(c *gin.Context) {
	// Get information from the request context
	userIP := c.ClientIP()
	timestamp := time.Now().Unix()

	// Get postform data
	listingID := c.PostForm("listing_id")
	userID := c.PostForm("user_id")
	username := c.PostForm("username")
	commentText := c.PostForm("comment_text")

	log.Printf("PostListingComment called with listing_id: %s, user_id: %s, username: %s, comment_text: %s\nfrom IP: %s\nat timestamp: %d",
		listingID, userID, username, commentText, userIP, timestamp)

	// Validate input data
	{
		if listingID == "" || userID == "" || username == "" || commentText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		if len(commentText) > 300 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Comment text exceeds maximum length of 300 characters"})
			return
		}

		if len(username) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username exceeds maximum length of 50 characters"})
			return
		}
	}

	// Create a new comment
	newComment := models.Comment{
		TargetListing: listingID,
		CommentID:     "cmt" + time.Now().Format("20060102150405"), // Unique comment ID based on timestamp
		UserIP:        userIP,
		UserID:        userID,
		Username:      username,
		CommentText:   commentText,
		Timestamp:     timestamp,
	}

	// Add the comment to the temporary comment database
	if _, error := server.getComments(listingID); error != nil {
		models.TempCommentDB[listingID] = []models.Comment{}
	}

	models.TempCommentDB[listingID] = append(models.TempCommentDB[listingID], newComment)

	// Log the new comment creation
	log.Println("New comment created for listing:", listingID, "by user:", username, "at timestamp:", timestamp)

	// Return the new comments list as a JSON response
	comments, err := server.getComments(listingID)
	if err != nil {
		log.Println("Error retrieving comments for listing:", listingID, "-", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	responseComments := models.ToResponseSlice(comments)
	//log.Println("Response comments for listing:", listingID, ":", responseComments)
	c.JSON(http.StatusCreated, responseComments)
}

// Helper function to get comments for a specific listing.
//
// Input:
//   - listingID: The zillow listing ID for which to retrieve comments.
//
// Output:
//   - A slice of Comment structs containing the comments for the specified listing.
//   - An error if the listing doesn't exist in the DB.
func (server Server) getComments(listingID string) ([]models.Comment, error) {
	// Acquire a Postgres connection from the pool
	postgresPool, err := server.GetPostgresPool().Acquire(context.TODO())
	if err != nil {
		log.Println("Error acquiring Postgres connection:", err)
		return nil, errors.Join(err, errors.New("failed to acquire postgres connection"))
	}
	defer postgresPool.Release()
	postgresQueryClient := sqlc.New(postgresPool)

	// Query the database for comments by listing ID
	commentRows, err := postgresQueryClient.GetCommentsByListingID(context.TODO(), listingID)
	if err != nil {
		log.Println("Error retrieving comments from database for listing:", listingID, "-", err)
		return nil, errors.Join(err, errors.New("failed to retrieve comments from database"))
	}

	//

	comments, exists := models.TempCommentDB[listingID]
	if !exists {
		return nil, errors.New("Listing does not exist in TempDB") // Assuming ErrListingNotFound is defined in models package
	}

	slices.SortStableFunc(comments, func(a, b models.Comment) int {
		return int(b.Timestamp) - int(a.Timestamp) // Sort by timestamp in descending order
	})

	return comments, nil
}
