// Licensed to ClickHouse, Inc. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. ClickHouse, Inc. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package std

import (
	"context"
	"database/sql"
	"fmt"
	clickhouse_tests "github.com/ClickHouse/clickhouse-go/v2/tests"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStdConnCheck(t *testing.T) {
	const (
		ddl = `
		CREATE TABLE clickhouse_test_conncheck (
			Value String
		) Engine Memory
		`
		dml = `INSERT INTO clickhouse_test_conncheck VALUES `
	)

	env, err := clickhouse_tests.GetTestEnvironment("std")
	require.NoError(t, err)

	dsns := map[string]string{"Native": fmt.Sprintf("clickhouse://%s:%d?username=%s&password=%s", env.Host, env.Port, env.Username, env.Password),
		"Http": fmt.Sprintf("http://%s:%d?username=%s&password=%s&session_id=session", env.Host, env.HttpPort, env.Username, env.Password)}

	for name, dsn := range dsns {
		t.Run(fmt.Sprintf("%s Protocol", name), func(t *testing.T) {
			connect, err := sql.Open("clickhouse", dsn)
			require.NoError(t, err)
			// We can only change the settings at the connection level.
			// If we have only one connection, we change the settings specifically for that connection.
			connect.SetMaxOpenConns(1)
			_, err = connect.Exec("DROP TABLE IF EXISTS clickhouse_test_conncheck")
			require.NoError(t, err)
			_, err = connect.Exec(ddl)
			require.NoError(t, err)
			_, err = connect.Exec("set idle_connection_timeout=1")
			assert.NoError(t, err)

			_, err = connect.Exec("set tcp_keep_alive_timeout=0")
			assert.NoError(t, err)

			time.Sleep(1100 * time.Millisecond)
			ctx := context.Background()
			tx, err := connect.BeginTx(ctx, nil)
			assert.NoError(t, err)

			_, err = tx.PrepareContext(ctx, dml)
			assert.NoError(t, err)
			assert.NoError(t, tx.Commit())

			connect.Exec("DROP TABLE IF EXISTS clickhouse_test_conncheck")
		})
	}
}
