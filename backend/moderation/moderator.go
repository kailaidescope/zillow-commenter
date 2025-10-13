package moderation

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/genai"
	errorhandling "zillow-commenter.com/m/errorHandling"
)

// Singleton Moderator item which contains references to moderation configs, AI clients, and DBs
type Moderator struct {
	Blacklist        Blacklist
	AIClient         *genai.Client
	DbConnectionPool *pgxpool.Pool
}

// Blacklist of filtered words
type Blacklist struct {
	Level1   []string
	Level2   []string
	Level3   []string
	Variants []string
}

// NewModerator gets a new Moderator singleton instance.
// Along the way it initializes the blacklist from the local config file and creates an AI client.
//
// Input:
//   - aiAPIKey: the API key for Gemini AI
//   - dbConnectionPool: the Postgres DB connection pool
//
// Output:
//   - Moderator: a new Moderator singleton
//   - error: non-nil if any errors occurred during initialization
func NewModerator(moderationFile string, aiAPIKey string, dbConnectionPool *pgxpool.Pool) (*Moderator, error) {

	// Initialize struct
	moderator := &Moderator{
		DbConnectionPool: dbConnectionPool,
	}

	// Get AI Client

	// Set timeout for creating new AI client
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: aiAPIKey,
	})
	if err != nil {
		return nil, errorhandling.ErrorAnd("failed to create AI client", err)
	}

	moderator.AIClient = client

	// Get Blacklist
	blacklist, err := GetBlacklist(moderationFile)
	if err != nil {
		return nil, errorhandling.ErrorAnd("failed to parse blacklist during moderator initialization", err)
	}

	moderator.Blacklist = blacklist

	return moderator, nil
}

// GetBlacklist reads the moderationFile and parses the blacklist JSON.
//
// Input:
//   - moderationFile: a string containing the path to the moderation file
//
// Output:
//   - Blacklist: the initialized blacklist
//   - error: non-nil if a parsing or read error occurred
func GetBlacklist(moderationFile string) (Blacklist, error) {
	// Setup parsing schema
	type blacklistJSON struct {
		Blacklist struct {
			Level1     []string `json:"level1"`
			Level2     []string `json:"level2"`
			Level3     []string `json:"level3"`
			Variations []string `json:"variations"`
		} `json:"blacklist"`
	}

	// Read from file
	var result Blacklist

	file, err := os.Open(moderationFile)
	if err != nil {
		return result, err
	}
	defer file.Close()

	var data blacklistJSON
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return result, err
	}

	// Initialize values
	result.Level1 = data.Blacklist.Level1
	result.Level2 = data.Blacklist.Level2
	result.Level3 = data.Blacklist.Level3
	result.Variants = data.Blacklist.Variations

	return result, nil
}
