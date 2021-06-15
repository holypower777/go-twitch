package bot

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestStreamMarshal(t *testing.T) {
	assertJSONMarshal(t, &Stream{}, "{}")

	s := &Stream{
		UserId:      "1",
		ViewerCount: 10,
		StartedAt:   Timestamp{referenceTime},
		TagIds:      []string{"kek", "lol"},
	}

	want := `{
		"user_id": "1",
		"viewer_count": 10,
		"started_at": ` + referenceTimeStr + `,
		"tag_ids": ["kek", "lol"]
	}`

	assertJSONMarshal(t, s, want)
}

func TestGetStreams(t *testing.T) {
	t.Run("tests parameters and body to be valid", func(t *testing.T) {
		c, mux, _, teardown := setup()
		defer teardown()
		dataCursor := "Mg=="

		mux.HandleFunc("/"+getStreamsPath, func(w http.ResponseWriter, r *http.Request) {
			assertMethod(t, r, http.MethodGet)
			assertQuery(t, r, params{
				"first":   "1",
				"user_id": "115141884",
			})
			fmt.Fprint(w, `{"data":[{"user_id":"115141884","user_name":"GRPZDC","viewer_count":379,"started_at": `+referenceTimeStr+`,"tag_ids":["0569b171-2a2b-476e-a596-5bdfb45a1327"],"is_mature":true}],"pagination":{"cursor": "`+dataCursor+`"}}`)
		})

		ctx := context.Background()
		streamsResp, _, err := c.Streams.GetStreams(ctx, &StreamsOptions{
			First:  1,
			UserId: "115141884",
		})
		assertNoError(t, err)

		want := []*Stream{{
			UserId:      "115141884",
			Username:    "GRPZDC",
			ViewerCount: 379,
			StartedAt:   Timestamp{referenceTime},
			TagIds:      []string{"0569b171-2a2b-476e-a596-5bdfb45a1327"},
			IsMature:    true,
		}}

		if !reflect.DeepEqual(streamsResp.Data, want) {
			t.Errorf("\ngot: %v\nwant: %v", streamsResp.Data, want)
		}

		if got := streamsResp.Pagination.Cursor; got != dataCursor {
			t.Errorf("\ngot: %s\nwant: %s", got, dataCursor)
		}
	})

	t.Run("no query must pass test and paginaiton must be empty", func(t *testing.T) {
		c, mux, _, teardown := setup()
		defer teardown()

		mux.HandleFunc("/"+getStreamsPath, func(w http.ResponseWriter, r *http.Request) {
			assertMethod(t, r, http.MethodGet)
			assertQuery(t, r, params{})
			fmt.Fprint(w, `{"data":[{"user_id":"11"}],"pagination":{}}`)
		})

		ctx := context.Background()
		streamsResp, _, err := c.Streams.GetStreams(ctx, nil)
		assertNoError(t, err)

		want := []*Stream{{UserId: "11"}}

		if !reflect.DeepEqual(streamsResp.Data, want) {
			t.Errorf("\ngot: %v\nwant: %v", streamsResp.Data, want)
		}

		if !reflect.DeepEqual(streamsResp.Pagination, Pagination{}) {
			t.Errorf("\ngot: %v\nwant: %s", streamsResp, "{}")
		}
	})
}

func TestGetFollowedStreams(t *testing.T) {
	t.Run("tests parameters and body to be valid", func(t *testing.T) {
		c, mux, _, teardown := setup()
		defer teardown()

		mux.HandleFunc("/"+getFollowedStreamsPath, func(w http.ResponseWriter, r *http.Request) {
			assertMethod(t, r, http.MethodGet)
			assertRequiredParameters(t, r, params{
				"user_id": "",
			})
			assertQuery(t, r, params{
				"user_id": "12",
			})
			fmt.Fprint(w, `{"data":[{"user_id":"12"}],"pagination":{}}`)
		})

		ctx := context.Background()
		streamsResp, _, err := c.Streams.GetFollowedStreams(ctx, &StreamsOptions{
			UserId: "12",
		})
		assertNoError(t, err)

		want := []*Stream{{UserId: "12"}}

		if !reflect.DeepEqual(streamsResp.Data, want) {
			t.Errorf("\ngot: %v\nwant: %v", streamsResp.Data, want)
		}
	})

	t.Run("must return error, when user_id is not provided", func(t *testing.T) {
		client, _ := NewClient(creds, nil)
		ctx := context.Background()
		_, _, err := client.Streams.GetFollowedStreams(ctx, nil)
		assertErrorPresence(t, err)
		assertErrorMessage(t, err, userIdIsRequired)

		_, _, err = client.Streams.GetFollowedStreams(ctx, &StreamsOptions{
			GameId: "11",
		})
		assertErrorPresence(t, err)
	})
}

func TestGetStreamKey(t *testing.T) {
	t.Run("tests parameters and body to be valid", func(t *testing.T) {
		c, mux, _, teardown := setup()
		defer teardown()

		prms := params{"broadcaster_id": "12"}
		mux.HandleFunc("/"+getStreamKeyPath, func(w http.ResponseWriter, r *http.Request) {
			assertMethod(t, r, http.MethodGet)
			assertRequiredParameters(t, r, prms)
			assertQuery(t, r, prms)
			fmt.Fprint(w, `{"data":[{"stream_key":"live_44322889_a34ub37c8ajv98a0"}]}`)
		})

		ctx := context.Background()
		key, _, err := c.Streams.GetStreamKey(ctx, &BroadcasterID{"12"})
		assertNoError(t, err)

		var want StreamKey = "live_44322889_a34ub37c8ajv98a0"

		if key != want {
			t.Errorf("wrong key\ngot: %v\nwant: %v", key, want)
		}
	})

	t.Run("must return error, when broadcaster_id is not provided", func(t *testing.T) {
		client, _ := NewClient(creds, nil)
		ctx := context.Background()
		_, _, err := client.Streams.GetStreamKey(ctx, nil)
		assertErrorPresence(t, err)
		assertErrorMessage(t, err, broadcasterIdIsRequired)
	})
}

// func TestGetStreamMarkers(t *testing.T) {
// 	c, mux, _, teardown := setup()
// 	defer teardown()

// 	mux.HandleFunc("/"+getStreamMarkersPath, func(w http.ResponseWriter, r *http.Request) {
// 		assertMethod(t, r, http.MethodGet)
// 		assertRequiredParameters(t, r, params{"user_id": ""})
// 		assertQuery(t, r, params{"user_id": "123"})
// 		fmt.Fprint(w, `{"data":[],"pagination":{"cursor":""}}`)
// 	})
// }
