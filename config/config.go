//   Copyright 2017 MSolution.IO
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

package config

import ()

// Configuration holds a configuration for the Trackit server.
type Configuration struct {
	// HTTPAddress is the address and port the server shall bind to.
	HTTPAddress string
	// SQLProtocol is the name of the SQL database, as used in the protocol in the URL.
	SQLProtocol string
	// SQLAddress is the string passed to the SQL driver to connect to the database.
	SQLAddress string
	// HashDifficulty is the difficulty used by the hashing function used to store passwords.
	HashDifficulty int
	// AuthIssuer is the issuer included in JWT tokens.
	AuthIssuer string
	// AuthSecret is the secret used to sign and verify JWT tokens.
	AuthSecret []byte
	// AwsRegion is the AWS region the product operates in.
	AwsRegion string
	// BackendId is an identifier for the current instance of the server.
	BackendId string
}

var configurationInitialized = false
var configuration Configuration

// LoadConfiguration loads the server's configuration.
func LoadConfiguration() Configuration {
	if !configurationInitialized {
		configuration = BuildDefaultConfiguration()
		configurationInitialized = true
	}
	return configuration
}

// BuildDefaultConfiguration returns a sane default configuration for the
// server.
func BuildDefaultConfiguration() Configuration {
	return Configuration{
		HTTPAddress:    "[::]:8080",
		SQLProtocol:    "mysql",
		SQLAddress:     "root:rootpassword@tcp([::1]:3306)/db",
		HashDifficulty: 12,
		AuthIssuer:     "trackit",
		AuthSecret:     []byte("trackitdefaultsecret"),
		AwsRegion:      "us-west-2",
		BackendId:      "dev",
	}
}
