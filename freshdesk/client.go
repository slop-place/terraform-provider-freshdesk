// Package freshdesk implements a complete Go client for the Freshdesk API v2.
//
// It is a standalone library: it has no dependency on Terraform and may be
// imported directly by any Go program that needs to talk to Freshdesk.
package freshdesk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultUserAgent is sent with every request unless overridden.
	DefaultUserAgent = "terraform-provider-freshdesk"
	// DefaultMaxRetries bounds automatic retries for throttled/transient failures.
	DefaultMaxRetries = 5
	// DefaultTimeout bounds a single HTTP round trip including retries' body reads.
	DefaultTimeout = 60 * time.Second

	// maxBackoff caps the exponential delay between retries.
	maxBackoff = 30 * time.Second
	// maxRetryAfter caps how long a server-supplied Retry-After may hold us.
	maxRetryAfter = 5 * time.Minute
	// backoffBase is the multiplier of the exponential retry delay.
	backoffBase = 2
)

// AttachmentsField is the multipart field name Freshdesk expects for file
// uploads on tickets, notes, replies and canned responses.
const AttachmentsField = "attachments[]"

// Config configures a Client.
type Config struct {
	// Domain is either the Freshdesk subdomain ("acme") or a full host
	// ("acme.freshdesk.com" / "https://acme.freshdesk.com").
	Domain string
	// APIKey is the Freshdesk API key, used as the basic-auth username.
	APIKey string
	// UserAgent overrides DefaultUserAgent when set.
	UserAgent string
	// MaxRetries overrides DefaultMaxRetries. Zero means "use the default";
	// pass a negative value to disable retries entirely.
	MaxRetries int
	// Timeout overrides DefaultTimeout when > 0.
	Timeout time.Duration
	// HTTPClient overrides the default transport (used by tests).
	HTTPClient *http.Client
}

// Client is a Freshdesk API v2 client. It is safe for concurrent use.
type Client struct {
	baseURL    *url.URL
	authHeader string
	userAgent  string
	maxRetries int
	httpClient *http.Client

	// sleep is swappable so tests do not pay real backoff delays.
	sleep func(context.Context, time.Duration) error
}

// NormalizeDomain turns any accepted domain spelling into a canonical host.
func NormalizeDomain(domain string) (string, error) {
	d := strings.TrimSpace(domain)
	if d == "" {
		return "", ErrEmptyDomain
	}
	d = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(d, "https://"), "http://"), "/")
	if d == "" {
		return "", ErrEmptyDomain
	}
	if strings.ContainsAny(d, "/ ") {
		return "", fmt.Errorf("%w: %q", ErrInvalidDomain, domain)
	}
	if !strings.Contains(d, ".") {
		d += ".freshdesk.com"
	}
	return strings.ToLower(d), nil
}

// New builds a Client from cfg.
func New(cfg Config) (*Client, error) {
	host, err := NormalizeDomain(cfg.Domain)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrEmptyAPIKey
	}

	base, err := url.Parse("https://" + host + "/api/v2/")
	if err != nil {
		return nil, fmt.Errorf("building freshdesk base url: %w", err)
	}

	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = DefaultTimeout
		}
		hc = &http.Client{Timeout: timeout}
	}

	ua := cfg.UserAgent
	if ua == "" {
		ua = DefaultUserAgent
	}

	retries := cfg.MaxRetries
	switch {
	case retries == 0:
		retries = DefaultMaxRetries
	case retries < 0:
		// A negative value is the caller opting out of retries.
		retries = 0
	}

	// Freshdesk uses the API key as the basic-auth username with any password.
	token := base64.StdEncoding.EncodeToString([]byte(cfg.APIKey + ":X"))

	return &Client{
		baseURL:    base,
		authHeader: "Basic " + token,
		userAgent:  ua,
		maxRetries: retries,
		httpClient: hc,
		sleep:      sleepCtx,
	}, nil
}

// BaseURL returns the API root the client talks to.
func (c *Client) BaseURL() string { return c.baseURL.String() }

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("waiting to retry freshdesk request: %w", ctx.Err())
	case <-t.C:
		return nil
	}
}

// Request describes a single API call. Callers normally use Get/Post/Put/
// Delete; Request plus Do is the escape hatch for endpoints this library does
// not yet model.
type Request struct {
	// Method is the HTTP verb.
	Method string
	// Path is relative to the API root ("tickets/42"); a leading slash is fine.
	Path string
	// Query holds additional query-string parameters.
	Query url.Values
	// Body is JSON-encoded when set and Raw is nil.
	Body any
	// Raw, when set, is used as the request body verbatim with ContentType.
	Raw []byte
	// ContentType overrides the default of application/json.
	ContentType string
}

// Do executes req, retrying throttled and transient responses, and decodes a
// successful JSON body into out (which may be nil). It returns the response
// headers of the final attempt.
func (c *Client) Do(ctx context.Context, req Request, out any) (http.Header, error) {
	endpoint, err := c.resolve(req.Path, req.Query)
	if err != nil {
		return nil, err
	}

	payload, contentType, err := encodeBody(req)
	if err != nil {
		return nil, err
	}

	for attempt := 0; ; attempt++ {
		hdr, retryAfter, err := c.attempt(ctx, req.Method, endpoint, payload, contentType, out)
		if err == nil {
			return hdr, nil
		}

		// A cancelled context is final, however the attempt failed.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("freshdesk request cancelled: %w", ctxErr)
		}
		if !retryable(err) || attempt >= c.maxRetries {
			return hdr, err
		}
		if serr := c.sleep(ctx, backoff(attempt, retryAfter)); serr != nil {
			return nil, serr
		}
	}
}

// encodeBody renders a request body and its content type.
func encodeBody(req Request) ([]byte, string, error) {
	contentType := req.ContentType

	switch {
	case req.Raw != nil:
		return req.Raw, contentType, nil
	case req.Body != nil:
		payload, err := json.Marshal(req.Body)
		if err != nil {
			return nil, "", fmt.Errorf("encoding request body: %w", err)
		}
		if contentType == "" {
			contentType = "application/json"
		}

		return payload, contentType, nil
	default:
		return nil, contentType, nil
	}
}

// transportError marks a failure to reach Freshdesk at all, as opposed to a
// response the API returned. Transport failures are always worth retrying.
type transportError struct{ err error }

func (e *transportError) Error() string { return "freshdesk request failed: " + e.err.Error() }
func (e *transportError) Unwrap() error { return e.err }

// retryable reports whether err is worth another attempt.
func retryable(err error) bool {
	var transport *transportError
	if errors.As(err, &transport) {
		return true
	}

	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.Retryable()
	}

	// A decode failure means the response arrived but made no sense; sending
	// the same request again will not help.
	return false
}

// backoff returns the delay before the next attempt, honouring Retry-After.
func backoff(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && secs >= 0 {
			return min(time.Duration(secs)*time.Second, maxRetryAfter)
		}
	}
	// Exponential: 1s, 2s, 4s, 8s, 16s, capped at maxBackoff.
	return min(time.Duration(math.Pow(backoffBase, float64(attempt)))*time.Second, maxBackoff)
}

// Get issues a GET and decodes the JSON response into out.
func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodGet, Path: path, Query: query}, out)
	return err
}

// Post issues a POST with a JSON body and decodes the response into out.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPost, Path: path, Body: body}, out)
	return err
}

// Put issues a PUT with a JSON body and decodes the response into out.
func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPut, Path: path, Body: body}, out)
	return err
}

// Patch issues a PATCH with a JSON body and decodes the response into out.
// A few endpoints — skills among them — accept PATCH where the rest of the API
// takes PUT.
func (c *Client) Patch(ctx context.Context, path string, body, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPatch, Path: path, Body: body}, out)

	return err
}

// Delete issues a DELETE.
func (c *Client) Delete(ctx context.Context, path string) error {
	_, err := c.Do(ctx, Request{Method: http.MethodDelete, Path: path}, nil)
	return err
}

// PostMultipart uploads files alongside form fields. Freshdesk requires a
// multipart body whenever attachments are present.
func (c *Client) PostMultipart(
	ctx context.Context,
	path string,
	fields map[string][]string,
	files map[string][]string,
	out any,
) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for name, values := range fields {
		for _, v := range values {
			if err := mw.WriteField(name, v); err != nil {
				return fmt.Errorf("writing multipart field %q: %w", name, err)
			}
		}
	}
	for name, paths := range files {
		for _, p := range paths {
			if err := writeFilePart(mw, name, p); err != nil {
				return err
			}
		}
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("finalizing multipart body: %w", err)
	}

	_, err := c.Do(ctx, Request{
		Method:      http.MethodPost,
		Path:        path,
		Raw:         buf.Bytes(),
		ContentType: mw.FormDataContentType(),
	}, out)
	return err
}

func writeFilePart(mw *multipart.Writer, field, path string) error {
	// The caller chooses which local files to attach, so a variable path here
	// is the feature rather than a risk.
	f, err := os.Open(path) //nolint:gosec // caller-supplied attachment path
	if err != nil {
		return fmt.Errorf("opening attachment %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		escapeQuotes(field), escapeQuotes(filepath.Base(path))))
	h.Set("Content-Type", "application/octet-stream")

	part, err := mw.CreatePart(h)
	if err != nil {
		return fmt.Errorf("creating multipart part for %q: %w", path, err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("copying attachment %q: %w", path, err)
	}
	return nil
}

// attempt performs one HTTP round trip. On a non-2xx response it returns the
// API error along with the response's Retry-After header, if any.
func (c *Client) attempt(
	ctx context.Context,
	method, endpoint string,
	payload []byte,
	contentType string,
	out any,
) (http.Header, string, error) {
	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewReader(payload)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, "", fmt.Errorf("building request: %w", err)
	}
	httpReq.Header.Set("Authorization", c.authHeader)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", &transportError{err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.Header, "", &transportError{err: err}
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		if out != nil && len(bytes.TrimSpace(respBody)) > 0 {
			if err := json.Unmarshal(respBody, out); err != nil {
				return resp.Header, "", fmt.Errorf("decoding freshdesk response (%s %s): %w",
					method, endpoint, err)
			}
		}

		return resp.Header, "", nil
	}

	return resp.Header, resp.Header.Get("Retry-After"), parseError(resp, respBody)
}

// resolve turns a relative API path and query into an absolute URL.
func (c *Client) resolve(path string, query url.Values) (string, error) {
	rel, err := url.Parse(strings.TrimPrefix(path, "/"))
	if err != nil {
		return "", fmt.Errorf("invalid api path %q: %w", path, err)
	}
	u := c.baseURL.ResolveReference(rel)
	if len(query) > 0 {
		existing := u.Query()
		for k, vs := range query {
			for _, v := range vs {
				existing.Add(k, v)
			}
		}
		u.RawQuery = existing.Encode()
	}
	return u.String(), nil
}

// quoteEscaper escapes multipart header values. It is immutable and shared.
//
//nolint:gochecknoglobals // an immutable, concurrency-safe replacer
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string { return quoteEscaper.Replace(s) }
