# my-msg

A Go message service rewrite using Gin, GORM, and MySQL. The current
implementation provides configuration validation, database initialization,
health routes, and template data access. Template HTTP CRUD remains a guide
in `reports/stage2-http-guide.md` and is not registered by the application.

## Run locally

Requirements: Go 1.26.4 and a MySQL database initialized with
`sql/msgcenter.sql`.

```sh
cp config/config.example.toml config/config-test.toml
# Set local MySQL connection details in the ignored configuration file.
go run ./src --config ./config/config-test.toml
```

The default port is 8091. `GET /test` and `GET /health` return HTTP 200.
Database configuration or connection failures prevent HTTP startup.

## Validation

```sh
go test ./... -count=1 -timeout=30s
go vet ./...
go build -o /tmp/my-msg ./src
MY_MSG_TEST_CONFIG="$PWD/config/config-test.toml" \
  go test -tags=integration ./src/data -v -count=1 -timeout=30s
```

Integration tests require a configured database and the template table. Test
rows are created inside transactions and rolled back during cleanup.
