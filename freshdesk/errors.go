package freshdesk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors returned for configuration and protocol faults, so callers
// can match them with errors.Is.
var (
	// ErrEmptyDomain is returned when no Freshdesk domain was supplied.
	ErrEmptyDomain = errors.New("freshdesk domain is empty")
	// ErrInvalidDomain is returned for a domain that is not a bare host.
	ErrInvalidDomain = errors.New("invalid freshdesk domain")
	// ErrEmptyAPIKey is returned when no API key was supplied.
	ErrEmptyAPIKey = errors.New("freshdesk api_key is empty")
	// ErrPaginationLimit is returned when a collection exceeds maxListPages.
	ErrPaginationLimit = errors.New("freshdesk pagination limit exceeded")
)

// maxErrorBodyLength caps how much of an unparseable response body is kept in
// an Error's message.
const maxErrorBodyLength = 512

// FieldError is a single per-field validation failure returned by Freshdesk.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (f FieldError) String() string {
	switch {
	case f.Field != "" && f.Code != "":
		return fmt.Sprintf("%s: %s (%s)", f.Field, f.Message, f.Code)
	case f.Field != "":
		return fmt.Sprintf("%s: %s", f.Field, f.Message)
	default:
		return f.Message
	}
}

// Error is a non-2xx response from the Freshdesk API.
type Error struct {
	// StatusCode is the HTTP status of the response.
	StatusCode int
	// Code is Freshdesk's machine-readable error code, when present.
	Code string
	// Message is Freshdesk's human-readable description.
	Message string
	// Errors holds per-field validation failures, when present.
	Errors []FieldError
	// Body is the raw response body, for diagnostics.
	Body string
	// RequestID is Freshdesk's request identifier, when present.
	RequestID string
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "freshdesk api error: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
	if e.Code != "" {
		fmt.Fprintf(&b, " (%s)", e.Code)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	for _, fe := range e.Errors {
		fmt.Fprintf(&b, "\n  - %s", fe.String())
	}
	if e.RequestID != "" {
		fmt.Fprintf(&b, "\n  request-id: %s", e.RequestID)
	}
	return b.String()
}

// Retryable reports whether the request may succeed if sent again unchanged.
func (e *Error) Retryable() bool {
	switch e.StatusCode {
	case http.StatusTooManyRequests, // 429 - rate limited
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// NotFound reports whether err is a 404 from Freshdesk. Resource Read
// implementations use it to detect out-of-band deletion.
func NotFound(err error) bool {
	var apiErr *Error
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// Unauthorized reports whether err is a 401/403 from Freshdesk.
func Unauthorized(err error) bool {
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden
}

// AsError extracts the *Error from err, if there is one.
func AsError(err error) (*Error, bool) {
	var apiErr *Error
	ok := errors.As(err, &apiErr)
	return apiErr, ok
}

// parseError builds an *Error from a non-2xx response.
func parseError(resp *http.Response, body []byte) *Error {
	e := &Error{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		RequestID:  resp.Header.Get("X-Request-Id"),
	}

	var payload struct {
		Code        string       `json:"code"`
		Message     string       `json:"message"`
		Description string       `json:"description"`
		Errors      []FieldError `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		e.Code = payload.Code
		e.Message = payload.Message
		if e.Message == "" {
			e.Message = payload.Description
		}
		e.Errors = payload.Errors
	}

	if e.Message == "" {
		// Fall back to a trimmed raw body so the error is never empty.
		msg := strings.TrimSpace(string(body))
		if len(msg) > maxErrorBodyLength {
			msg = msg[:maxErrorBodyLength] + "..."
		}
		e.Message = msg
	}
	return e
}
