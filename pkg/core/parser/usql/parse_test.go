package usql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestParseUSQLConfig(t *testing.T) {
	data := `
connections:
  my_couchbase_conn: couchbase://Administrator:P4ssw0rd@localhost
  my_clickhouse_conn: clickhouse://clickhouse:P4ssw0rd@localhost
  my_godror_conn:
    protocol: godror
    username: system
    password: P4ssw0rd
    hostname: localhost
    port: 1521
    database: free
`

	var config Config
	err := yaml.Unmarshal([]byte(data), &config)
	assert.NoError(t, err)

	assert.Len(t, config.Connections, 3)
}
