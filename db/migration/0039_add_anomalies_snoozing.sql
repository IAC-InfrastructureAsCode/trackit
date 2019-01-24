--   Copyright 2018 MSolution.IO
--
--   Licensed under the Apache License, Version 2.0 (the "License");
--   you may not use this file except in compliance with the License.
--   You may obtain a copy of the License at
--
--       http://www.apache.org/licenses/LICENSE-2.0
--
--   Unless required by applicable law or agreed to in writing, software
--   distributed under the License is distributed on an "AS IS" BASIS,
--   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
--   See the License for the specific language governing permissions and
--   limitations under the License.

CREATE TABLE anomaly_snoozing (
	id       INTEGER      NOT NULL AUTO_INCREMENT,
	created  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
	user_id  INTEGER      NOT NULL,
	anomaly_id   VARCHAR(255) NOT NULL,
	CONSTRAINT PRIMARY KEY (id),
	CONSTRAINT UNIQUE (anomaly_id),
	CONSTRAINT foreign_user FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);