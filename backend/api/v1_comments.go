package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"zillow-commenter.com/m/db/postgres/sqlc"

	ginadaptercore "github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	userIP, err := getUserIP(c)
	if err != nil {
		log.Println("Error getting user IP:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	timestamp := time.Now().Unix()

	log.Println("GetListingComments called with listing_id:", listingID, "\nfrom IP:", userIP, "\nat timestamp:", timestamp)

	// Check if the listing exists in the temporary comment database
	comments, err := server.getComments(listingID)
	if err != nil {
		log.Println("Error getting comments from db", listingID)

		// Tell the client that something went wrong
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Prepare the response comments
	responseComments := models.ToResponseSlice(comments)

	// Return the comments as a JSON response
	c.JSON(http.StatusOK, responseComments)
	log.Println("Successfully returning comments for listing:", listingID, ":", responseComments)
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
	userIP, err := getUserIP(c)
	if err != nil {
		log.Println("Error getting user IP:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	timestamp := time.Now().Unix()

	// Get postform data
	listingID := c.PostForm("listing_id")
	userID := c.PostForm("user_id")
	username := c.PostForm("username")
	commentText := c.PostForm("comment_text")

	// Log the request details
	log.Printf("PostListingComment called with listing_id: %s, user_id: %s, username: %s, comment_text: %s\nfrom IP: %s\nat timestamp: %d",
		listingID, userID, username, commentText, userIP, timestamp)

	// Generate a new UUID for the comment using a timestamp-based version (v7) to ensure uniqueness
	commentID, err := uuid.NewV7()
	if err != nil {
		log.Println("Error generating new comment UUID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Create a new comment
	newComment := sqlc.PostCommentParams{
		CommentID:   pgtype.UUID{Bytes: [16]byte(commentID), Valid: true}, // Unique comment ID based on timestamp
		ListingID:   listingID,
		UserIp:      userIP,
		UserID:      userID,
		Username:    username,
		CommentText: commentText,
	}

	// Log the new comment creation
	log.Println("New comment created for listing:", listingID, "by user:", username, "at timestamp:", timestamp)
	log.Println("Comment details:", newComment)
	log.Println("Sanitizing and validating comment parameters...")

	// Perform first round validation on new comment parameters
	if err := server.Validator.Struct(newComment); err != nil {
		log.Println("Failed first round of validation for new comment:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Sanitize the comment parameters to prevent XSS attacks
	newComment = newComment.Sanitize(*server.SantizationPolicy)

	// Perform second round validation on sanitized new comment parameters
	//
	// Ensures that the comment parameters are safe and valid after sanitization
	if err := server.Validator.Struct(newComment); err != nil {
		log.Println("Failed second round of validation for new comment:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Acquire a Postgres connection from the pool
	postgresConnection, err := server.GetPostgresPool().Acquire(context.TODO())
	if err != nil {
		log.Println("Error acquiring Postgres connection:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer postgresConnection.Release()
	postgresQueryClient := sqlc.New(postgresConnection)

	// Insert the new comment into the database
	postCommentRow, err := postgresQueryClient.PostComment(context.TODO(), newComment)
	if err != nil {
		log.Println("Error inserting new comment into database for listing:", listingID, "-", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	/* // Convert the sqlc.PostCommentRow struct to a models.Comment struct
	newCommentFromDB, err := models.GenericRowToComment(postCommentRow)
	if err != nil {
		log.Println("Error converting new comment row to models.Comment struct for listing:", listingID, "-", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log the new comment from the database
	log.Println("New comment from database for listing:", listingID, ":", newCommentFromDB)

	//log.Println("Response comments for listing:", listingID, ":", responseComments)
	c.JSON(http.StatusCreated, newCommentFromDB) */

	/* returnedComment, err := models.CommentRowToComment(postCommentRow)
	if err != nil {
		log.Println("Error converting new comment row to models.Comment struct for listing:", listingID, "-", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	} */

	// Log the successful creation of the new comment
	c.JSON(http.StatusCreated, postCommentRow)
	log.Println("New comment successfully created for listing:", listingID, ":", postCommentRow)
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
	postgresConnection, err := server.GetPostgresPool().Acquire(context.TODO())
	if err != nil {
		log.Println("Error acquiring Postgres connection:", err)
		return nil, errors.Join(err, errors.New("failed to acquire postgres connection"))
	}
	defer postgresConnection.Release()
	postgresQueryClient := sqlc.New(postgresConnection)

	// Query the database for comments by listing ID
	commentRows, err := postgresQueryClient.GetCommentsByListingID(context.TODO(), listingID)
	if err != nil {
		log.Println("Error retrieving comments from database for listing:", listingID, "-", err)
		return nil, errors.Join(err, errors.New("failed to retrieve comments from database"))
	}

	// Convert the sqlc.GetCommentsByListingIDRow structs to models.Comment structs
	comments, err := models.GetCommentRowsToComments(commentRows)
	if err != nil {
		log.Println("Error converting comment rows to models. Comment structs for listing:", listingID, "-", err)
		return nil, errors.Join(err, errors.New("failed to convert comment rows to models.Comment structs"))
	}

	// Return the comments to the client
	/* slices.SortStableFunc(comments, func(a, b models.Comment) int {
		return int(b.Timestamp) - int(a.Timestamp) // Sort by timestamp in descending order
	}) */

	return comments, nil
}

// GenerateUserID generates a new user ID for the client.
//
// GET api/v1/user/user_id
//
// Output:
//   - 200: A JSON object containing the generated user ID. ID is a V7 (Time) UUID.
func (server *Server) GenerateUserID(c *gin.Context) {
	// Get information from the request context
	userIP, err := getUserIP(c)
	if err != nil {
		log.Println("Error getting user IP:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	timestamp := time.Now().Unix()
	log.Println("GenerateUserID called from IP:", userIP, "at timestamp:", timestamp)

	// Generate a new UUID for the user using a timestamp-based version (v7) to ensure uniqueness
	userID, err := uuid.NewV7()
	if err != nil {
		log.Println("Error generating new user UUID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Log the generated user ID
	log.Println("Generated new user ID:", userID)

	// Return the user ID as a JSON response
	c.JSON(http.StatusOK, gin.H{"user_id": userID.String()})
	log.Println("Successfully returned user ID:", userID.String())
}

// getUserIP retrieves the user's IP address from the API Gateway context.
//
// Input:
//   - c: The gin context containing the request.
//
// Output:
//   - A pointer to a string containing the user's IP address.
//   - An error if the API Gateway context does not contain a valid SourceIP.
func getUserIP(c *gin.Context) (string, error) {
	apiGatewayContext, ok := ginadaptercore.GetAPIGatewayContextFromContext(c.Request.Context())
	if !ok {
		return "", errors.New("failed to get API Gateway context from request context")
	}
	if apiGatewayContext.Identity.SourceIP == "" {
		return "", errors.New("API Gateway context does not contain a valid SourceIP")
	}

	userIP := apiGatewayContext.Identity.SourceIP
	return userIP, nil
}

// debugAPIGatewayContext logs the API Gateway context information from the gin context.
//
// This function is useful for debugging purposes to inspect the API Gateway context.
//
// Input:
//   - c: The gin context containing the request.
func debugAPIGatewayContext(c *gin.Context) {
	// Debug gin context params
	for k, v := range c.Params {
		log.Printf("Context key: %v, value: %v\n", k, v)
	}

	// Debug gin context request
	log.Println("Context request method:", c.Request)

	// Debug gin context errors
	for k, v := range c.Errors {
		log.Printf("Context error key: %v, value: %v\n", k, v)
	}

	// the methods are available in your instance of the GinLambda
	// object and receive the context
	apiGwContext, contextOk := ginadaptercore.GetAPIGatewayContextFromContext(c.Request.Context())
	apiGwStageVars, varsOk := ginadaptercore.GetStageVarsFromContext(c.Request.Context())
	runtimeContext, runtimeCtxOk := ginadaptercore.GetRuntimeContextFromContext(c.Request.Context())

	// you can access the properties of the context directly
	log.Println("API GW Context:", apiGwContext, ", Okay: ", contextOk)
	log.Println("API GW Context Request ID:", apiGwContext.RequestID, ", Okay: ", contextOk)
	log.Println("API GW Context Stage:", apiGwContext.Stage, ", Okay: ", contextOk)
	log.Println("API GW User IP:", apiGwContext.Identity.SourceIP)
	log.Println("API GW Context Stage Variables:", apiGwStageVars, ", Okay: ", varsOk)
	if runtimeContext != nil {
		log.Println("Runtime Context Invoked Function ARN: ", runtimeContext.InvokedFunctionArn, ", Okay: ", runtimeCtxOk)
	} else {
		log.Println("Runtime Context is nil, Okay: ", runtimeCtxOk)
	}
}
