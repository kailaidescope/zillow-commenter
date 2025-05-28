// The api package contains the server and its routes for the application.
package api

import (
	"context"
	"net/http"
	"os"

	"zillow-commenter.com/m/api/models"
	"zillow-commenter.com/m/token"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Server struct {
	Router *gin.Engine
	maker  *token.PasetoMaker
	pool   *pgxpool.Pool
}

func (server *Server) GetPostgresPool() *pgxpool.Pool {
	return server.pool
}

func GetNewServer() (*Server, error) {
	//load env vars
	godotenv.Load()
	key := os.Getenv("TOKEN_KEY")
	tokenMaker, err := token.NewPasetoMaker(key)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}

	router := gin.Default()
	// Set up CORS middleware to allow all origins, methods, and headers
	router.Use(cors.Default())

	server := &Server{
		Router: router,
		maker:  tokenMaker,
		pool:   pool,
	}

	// =============================================================================================================== //
	//                                             Mount routes below                                                  //
	// =============================================================================================================== //

	// Top-leve api routes
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

	models.InitTempCommentDB() // Initialize the temporary comment database

	// load router
	// load token maker
	return server, nil
}

func (server *Server) NotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"status": "This resource is not yet implemented, but will be in the future"})
}
