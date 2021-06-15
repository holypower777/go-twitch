package bot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

const (
	baseURLPath = "/helix"
)

var creds = &Credentials{
	ClientId:     "ClientId",
	ClientSecret: "ClientSecret",
}

var httpClient = &http.Client{}

type params map[string]string

func assertJSONMarshal(t *testing.T, v interface{}, want string) {
	t.Helper()
	// Unmarshal the wanted JSON, to verify its correctness, and marshal it back
	// to sort the keys.
	u := reflect.New(reflect.TypeOf(v)).Interface()
	if err := json.Unmarshal([]byte(want), &u); err != nil {
		t.Errorf("Unable to unmarshal JSON for %v: %v", want, err)
	}
	w, err := json.Marshal(u)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %#v", u)
	}

	// Marshal the target value.
	j, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %#v", v)
	}

	if string(w) != string(j) {
		t.Errorf("json.Marshal(%q) returned %s, want %s", v, j, w)
	}
}

func assertErrorMessage(t testing.TB, err error, msg string) {
	t.Helper()

	if got, want := err.Error(), "Message: "+msg; got != want {
		t.Errorf("error message is wrong\ngot: %s\nwant: %s", got, want)
	}
}

func assertQuery(t testing.TB, r *http.Request, query params) {
	t.Helper()

	want := url.Values{}
	for k, v := range query {
		want.Set(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("request parameters are not equal\ngot: %v\nwant: %v", got, want)
	}
}

func assertRequiredParameters(t testing.TB, r *http.Request, query params) {
	t.Helper()

	r.ParseForm()

	for k := range query {
		if r.Form.Get(k) == "" {
			t.Fatalf("parameter %s is requiered", k)
		}
	}
}

func assertMethod(t testing.TB, r *http.Request, want string) {
	t.Helper()

	if got := r.Method; got != want {
		t.Errorf("bad method\ngot: %s\nwant: %s\n", got, want)
	}
}

func assertUrlParseError(t testing.TB, err error) {
	t.Helper()

	assertErrorPresence(t, err)
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("expected url parse error, got %+v", err)
	}
}

func assertErrorPresence(t testing.TB, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error to be returned")
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("doesn't expect error there: %v", err)
	}
}

func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	client, _ = NewClient(creds, httpClient)
	url, _ := url.Parse(server.URL + baseURLPath)
	client.BaseURL = url

	return client, mux, server.URL, server.Close
}
