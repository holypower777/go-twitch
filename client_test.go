package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("Client must be valid", func(t *testing.T) {
		client, err := NewClient(creds, nil)
		assertNoError(t, err)

		if got, want := client.BaseURL.String(), defaultBaseURL; got != want {
			t.Errorf("wrong base url\ngot: %s\nwant: %s\n", got, want)
		}

		if got, want := client.AuthURL.String(), defaultAuthURL; got != want {
			t.Errorf("wrong auth url\ngot: %s\nwant: %s\n", got, want)
		}

		client2, _ := NewClient(creds, nil)

		if client.HTTPClient == client2.HTTPClient {
			t.Error("NewClient returned same http.Clients, but they should differ")
		}
	})

	t.Run("Bad credentials", func(t *testing.T) {
		_, err := NewClient(&Credentials{}, nil)
		assertErrorPresence(t, err)
		assertErrorMessage(t, err, "ClientId field is required")

		_, err = NewClient(&Credentials{ClientId: "kek"}, nil)
		assertErrorPresence(t, err)
		assertErrorMessage(t, err, "ClientSecret field is required")
	})
}

func TestNewRequest(t *testing.T) {
	t.Run("test url, body, client-id and user-agent treated right", func(t *testing.T) {
		c, _ := NewClient(creds, nil)

		inURL, outURL := "kek", defaultBaseURL+"kek"
		inBody, outBody := struct{ Kek string }{"lol"}, `{"Kek":"lol"}`+"\n"

		req, err := c.NewRequest(http.MethodGet, inURL, inBody)
		assertNoError(t, err)

		if got, want := req.URL.String(), outURL; got != want {
			t.Errorf("bad url\ngot: %s\nwant: %s\n", got, want)
		}

		body, _ := ioutil.ReadAll(req.Body)
		if got, want := string(body), outBody; got != want {
			t.Errorf("bad body\ngot: %s\nwant: %s\n", got, want)
		}

		if got, want := req.Header.Get("User-Agent"), userAgent; got != want {
			t.Errorf("bad user-agent\ngot: %s\nwant: %s\n", got, want)
		}

		if got, want := req.Header.Get("Client-Id"), creds.ClientId; got != want {
			t.Errorf("client-id header is wrong\ngot: %s\nwant: %s", got, want)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		c, _ := NewClient(creds, nil)

		type T struct {
			A map[interface{}]interface{}
		}

		_, err := c.NewRequest(http.MethodGet, ".", &T{})

		assertErrorPresence(t, err)
		if err, ok := err.(*json.UnsupportedTypeError); !ok {
			t.Errorf("expected a JSON error; got %#v.", err)
		}
	})

	t.Run("bad URL", func(t *testing.T) {
		c, _ := NewClient(creds, nil)

		_, err := c.NewRequest(http.MethodGet, ":", nil)
		assertUrlParseError(t, err)
	})

	t.Run("bad method", func(t *testing.T) {
		c, _ := NewClient(creds, nil)

		_, err := c.NewRequest("AMOGUS\n", ".", nil)
		assertErrorPresence(t, err)
	})

	t.Run("empty body must not return an error", func(t *testing.T) {
		c, _ := NewClient(creds, nil)

		req, _ := c.NewRequest(http.MethodGet, ".", nil)

		if req.Body != nil {
			t.Error("builded request contains a non-nil")
		}
	})
}

func TestDo(t *testing.T) {
	t.Run("must send request and write result to provided body", func(t *testing.T) {
		c, mux, _, teardown := setup()
		defer teardown()

		type foo struct {
			A string
		}

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			assertMethod(t, r, http.MethodGet)
			fmt.Fprint(w, `{"A":"a"}`)
		})

		req, _ := c.NewRequest(http.MethodGet, ".", nil)

		body := new(foo)
		ctx := context.Background()
		_, doErr := c.Do(ctx, req, body)
		assertNoError(t, doErr)
		want := &foo{"a"}

		if !reflect.DeepEqual(body, want) {
			t.Errorf("body is not equal\ngot: %v\nwant: %v\n", body, want)
		}
	})

	t.Run("must return error when context is nil", func(t *testing.T) {
		c, _, _, teardown := setup()
		defer teardown()

		req, _ := c.NewRequest(http.MethodGet, ".", nil)
		_, err := c.Do(nil, req, nil)
		assertErrorPresence(t, err)

		if !errors.Is(err, errNonNilContext) {
			t.Errorf("expected context must be non-nil error")
		}
	})

	t.Run("must not return nil, ErrorResponse", func(t *testing.T) {
		c, mux, _, teardown := setup()
		defer teardown()

		mux.HandleFunc("/bad", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		req, _ := c.NewRequest(http.MethodGet, "/bad", nil)
		ctx := context.Background()
		resp, err := c.Do(ctx, req, struct{}{})

		assertErrorPresence(t, err)
		if resp != nil {
			t.Errorf("expected nil as returned response, but got: %v", resp)
		}
	})
}

func TestNewResponse(t *testing.T) {
	c, mux, _, teardown := setup()
	defer teardown()

	var (
		rateLimit     = 10
		rateRemaining = 9
		rateReset     = time.Now()
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/rate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRateLimit, strconv.Itoa(rateLimit))
		w.Header().Set(headerRateRemaining, strconv.Itoa(rateRemaining))
		w.Header().Set(headerRateReset, strconv.Itoa(int(rateReset.Unix())))
		w.WriteHeader(http.StatusCreated)
	})

	t.Run("must return Response type with status 200", func(t *testing.T) {
		req, _ := c.NewRequest(http.MethodGet, "", nil)
		ctx := context.Background()
		resp, err := c.Do(ctx, req, struct{}{})
		assertNoError(t, err)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("response code is not equal\ngot: %d\nwant: %d\n", resp.StatusCode, http.StatusOK)
		}

		want := 0
		if resp.Rate.Limit != want {
			t.Errorf("response rate.limit is not equal 0\ngot: %d\nwant: %d\n", resp.Rate.Limit, want)
		}
	})

	t.Run("Response must has info from Rate headers", func(t *testing.T) {
		req, _ := c.NewRequest(http.MethodGet, "/rate", nil)
		ctx := context.Background()
		resp, err := c.Do(ctx, req, struct{}{})

		assertNoError(t, err)
		if got, want := resp.Rate.Limit, rateLimit; got != want {
			t.Errorf("rate limit is not equal\ngot: %d\nwant: %d\n", got, want)
		}

		if got, want := resp.Rate.Remaining, rateRemaining; got != want {
			t.Errorf("rate remaining is not equal\ngot: %d\nwant: %d\n", got, want)
		}

		if got, want := resp.Rate.Reset.Unix(), rateReset.Unix(); got != want {
			t.Errorf("rate reset is not equal\ngot: %v\nwant: %v\n", got, want)
		}
	})
}
