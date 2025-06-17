// The api package contains the server and its routes for the application.
package api

import (
	"context"
	"net/http"
	"os"

	"zillow-commenter.com/m/db/postgres/sqlc"
	"zillow-commenter.com/m/token"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
)

type Server struct {
	Router            *gin.Engine
	LambdaAdapter     *ginadapter.GinLambda
	Validator         *validator.Validate
	SantizationPolicy *bluemonday.Policy
	maker             *token.PasetoMaker
	pool              *pgxpool.Pool
}

func (server *Server) GetPostgresPool() *pgxpool.Pool {
	return server.pool
}

func GetNewServer() (*Server, error) {
	//load env vars
	godotenv.Load()

	// TOKEN MAKER

	key := os.Getenv("TOKEN_KEY")
	tokenMaker, err := token.NewPasetoMaker(key)
	if err != nil {
		return nil, err
	}

	// POSTGRES CONNECTION

	pool, err := pgxpool.New(context.Background(), os.Getenv("CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}

	// ROUTER

	// Create a new Gin router
	router := gin.Default()
	// Set up CORS middleware to allow all origins, methods, and headers
	router.Use(cors.Default())

	// VALIDATOR

	// Set up the validator with required struct validation enabled
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Register custom validation for structs
	validate.RegisterStructValidation(sqlc.PostCommentParamsValidation, sqlc.PostCommentParams{})

	// SANITIZER

	// Initialize bluemonday sanitization policy
	//
	// We use the strict policy because there should be no reason to include *ANY* HTML in our comments
	sanitizationPolicy := bluemonday.StrictPolicy()

	// Collect server singleton variables
	server := &Server{
		Router:            router,
		Validator:         validate,
		SantizationPolicy: sanitizationPolicy,
		maker:             tokenMaker,
		pool:              pool,
	}

	// =============================================================================================================== //
	//                                             Mount routes below                                                  //
	// =============================================================================================================== //

	// Top-level api routes
	api := router.Group("/api")
	{
		// Gives information about the API in general, particularly about how to switch between versions
		api.GET("", server.NotImplemented)

		// Version 1 of the API routes
		api_v1 := api.Group("/v1")
		{
			// Gives information about the first version of the API
			api_v1.GET("", server.NotImplemented)

			// Comment routes
			comments := api_v1.Group("/comments")
			{
				// Gets all comments for a specific zillow listing
				comments.GET(":listing_id", server.GetListingComments)

				// Creates a new comment for a specific zillow listing
				comments.POST("", server.PostListingComment)
			}

			// User routes
			user := api_v1.Group("/user")
			{
				user.GET("/user_id", server.GenerateUserID)
			}
		}
	}

	// =============================================================================================================== //
	//                                             End of mounting routes                                              //
	// =============================================================================================================== //

	//models.InitTempCommentDB() // Initialize the temporary comment database

	server.LambdaAdapter = ginadapter.New(router)

	// load router
	// load token maker
	return server, nil
}

func (server *Server) NotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "This resource is not yet implemented, but will be in the future"})
}
