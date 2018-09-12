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
	"context"

	"github.com/trackit/jsonlog"

	"github.com/trackit/trackit-server/models"
)

type SharedResults struct {
	ShareId int
	Mail string
	Level int
	UserId int
	SharingStatus bool
}

// GetSharingList returns a list of users who have access to a specific AWS account
func GetSharingList(ctx context.Context, db models.XODB, accountId int) ([]SharedResults, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	var response []SharedResults
	dbSharedAccounts, err := models.SharedAccountsByAccountID(db, accountId)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		logger.Error("Error getting shared account from database.", err.Error())
		return nil, err
	} else {
		for _, key := range dbSharedAccounts {
			dbUser, err := models.UserByID(db, key.UserID)
			if err != nil {
				logger.Error("Error getting users from database.", err.Error())
				return nil, err
			}
			response = append(response, SharedResults{key.ID, dbUser.Email,key.UserPermission,key.UserID,key.SharingAccepted})
		}
		return response, nil
	}
}

// UpdateSharedUser updates user permission level
func UpdateSharedUser(ctx context.Context, db models.XODB, shareId int, permissionLevel int) (interface{}, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	dbSharedAccount, err := models.SharedAccountByID(db, shareId)
	if err != nil {
		logger.Error("Error while getting shared user information", err)
		return nil, err
	}
	dbSharedAccount.UserPermission = permissionLevel
	err = dbSharedAccount.Update(db)
	if err != nil {
		logger.Error("Error while updating user permission", err)
		return nil, err
	}
	return dbSharedAccount, nil
}


// DeleteSharedUser deletes a user access to an AWS account by removing entry in shared_account database table.
func DeleteSharedUser(ctx context.Context, db models.XODB, shareId int) (error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	dbSharedAccount, err := models.SharedAccountByID(db, shareId)
	if err != nil {
		logger.Error("Error while getting shared user information", err)
		return err
	}
	err = dbSharedAccount.Delete(db)
	if err != nil {
		logger.Error("Error while deleting shared user", err)
		return err
	}
	return nil
}
