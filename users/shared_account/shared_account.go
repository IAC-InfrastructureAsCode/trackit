//   Copyright 2018 MSolution.IO
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package shared_account

import (
	"database/sql"
	"net/http"

	"github.com/trackit/trackit-server/routes"
	"github.com/trackit/trackit-server/users"
	"github.com/trackit/trackit-server/db"
)

// inviteUserRequest is the expected request body for the invite user route handler.
type InviteUserRequest struct {
	Email           string `json:"email" req:"nonzero"`
	AccountId       int    `json:"accountId"`
	PermissionLevel int    `json:"permissionLevel"`
}

var listUserSharedAccountRequest = routes.QueryArg{
	Name:        "AccountId",
	Type:        routes.QueryArgInt{},
	Description: "AWS Account ID",
	Optional:    false,
}

type updateUsersSharedAccountRequest struct {
	ShareId         int `json:"shareId" req:"nonzero"`
	PermissionLevel int `json:"permissionLevel"`
}

type deleteUsersSharedAccountRequest struct {
	ShareId         int `json:"shareId" req:"nonzero"`
}

func init() {
	routes.MethodMuxer{
		http.MethodGet: routes.H(listSharedUsers).With(
			db.RequestTransaction{db.Db},
			users.RequireAuthenticatedUser{users.ViewerAsParent},
			routes.Documentation{
				Summary:     "List shared users",
				Description: "Return a list of user who have an access to an AWS account on Trackit",
			},
			routes.QueryArgs{
				listUserSharedAccountRequest,
			},
		),
		http.MethodPost: routes.H(inviteUser).With(
			db.RequestTransaction{db.Db},
			users.RequireAuthenticatedUser{users.ViewerAsParent},
			routes.RequestContentType{"application/json"},
			routes.RequestBody{InviteUserRequest{"example@example.com", 1234, 0}},
			routes.Documentation{
				Summary:     "Creates an invite",
				Description: "Creates an invite for account team sharing",
			},
		),
		http.MethodPatch: routes.H(updateSharedUsers).With(
			db.RequestTransaction{db.Db},
			users.RequireAuthenticatedUser{users.ViewerAsParent},
			routes.RequestContentType{"application/json"},
			routes.RequestBody{updateUsersSharedAccountRequest{1, 2}},
			routes.Documentation{
				Summary:     "Update shared users",
				Description: "Update shared users associated with a specific AWS account",
			},
		),
		http.MethodDelete: routes.H(deleteSharedUsers).With(
			db.RequestTransaction{db.Db},
			users.RequireAuthenticatedUser{users.ViewerAsParent},
			routes.RequestBody{deleteUsersSharedAccountRequest{1}},
			routes.Documentation{
				Summary:     "Delete shared users",
				Description: "Delete shared users associated with a specific AWS account",
			},
		),
	}.H().With(
		db.RequestTransaction{db.Db},
		routes.Documentation{
			Summary: "interact with shared accounts",
		},
	).Register("/user/share")
}

// inviteUser handles users invite for team sharing.
func inviteUser(request *http.Request, a routes.Arguments) (int, interface{}) {
	var body InviteUserRequest
	routes.MustRequestBody(a, &body)
	tx := a[db.Transaction].(*sql.Tx)
	user := a[users.AuthenticatedUser].(users.User)
	return InviteUserWithValidBody(request, body, tx, user)
}

// listSharedUsers handles listing of users who have an access to an AWS account.
func listSharedUsers(request *http.Request, a routes.Arguments) (int, interface{}) {
	body := a[listUserSharedAccountRequest].(int)
	tx := a[db.Transaction].(*sql.Tx)
	user := a[users.AuthenticatedUser].(users.User)
	return listSharedUserAccessWithValidBody(request, body, tx, user)
}

// updateSharedUsers handles updates of user permission level for team sharing.
func updateSharedUsers(request *http.Request, a routes.Arguments) (int, interface{}) {
	var body updateUsersSharedAccountRequest
	routes.MustRequestBody(a, &body)
	tx := a[db.Transaction].(*sql.Tx)
	return updateSharedUserAccessWithValidBody(request, body, tx)
}

// deleteSharedUsers handles user access deletion for team sharing
func deleteSharedUsers(request *http.Request, a routes.Arguments) (int, interface{}) {
	var body deleteUsersSharedAccountRequest
	routes.MustRequestBody(a, &body)
	tx := a[db.Transaction].(*sql.Tx)
	return deleteSharedUserAccessWithValidBody(request, body, tx)
}

// listSharedUserAccessWithValidBody tries to list users who have an access to an AWS account
func listSharedUserAccessWithValidBody(request *http.Request, body int, tx *sql.Tx, user users.User) (int, interface{}) {
	res, err := GetSharingList(request.Context(), db.Db, body)
	if err != nil {
		return 403, "Error retrieving shared users list"
	} else {
		return 200, res
	}
}

// updateSharedUserAccessWithValidBody tries to update users permission level for team sharing
func updateSharedUserAccessWithValidBody(request *http.Request, body updateUsersSharedAccountRequest, tx *sql.Tx) (int, interface{}) {
	res, err := UpdateSharedUser(request.Context(), db.Db, body.ShareId, body.PermissionLevel)
	if err != nil {
		return 403, "Error updating shared user list"
	}
	return 200, res
}

// deleteSharedUserAccessWithValidBody tries to delete users from accessing specific shared aws account
func deleteSharedUserAccessWithValidBody(request *http.Request, body deleteUsersSharedAccountRequest, tx *sql.Tx) (int, interface{}) {
	err := DeleteSharedUser(request.Context(), db.Db, body.ShareId)
	if err != nil {
		return 403, "Error deleting shared user"
	}
	return 200, nil
}
