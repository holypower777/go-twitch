package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
)

const (
	defaultBaseURL          = "https://api.twitch.tv/helix/"
	defaultAuthURL          = "https://id.twitch.tv/oauth2/"
	applicationJSON         = "application/json"
	userAgent               = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.162 Safari/537.36"
	headerRateLimit         = "Ratelimit-Limit"
	headerRateReset         = "Ratelimit-Reset"
	headerRateRemaining     = "Ratelimit-Remaining"
	notSuccessResponse      = "response is not success"
	userIdIsRequired        = "user_id is required"
	userIdLoginIsRequired   = "id or login parameter is required"
	broadcasterIdIsRequired = "broadcaster_id is required"
)

var errNonNilContext = errors.New("context must be non-nil")

func addParams(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

type Client struct {
	credentials *Credentials
	HTTPClient  *http.Client
	BaseURL     *url.URL
	AuthURL     *url.URL
	UserAgent   string

	Streams *StreamsService
	Users   *UsersService

	common service
}

type service struct {
	client *Client
}

type Credentials struct {
	ClientId     string
	ClientSecret string
	OAuthToken   *oauth2.Token
}

type ErrorEmptyCredentials struct {
	Field string
}

func (e *ErrorEmptyCredentials) Error() string {
	return fmt.Sprintf("Message: %s field is required", e.Field)
}

func NewClient(creds *Credentials, httpClient *http.Client) (*Client, error) {
	if creds.ClientId == "" {
		return nil, &ErrorEmptyCredentials{"ClientId"}
	}

	if creds.ClientSecret == "" {
		return nil, &ErrorEmptyCredentials{"ClientSecret"}
	}

	authURL, _ := url.Parse(defaultAuthURL)

	// If OAuthToken is provided, the httpClient will contain
	// provided OAuth token.
	// The token will auto-refresh as necessary.
	// THe token will auto-validate every hour.
	if creds.OAuthToken != nil {
		oauth2Config := &oauth2.Config{
			ClientID:     creds.ClientId,
			ClientSecret: creds.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL: authURL.String(),
			},
		}

		ticker := time.NewTicker(30 * time.Minute)
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()

		httpClient = oauth2Config.Client(context.Background(), creds.OAuthToken)
	}

	// If OAuthToken is not provided, the httpClient will contain
	// provided user access token.
	// The token will auto-refresh as necessary.
	if creds.OAuthToken == nil && httpClient == nil {
		oauth2Config := &clientcredentials.Config{
			ClientID:     creds.ClientId,
			ClientSecret: creds.ClientSecret,
			TokenURL:     twitch.Endpoint.TokenURL,
		}

		httpClient = oauth2Config.Client(context.Background())
	}

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		credentials: creds,
		HTTPClient:  httpClient,
		BaseURL:     baseURL,
		AuthURL:     authURL,
		UserAgent:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.162 Safari/537.36",
	}
	c.common.client = c
	c.Streams = (*StreamsService)(&c.common)
	c.Users = (*UsersService)(&c.common)

	return c, nil
}

func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(path)

	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)

		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)

	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", applicationJSON)
	}

	req.Header.Set("Client-Id", c.credentials.ClientId)
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

type Rate struct {
	Remaining int
	Limit     int
	Reset     time.Time
}

type Response struct {
	*http.Response

	Rate Rate
}

type Pagination struct {
	Cursor string `json:"cursor,omitempty"`
}

type ErrorResponse struct {
	*http.Response

	Message string
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("Method: %v\nURL: %v\nStatus Code: %d\nMessage: %v\nResponse: %v",
		e.Request.Method,
		e.Request.URL,
		e.StatusCode,
		e.Message,
		e.Response,
	)
}

type ErrorInvalidOptions struct {
	Options interface{}
	Message string
}

func (e *ErrorInvalidOptions) Error() string {
	return fmt.Sprintf("Message: %s", e.Message)
}

func NewResponse(r *http.Response) *Response {
	resp := &Response{Response: r}
	resp.parseRate()

	return resp
}

func (r *Response) parseRate() {
	var rate Rate
	if limit := r.Response.Header.Get(headerRateLimit); limit != "" {
		rate.Limit, _ = strconv.Atoi(limit)
	}

	if reset := r.Response.Header.Get(headerRateReset); reset != "" {
		rst, _ := strconv.ParseInt(reset, 10, 64)
		rate.Reset = time.Unix(rst, 0)
	}

	if remaining := r.Response.Header.Get(headerRateRemaining); remaining != "" {
		rate.Remaining, _ = strconv.Atoi(remaining)
	}

	r.Rate = rate
}

func (r *Response) isSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode <= 299
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}

	req = req.WithContext(ctx)

	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	defer resp.Body.Close()

	response := NewResponse(resp)

	if success := response.isSuccess(); !success {
		return nil, &ErrorResponse{resp, notSuccessResponse}
	}

	if v != nil {
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil
		}
		if decErr != nil {
			err = decErr
		}
	}

	return response, err
}
