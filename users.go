package bot

import (
	"context"
	"net/http"
)

const (
	getUsersPath       = "users"
	users100LimitError = "The limit of 100 IDs and login names is the total limit. You can request, for example, 50 of each or 100 of one of them. You cannot request 100 of both."
)

type UsersService service

type UsersOptions struct {
	Ids    []string `url:"id,omitempty"`
	Logins []string `url:"id,omitempty"`
}

type User struct {
	BroadcasterType string    `json:"broadcaster_type,omitempty"`
	Description     string    `json:"description,omitempty"`
	DisplayName     string    `json:"display_name,omitempty"`
	Id              string    `json:"id,omitempty"`
	Login           string    `json:"login,omitempty"`
	OfflineImageURL string    `json:"offline_image_url,omitempty"`
	ProfileImageURL string    `json:"profile_image_url,omitempty"`
	Type            string    `json:"type,omitempty"`
	ViewCount       int       `json:"view_count,omitempty"`
	Email           string    `json:"email,omitempty"`
	CreatedAt       Timestamp `json:"created_at,omitempty"`
}

type UsersResponse struct {
	Data []*User `json:"data,omitempty"`
}

func (s *UsersService) GetUsers(ctx context.Context, opts *UsersOptions) ([]*User, *Response, error) {
	if opts == nil || opts.Ids == nil && opts.Logins == nil {
		return nil, nil, &ErrorInvalidOptions{
			Options: opts,
			Message: userIdLoginIsRequired,
		}
	}

	if len(opts.Ids)+len(opts.Logins) > 100 {
		return nil, nil, &ErrorInvalidOptions{
			Options: opts,
			Message: users100LimitError,
		}
	}

	u, err := addParams(getUsersPath, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	usersResp := new(UsersResponse)
	resp, err := s.client.Do(ctx, req, usersResp)
	if err != nil {
		return nil, resp, err
	}

	return usersResp.Data, resp, nil
}
