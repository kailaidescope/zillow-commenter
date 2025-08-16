// Package tests contains blackbox tests for the API.
package blackbox_tests

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"zillow-commenter.com/m/api/models"
)

// ===================================================================================================================== //
//                                              Testing Suite Setup                                                      //
// ===================================================================================================================== //

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
		tb.Log("Connection to server closed")
	}, apiIP
}

func formatResponse(resp *resty.Response) string {
	return resp.Status() + ", " + resp.String()
}

func getTestListingId() string {
	return "0"
}

// ===================================================================================================================== //
//                                               Sanitization Tests                                                      //
// ===================================================================================================================== //

func TestPostComment_ValidateListingID_InvalidID(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	v7, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("Failed to generate V7 UUID: %v", err)
	}

	values := url.Values{}
	values.Set("listing_id", "<b>"+getTestListingId()+"</b>")
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
	values.Set("listing_id", getTestListingId())
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
	values.Set("listing_id", getTestListingId())
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
	values.Set("listing_id", getTestListingId())
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

// Tests for removing links, emails, and phone numbers from comment text

func runSanitizationTestCases(t *testing.T, replacementText string, cases []struct {
	input    string
	expected string
}) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)

	t.Log("Running sanitization test cases for replacement:", replacementText)

	for _, c := range cases {
		// Random sleep to avoid rate limiting
		time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)

		// Create a new comment to send
		v7, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate V7 UUID: %v", err)
		}

		values := url.Values{}
		values.Set("listing_id", getTestListingId())
		values.Set("user_id", v7.String())
		values.Set("username", "TestUser")
		values.Set("comment_text", c.input)

		// Send comment to API
		client := resty.New()
		resp, err := client.R().
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetFormDataFromValues(values).
			Post(apiIP + "/api/v1/comments")

		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode() != 201 {
			t.Errorf("Expected 201 for input '%s', got %d: %s", c.input, resp.StatusCode(), formatResponse(resp))
		}

		// Unmarshal the response

		var responseComment models.ResponseComment
		json.NewDecoder(bytes.NewReader(resp.Body())).Decode(&responseComment)
		if err != nil {
			t.Fatal("Failed to decode response: ", err)
		}

		// Check if the comment text matches the expected sanitized output

		if responseComment.CommentText != c.expected {
			t.Errorf("Sanitization failed for input '%s': expected '%s', got '%s'", c.input, c.expected, responseComment.CommentText)
		}
	}
}

func TestRemoveLinks(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Check this out: http://example.com", "Check this out: [link removed]"},
		{"Visit https://secure.com for info", "Visit [link removed] for info"},
		{"Go to www.website.org now!", "Go to [link removed] now!"},
		{"No links here", "No links here"},
		{"Multiple links: http://a.com and https://b.com", "Multiple links: [link removed] and [link removed]"},
		{"Text before http://foo.com and after", "Text before [link removed] and after"},
		{"https://abc.com?query=1", "[link removed]"},
		{"www.abc.com/page.html", "[link removed]"},
		{"Mixed: www.abc.com, http://def.com, and text", "Mixed: [link removed], [link removed], and text"},
		{"ftp://notalink.com", "ftp://notalink.com"},
		{"http://", "[link removed]"},
		{"www.", "www."},
		{"https://sub.domain.com/path", "[link removed]"},
		{"Check www.site.com and http://site.com", "Check [link removed] and [link removed]"},
		{"Just text", "Just text"},
		{"http://example.com/path?query=1#fragment", "[link removed]"},
		{"www.example.com:8080", "[link removed]"},
		{"http://example.com.", "[link removed]."},
	}
	runSanitizationTestCases(t, "[link removed]", cases)
}

func TestRemoveEmails(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Contact me at test@example.com", "Contact me at [email removed]"},
		{"Emails: foo@bar.com, bar@foo.org", "Emails: [email removed], [email removed]"},
		{"No email here", "No email here"},
		{"Edge case: a@b.c", "Edge case: a@b.c"},
		{"Send to john.doe@company.co.uk", "Send to [email removed]"},
		{"Multiple: a@b.com b@c.net c@d.org", "Multiple: [email removed] [email removed] [email removed]"},
		{"test@sub.domain.com", "[email removed]"},
		{"user+tag@domain.com", "[email removed]"},
		{"user_name@domain.io", "[email removed]"},
		{"user@domain", "user@domain"},
		{"user@domain.c", "user@domain.c"},
		{"user@domain.comm", "[email removed]"},
		{"user@domain.com.", "[email removed]."},
		{"user@domain.com!", "[email removed]!"},
		{"user@domain.com?subject=hi", "[email removed]?subject=hi"},
		{"user@domain.com;user2@domain.com", "[email removed];[email removed]"},
	}
	runSanitizationTestCases(t, "[email removed]", cases)
}

func TestRemovePhoneNumbers(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Call me at 555-123-4567", "Call me at [phone number removed]"},
		{"My number is (555) 123-4567.", "My number is [phone number removed]."},
		{"+1 555 123 4567 is my office.", "[phone number removed] is my office."},
		{"No phone here", "No phone here"},
		{"Multiple: 555.123.4567 and 5551234567", "Multiple: [phone number removed] and [phone number removed]"},
		{"5551234567", "[phone number removed]"},
		{"(555)123-4567", "[phone number removed]"},
		{"555 123 4567", "[phone number removed]"},
		{"555.123.4567", "[phone number removed]"},
		{"+44 20 7946 0958", "[phone number removed]"},
		{"123-4567", "123-4567"},
		{"555-1234", "555-1234"},
		{"Phone: 555-123-4567, Alt: (555) 123-4567", "Phone: [phone number removed], Alt: [phone number removed]"},
		{"5551234567 ext. 89", "[phone number removed] ext. 89"},
		{"Text 555-123-4567 text", "Text [phone number removed] text"},
		{"(555)1234567", "[phone number removed]"},
		{"555123-4567", "[phone number removed]"},
	}
	runSanitizationTestCases(t, "[phone number removed]", cases)
}

// ===================================================================================================================== //
//                                                Validation Tests                                                       //
// ===================================================================================================================== //

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
	values.Set("listing_id", getTestListingId())
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
	values.Set("listing_id", getTestListingId())
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
	values.Set("listing_id", getTestListingId())
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

// ===================================================================================================================== //
//                                                     Helpers                                                           //
// ===================================================================================================================== //

func makeStringOfLength(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "a"
	}
	return s
}
