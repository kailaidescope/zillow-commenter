package sqlc

import (
	"testing"

	"github.com/microcosm-cc/bluemonday"
)

// ===================================================================================================================== //
//                                             Setup and Teardown                                                        //
// ===================================================================================================================== //

// SanitizationSetupAndTeardown initializes the sanitizer.
//
// IMPORTANT: This function should be called in each test case to ensure the sanitizer is set up correctly.
//
// Input:
//   - tb: A testing.TB interface that allows the function to log messages and handle test failures.
//
// Output:
//   - A function that can be deferred to perform teardown actions after the test completes.
//   - A pointer to a bluemonday.Policy instance that can be used to sanitize strings.
func SanitizationSetupAndTeardown(tb testing.TB) (func(tb testing.TB), *bluemonday.Policy) {
	// Create a sanitization policy
	sanitizationPolicy := bluemonday.StrictPolicy()

	return func(tb testing.TB) {
		tb.Log("Teardown complete")
	}, sanitizationPolicy
}

// ===================================================================================================================== //
//                                                Write tests below                                                      //
// ===================================================================================================================== //

//

// ===================================================================================================================== //
//                                               Sanitization Tests                                                      //
// ===================================================================================================================== //

func TestSanitize_ListingID(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingID = "<b>123456</b>"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.ListingID != "123456" {
		t.Errorf("Expected sanitized ListingID to be '123456', got '%s'", sanitized.ListingID)
	}
}

func TestSanitize_UserIp(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserIp = "<script>alert('x')</script>192.168.1.1"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserIp != "192.168.1.1" {
		t.Errorf("Expected sanitized UserIp to be '192.168.1.1', got '%s'", sanitized.UserIp)
	}
}

func TestSanitize_UserID(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserID = "<i>" + params.UserID + "</i>"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserID != params.UserID[3:len(params.UserID)-4] {
		t.Errorf("Expected sanitized UserID to be '%s', got '%s'", params.UserID[3:len(params.UserID)-4], sanitized.UserID)
	}
}

func TestSanitize_Username(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.Username = "<b>TestUser</b>"
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.Username != "TestUser" {
		t.Errorf("Expected sanitized Username to be 'TestUser', got '%s'", sanitized.Username)
	}
}

func TestSanitize_CommentText(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.CommentText = "<script>alert('xss')</script>This is a comment."
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.CommentText != "This is a comment." {
		t.Errorf("Expected sanitized CommentText to be 'This is a comment.', got '%s'", sanitized.CommentText)
	}
}

func TestSanitize_ListingID_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.ListingID = `<img src="x" onerror="alert('XSS')">123456<script>alert(1)</script>`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.ListingID != "123456" {
		t.Errorf("Expected sanitized ListingID to be '123456', got '%s'", sanitized.ListingID)
	}
}

func TestSanitize_UserIp_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.UserIp = `192.168.1.1"><svg/onload=alert(2)>`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserIp != `192.168.1.1&#34;&gt;` {
		t.Errorf("Expected sanitized UserIp to be '192.168.1.1\"', got '%s'", sanitized.UserIp)
	}
}

func TestSanitize_UserID_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	expectedUserID := params.UserID // Store the original UserID
	params.UserID = `<iframe src="javascript:alert('XSS')"></iframe>` + params.UserID + `<script>alert(3)</script>`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.UserID != expectedUserID {
		t.Errorf("Expected sanitized UserID to be '%s', got '%s'", params.UserID, sanitized.UserID)
	}
}

func TestSanitize_Username_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.Username = `"><img src=x onerror=alert(4)>TestUser<script>alert(5)</script>`
	sanitized := params.Sanitize(*sanitizer)
	// Should return the username with the characters encoded as HTML character entities
	if sanitized.Username != `&#34;&gt;TestUser` {
		t.Errorf("Expected sanitized Username to be '\">TestUser', got '%s'", sanitized.Username)
	}
}

func TestSanitize_CommentText_XSS(t *testing.T) {
	_, sanitizer := SanitizationSetupAndTeardown(t)
	params := GetValidPostCommentParams(ValidParamsIPv4)
	params.CommentText = `<script>alert('xss')</script>This is a comment.<img src="x" onerror="alert('XSS')">`
	sanitized := params.Sanitize(*sanitizer)
	if sanitized.CommentText != "This is a comment." {
		t.Errorf("Expected sanitized CommentText to be 'This is a comment.', got '%s'", sanitized.CommentText)
	}
}

// ===================================================================================================================== //
//                                         Unit Tests for String Sanitizers                                             //
// ===================================================================================================================== //

func TestRemoveLinks(t *testing.T) {
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

	for _, c := range cases {
		result := removeLinks(c.input)
		if result != c.expected {
			t.Error("removeLinks failed:", "input='"+c.input+"'", "expected='"+c.expected+"'", "got='"+result+"'")
		} else {
			//t.Logf("removeLinks passed: input='%s', expected='%s', got='%s'", c.input, c.expected, result)
		}
	}
}

func TestRemoveEmails(t *testing.T) {
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

	for _, c := range cases {
		result := removeEmails(c.input)
		if result != c.expected {
			t.Error("removeEmails failed:", "input='"+c.input+"'", "expected='"+c.expected+"'", "got='"+result+"'")
		} else {
			//t.Logf("removeEmails passed: input='%s', expected='%s', got='%s'", c.input, c.expected, result)
		}
	}
}

func TestRemovePhoneNumbers(t *testing.T) {
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

	for _, c := range cases {
		result := removePhoneNumbers(c.input)
		if result != c.expected {
			t.Error("removePhoneNumbers failed:", "input='"+c.input+"'", "expected='"+c.expected+"'", "got='"+result+"'")
		} else {
			//t.Logf("removePhoneNumbers passed: input='%s', expected='%s', got='%s'", c.input, c.expected, result)
		}
	}
}
