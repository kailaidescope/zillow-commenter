// The api package contains the server and its routes for the application.
package api

// =============================================================================================================== //
//                                            Dependency Overview                                                  //
// =============================================================================================================== //

// GIN

// [Documentation](https://github.com/gin-gonic/gin)

// Used to create a simple web API to be hosted on AWS.

//

// AWS-LAMBDA-GO-API-PROXY

// [Documentation](https://github.com/awslabs/aws-lambda-go-api-proxy)

// Used to interpret AWS Gateway and Lambda events in the backend API.

// [ginadapter gin documentation](https://github.com/awslabs/aws-lambda-go-api-proxy?tab=readme-ov-file#api-gateway-context-and-stage-variables)
// [ginadapter core documentation](https://pkg.go.dev/github.com/awslabs/aws-lambda-go-api-proxy@v0.16.2/core#RequestAccessor)

// Used to get client info from AWS Gateway, such as the client's IP address.

//

import (
	"context"
	"crypto/cipher"
	"errors"
	"net/http"
	"os"

	"zillow-commenter.com/m/db/postgres/sqlc"
	"zillow-commenter.com/m/encryption"
	"zillow-commenter.com/m/token"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/microcosm-cc/bluemonday"
)

type Server struct {
	Router            *gin.Engine
	LambdaAdapter     *ginadapter.GinLambda
	Validator         *validator.Validate
	SantizationPolicy *bluemonday.Policy
	maker             *token.PasetoMaker
	pool              *pgxpool.Pool
	optionsMode       ServerOptions
	AesCipherGCM      cipher.AEAD
}

func (server *Server) GetPostgresPool() *pgxpool.Pool {
	return server.pool
}

// ServerOptions defines the allowed database connection options for the server.
type ServerOptions string

const (
	Production ServerOptions = "production"
	Test       ServerOptions = "test"
)

// GetNewServer creates a new Server instance with all necessary dependencies initialized.
//
// Input:
//   - dbOptions: A enum containing database connection options. Allowed values are ["production", "test"]
func GetNewServer(serverOptions ServerOptions) (*Server, error) {
	// Load env vars. NOT NECESSARY FOR AWS LAMBDA. Also does not work in Docker environment *unless* env_file option is used. AWS Fargate supports
	// direct environment variable injection for secrets, so use that instead.

	/* err := godotenv.Load()
	if err != nil {
		return nil, errors.Join(errors.New("could not load environment variables"), err)
	} */

	// TOKEN MAKER

	key := os.Getenv("TOKEN_KEY")
	tokenMaker, err := token.NewPasetoMaker(key)
	if err != nil {
		return nil, errors.Join(errors.New("could not create token maker"), err)
	}

	// POSTGRES CONNECTION

	// Deny invalid serverOptions
	if serverOptions != Production && serverOptions != Test {
		return nil, errors.New("invalid serverOptions provided, must be either 'production' or 'test'")
	}

	// Create a new connection pool to the PostgreSQL database based on the serverOptions
	var pool *pgxpool.Pool
	switch serverOptions {
	case Test:
		pool, err = pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONNECTION_STRING_TEST"))
		if err != nil {
			return nil, errors.Join(errors.New("could not connect to the test database"), err)
		}
	case Production:
		pool, err = pgxpool.New(context.Background(), os.Getenv("CONNECTION_STRING"))
		if err != nil {
			return nil, errors.Join(errors.New("could not connect to the production database"), err)
		}
	default:
		return nil, errors.New("cannot start server with option that is not 'Production' or 'Test'")
	}

	// ROUTER

	// Create a new Gin router
	router := gin.Default()
	// Set up CORS middleware to allow all origins, methods, and headers
	router.Use(cors.Default())

	// VALIDATOR

	// Set up the validator with required struct validation enabled
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Register custom validations for structs and fields
	sqlc.RegisterValidators(validate)

	// SANITIZER

	// Initialize bluemonday sanitization policy
	//
	// We use the strict policy because there should be no reason to include *ANY* HTML in our comments
	sanitizationPolicy := bluemonday.StrictPolicy()

	// AES CIPHER

	// If key is not 32 bytes long, it is invalid as an AES-256 key
	aesKeyString := os.Getenv("ZILLOWETTE_AES_KEY")

	// Generate new AES cipher object, wrapped in a Galois Counter Mode
	aesCipher, err := encryption.GetNewAESCipher(aesKeyString)
	if err != nil {
		return nil, errors.Join(errors.New("failed to generate AES cipher for server"), err)
	}

	// Collect server singleton variables
	server := &Server{
		Router:            router,
		Validator:         validate,
		SantizationPolicy: sanitizationPolicy,
		maker:             tokenMaker,
		pool:              pool,
		optionsMode:       serverOptions,
		AesCipherGCM:      aesCipher,
	}

	// PLAYWRIGHT (DOES NOT WORK ON AWS LAMBDA)

	/* err = playwright.Install()
	if err != nil {
		return nil, errors.Join(errors.New("could not install playwright"), err)
	} */

	// =============================================================================================================== //
	//                                             Mount routes below                                                  //
	// =============================================================================================================== //

	// Root-level endpoint, to pass AWS health check with Fargate container
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "api is active"})
	})

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
