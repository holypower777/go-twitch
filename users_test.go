package bot

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetUsers(t *testing.T) {
	t.Run("tests parameters and body to be valid", func(t *testing.T) {
		c, mux, _, teardown := setup()
		defer teardown()

		mux.HandleFunc("/"+getUsersPath, func(w http.ResponseWriter, r *http.Request) {
			assertMethod(t, r, http.MethodGet)
			assertQuery(t, r, params{"id": "12"})
			fmt.Fprint(w, `{"data":[{"id":"12","display_name":"aboba"}]}`)
		})

		ctx := context.Background()
		users, _, err := c.Users.GetUsers(ctx, &UsersOptions{Ids: []string{"12"}})
		assertNoError(t, err)

		want := []*User{{
			Id:          "12",
			DisplayName: "aboba",
		}}

		if !reflect.DeepEqual(users, want) {
			t.Errorf("\ngot: %v\nwant: %v", users, want)
		}
	})

	t.Run("empty parameters returns error", func(t *testing.T) {
		client, _ := NewClient(creds, nil)
		ctx := context.Background()
		_, _, err := client.Users.GetUsers(ctx, nil)
		assertErrorPresence(t, err)
		assertErrorMessage(t, err, userIdLoginIsRequired)

		_, _, err = client.Users.GetUsers(ctx, &UsersOptions{})
		assertErrorPresence(t, err)
	})

	t.Run("tests limit of 100 parameters", func(t *testing.T) {
		c, _, _, teardown := setup()
		defer teardown()
		ctx := context.Background()

		ids := [101]string{}

		_, _, err := c.Users.GetUsers(ctx, &UsersOptions{
			Ids: ids[:],
		})
		assertErrorPresence(t, err)
		assertErrorMessage(t, err, users100LimitError)

		logins := [71]string{}

		_, _, err = c.Users.GetUsers(ctx, &UsersOptions{
			Ids:    ids[:30],
			Logins: logins[:],
		})
		assertErrorPresence(t, err)
		assertErrorMessage(t, err, users100LimitError)
	})
}
