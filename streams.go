package bot

import (
	"context"
	"net/http"
)

const (
	getStreamsPath         = "streams"
	getFollowedStreamsPath = "streams/followed"
	getStreamKeyPath       = "streams/key"
	getStreamMarkersPath   = "stream/markers"
)

type StreamsService service

type StreamsOptions struct {
	After     string `url:"after,omitempty"`
	Before    string `url:"before,omitempty"`
	First     int    `url:"first,omitempty"`
	GameId    string `url:"game_id,omitempty"`
	Language  string `url:"language,omitempty"`
	UserId    string `url:"user_id,omitempty"`
	UserLogin string `url:"user_login,omitempty"`
}

type Stream struct {
	Id          string    `json:"id,omitempty"`
	UserId      string    `json:"user_id,omitempty"`
	UserLogin   string    `json:"user_login,omitempty"`
	Username    string    `json:"user_name,omitempty"`
	GameId      string    `json:"game_id,omitempty"`
	GameName    string    `json:"game_name,omitempty"`
	Type        string    `json:"type,omitempty"`
	Title       string    `json:"title,omitempty"`
	ViewerCount int       `json:"viewer_count,omitempty"`
	StartedAt   Timestamp `json:"started_at,omitempty"`
	Language    string    `json:"language,omitempty"`
	ThumnailURL string    `json:"thumbnail_url,omitempty"`
	TagIds      []string  `json:"tag_ids,omitempty"`
	IsMature    bool      `json:"is_mature,omitempty"`
}

type StreamsResponse struct {
	Data       []*Stream `json:"data,omitempty"`
	Pagination `json:"pagination,omitempty"`
}

func (s *StreamsService) GetStreams(ctx context.Context, opts *StreamsOptions) (*StreamsResponse, *Response, error) {
	u, err := addParams(getStreamsPath, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	streams := new(StreamsResponse)
	resp, err := s.client.Do(ctx, req, streams)
	if err != nil {
		return nil, resp, err
	}

	return streams, resp, nil
}

func (s *StreamsService) GetFollowedStreams(ctx context.Context, opts *StreamsOptions) (*StreamsResponse, *Response, error) {
	if opts == nil || opts.UserId == "" {
		return nil, nil, &ErrorInvalidOptions{Options: opts, Message: userIdIsRequired}
	}

	u, err := addParams(getFollowedStreamsPath, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	streams := new(StreamsResponse)
	resp, err := s.client.Do(ctx, req, streams)
	if err != nil {
		return nil, resp, err
	}

	return streams, resp, nil
}

type BroadcasterID struct {
	Id string `url:"broadcaster_id,omitempty"`
}

type StreamKeyResponse struct {
	Data []*struct {
		Key StreamKey `json:"stream_key,omitempty"`
	} `json:"data,omitempty"`
}

type StreamKey string

func (s *StreamsService) GetStreamKey(ctx context.Context, opts *BroadcasterID) (StreamKey, *Response, error) {
	if opts == nil || opts.Id == "" {
		return "", nil, &ErrorInvalidOptions{Options: opts, Message: broadcasterIdIsRequired}
	}

	u, err := addParams(getStreamKeyPath, opts)
	if err != nil {
		return "", nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", nil, err
	}

	keyResp := new(StreamKeyResponse)
	resp, err := s.client.Do(ctx, req, keyResp)
	if err != nil {
		return "", resp, err
	}

	return keyResp.Data[0].Key, resp, nil
}
