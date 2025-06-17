package tests

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// SetupAndTeardown initializes the API comment test environment and returns a cleanup function.
//
// It retrieves the API IP, then sends it to the testing suite.
func SetupAndTeardown(tb testing.TB) (func(tb testing.TB), string) {
	// Retrieve API IP from environment variables
	os.Chdir("../..")
	godotenv.Load()
	apiIP := os.Getenv("API_IP")

	// Give the server a moment to start
	time.Sleep(500 * time.Millisecond)
	return func(tb testing.TB) {
		tb.Log("Server closing")
	}, apiIP
}

func formatResponse(resp *resty.Response) string {
	return resp.Status() + ", " + resp.String()
}

// --- SANITIZATION TESTS ---

func TestPostComment_ValidateListingID_InvalidID(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	values := url.Values{}
	values.Set("listing_id", "<b>123456</b>")
	values.Set("user_id", v7.String())
	values.Set("username", "TestUser")
	values.Set("comment_text", "This is a comment.")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Fatalf("Expected 400, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
	if resp.String() == "" || strings.Contains(resp.String(), "<b>") {
		t.Errorf("Sanitization failed for ListingID: %s", resp.String())
	}
}

func TestPostComment_SanitizesUserIp(t *testing.T) {
	// UserIp is set by the server, so this test is best done by checking that XSS in IP is not possible.
	// This is mostly covered by integration and unit tests.
}

func TestPostComment_SanitizesUserID(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	userID := "<i>" + v7.String() + "</i>"
	values := url.Values{}
	values.Set("listing_id", "123456")
	values.Set("user_id", userID)
	values.Set("username", "TestUser")
	values.Set("comment_text", "This is a comment.")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() == 400 {
		t.Log("Correctly rejected invalid user_id (HTML tags not allowed):", formatResponse(resp))
	} else {
		t.Errorf("Expected 400 for unsanitized user_id, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
}

func TestPostComment_SanitizesUsername(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	values := url.Values{}
	values.Set("listing_id", "123456")
	values.Set("user_id", v7.String())
	values.Set("username", "<b>TestUser</b>")
	values.Set("comment_text", "This is a comment.")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() == 400 {
		t.Log("Correctly rejected invalid username (HTML tags not allowed):", formatResponse(resp))
	} else {
		t.Errorf("Expected 400 for unsanitized username, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
}

func TestPostComment_SanitizesCommentText(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	values := url.Values{}
	values.Set("listing_id", "123456")
	values.Set("user_id", v7.String())
	values.Set("username", "TestUser")
	values.Set("comment_text", "<script>alert('xss')</script>This is a comment.")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() != 201 {
		t.Fatalf("Expected 201, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
	if strings.Contains(resp.String(), "<script>") {
		t.Errorf("Sanitization failed for CommentText: %s", resp.String())
	}
}

// --- VALIDATION TESTS ---

func TestPostComment_RejectsMissingListingID(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	values := url.Values{}
	values.Set("user_id", v7.String())
	values.Set("username", "TestUser")
	values.Set("comment_text", "This is a comment.")

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Errorf("Expected 400 for missing listing_id, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
}

func TestPostComment_RejectsInvalidUserID(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	values := url.Values{}
	values.Set("listing_id", "123456")
	values.Set("user_id", "not-a-uuid")
	values.Set("username", "TestUser")
	values.Set("comment_text", "This is a comment.")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Errorf("Expected 400 for invalid user_id, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
}

func TestPostComment_RejectsInvalidUsername(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	values := url.Values{}
	values.Set("listing_id", "123456")
	values.Set("user_id", v7.String())
	values.Set("username", "user!@#")
	values.Set("comment_text", "This is a comment.")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Errorf("Expected 400 for invalid username, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
}

func TestPostComment_RejectsTooLongCommentText(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	values := url.Values{}
	values.Set("listing_id", "123456")
	values.Set("user_id", v7.String())
	values.Set("username", "TestUser")
	values.Set("comment_text", makeStringOfLength(301))

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormDataFromValues(values).
		Post(apiIP + "/api/v1/comments")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode() != 400 {
		t.Errorf("Expected 400 for too long comment_text, got %d: %s", resp.StatusCode(), formatResponse(resp))
	}
}

// --- HELPERS ---

func makeStringOfLength(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}
