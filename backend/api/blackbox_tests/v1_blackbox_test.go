// Package tests contains blackbox tests for the API.
package blackbox_tests

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
	values.Set("listing_id", "<b>1</b>")
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
	values.Set("listing_id", "1")
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
	values.Set("listing_id", "1")
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
	values.Set("listing_id", "1")
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

/* func TestRemoveLinks(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)
	replacementText := "[link removed]"

	cases := []struct {
		input    string
		expected string
	}{
		{"Check this out: http://example.com", "Check this out: " + replacementText},
		{"Visit https://secure.com for info", "Visit " + replacementText + " for info"},
		{"Go to www.website.org now!", "Go to " + replacementText + " now!"},
		{"No links here", "No links here"},
		{"Multiple links: http://a.com and https://b.com", "Multiple links: " + replacementText + " and " + replacementText},
		{"Text before http://foo.com and after", "Text before " + replacementText + " and after"},
		{"https://abc.com?query=1", replacementText},
		{"www.abc.com/page.html", replacementText},
		{"Mixed: www.abc.com, http://def.com, and text", "Mixed: " + replacementText + ", " + replacementText + ", and text"},
		{"ftp://notalink.com", "ftp://notalink.com"}, // Should not match
		{"http://", replacementText},
		{"www.", "www."},
		{"https://sub.domain.com/path", replacementText},
		{"Check www.site.com and http://site.com", "Check " + replacementText + " and " + replacementText},
		{"Just text", "Just text"},
		{"http://example.com/path?query=1#fragment", replacementText},
		{"www.example.com:8080", replacementText},
		{"http://example.com.", replacementText + "."},
	}

	// Send all test cases to API and check results
	for _, c := range cases {
		v7, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate V7 UUID: %v", err)
		}

		values := url.Values{}
		values.Set("listing_id", "1")
		values.Set("user_id", v7.String())
		values.Set("username", "TestUser")
		values.Set("comment_text", c.input)

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

		// Check if the comment text was sanitized correctly

		// TODO: Unmarshal the response to check the comment text
		if !strings.Contains(resp.String(), c.expected) {
			t.Errorf("removeLinks failed for input '%s': expected '%s', got '%s'", c.input, c.expected, resp.String())
		} else {
			//t.Logf("removeLinks passed for input '%s': expected '%s', got '%s'", c.input, c.expected, resp.String())
		}
	}
}

func TestRemoveEmails(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)
	replacementText := "[email removed]"

	cases := []struct {
		input    string
		expected string
	}{
		{"Contact me at test@example.com", "Contact me at " + replacementText},
		{"Emails: foo@bar.com, bar@foo.org", "Emails: " + replacementText + ", " + replacementText},
		{"No email here", "No email here"},
		{"Edge case: a@b.c", "Edge case: a@b.c"}, // Should not match, as TLD is only 1 char
		{"Send to john.doe@company.co.uk", "Send to " + replacementText},
		{"Multiple: a@b.com b@c.net c@d.org", "Multiple: " + replacementText + " " + replacementText + " " + replacementText},
		{"test@sub.domain.com", replacementText},
		{"user+tag@domain.com", replacementText},
		{"user_name@domain.io", replacementText},
		{"user@domain", "user@domain"},     // Invalid, should not match
		{"user@domain.c", "user@domain.c"}, // TLD too short
		{"user@domain.comm", replacementText},
		{"user@domain.com.", replacementText + "."},
		{"user@domain.com!", replacementText + "!"},
		{"user@domain.com?subject=hi", replacementText + "?subject=hi"},
		{"user@domain.com;user2@domain.com", replacementText + ";" + replacementText},
	}

	// Send all test cases to API and check results
	for _, c := range cases {
		v7, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate V7 UUID: %v", err)
		}

		values := url.Values{}
		values.Set("listing_id", "1")
		values.Set("user_id", v7.String())
		values.Set("username", "TestUser")
		values.Set("comment_text", c.input)

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

		// Check if the comment text was sanitized correctly

		// TODO: Unmarshal the response to check the comment text
		if !strings.Contains(resp.String(), c.expected) {
			t.Errorf("removeLinks failed for input '%s': expected '%s', got '%s'", c.input, c.expected, resp.String())
		} else {
			//t.Logf("removeLinks passed for input '%s': expected '%s', got '%s'", c.input, c.expected, resp.String())
		}
	}
}

func TestRemovePhoneNumbers(t *testing.T) {
	testingSuite, apiIP := SetupAndTeardown(t)
	defer testingSuite(t)
	replacementText := "[phone number removed]"

	cases := []struct {
		input    string
		expected string
	}{
		{"Call me at 555-123-4567", "Call me at " + replacementText},
		{"My number is (555) 123-4567.", "My number is " + replacementText + "."},
		{"+1 555 123 4567 is my office.", replacementText + " is my office."},
		{"No phone here", "No phone here"},
		{"Multiple: 555.123.4567 and 5551234567", "Multiple: " + replacementText + " and " + replacementText},
		{"5551234567", replacementText},
		{"(555)123-4567", replacementText},
		{"555 123 4567", replacementText},
		{"555.123.4567", replacementText},
		{"+44 20 7946 0958", replacementText},
		{"123-4567", "123-4567"}, // Not a full phone number, should not match
		{"555-1234", "555-1234"}, // Not a full phone number, should not match
		{"Phone: 555-123-4567, Alt: (555) 123-4567", "Phone: " + replacementText + ", Alt: " + replacementText},
		{"5551234567 ext. 89", replacementText + " ext. 89"},
		{"Text 555-123-4567 text", "Text " + replacementText + " text"},
		{"(555)1234567", replacementText},
		{"555123-4567", replacementText},
	}

	// Send all test cases to API and check results
	for _, c := range cases {
		v7, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("Failed to generate V7 UUID: %v", err)
		}

		values := url.Values{}
		values.Set("listing_id", "1")
		values.Set("user_id", v7.String())
		values.Set("username", "TestUser")
		values.Set("comment_text", c.input)

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

		// Check if the comment text was sanitized correctly

		// TODO: Unmarshal the response to check the comment text
		if !strings.Contains(resp.String(), c.expected) {
			t.Errorf("removeLinks failed for input '%s': expected '%s', got '%s'", c.input, c.expected, resp.String())
		} else {
			//t.Logf("removeLinks passed for input '%s': expected '%s', got '%s'", c.input, c.expected, resp.String())
		}
	}
} */

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
	values.Set("listing_id", "1")
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
	values.Set("listing_id", "1")
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
	values.Set("listing_id", "1")
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
