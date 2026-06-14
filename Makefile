.PHONY: generate generate-spec generate-go generate-ts build test-integration

generate: generate-spec generate-go generate-ts

generate-spec:
	cd spec && npx tsp compile .

generate-go:
	go generate ./internal/api/gen/...

generate-ts:
	cd ui && npm run generate:api

build:
	go build ./cmd/server/...

test-integration:
	docker compose -f docker-compose.test.yml up -d --wait
	TEST_POSTGRES_URL="postgres://test:test@localhost:15432/genkitkraft?sslmode=disable" \
	TEST_MYSQL_URL="test:test@tcp(localhost:3306)/genkitkraft?parseTime=true" \
	TEST_MARIADB_URL="test:test@tcp(localhost:3307)/genkitkraft?parseTime=true" \
	  go test -tags integration -v \
	    ./internal/adapters/postgres_db/... \
	    ./internal/adapters/postgres_agent/... \
	    ./internal/adapters/postgres_agent_tool/... \
	    ./internal/adapters/postgres_http_tool/... \
	    ./internal/adapters/postgres_mcp_server/... \
	    ./internal/adapters/postgres_playground/... \
	    ./internal/adapters/postgres_prompt/... \
	    ./internal/adapters/postgres_provider/... \
	    ./internal/adapters/mysql_db/... \
	    ./internal/adapters/mysql_agent/... \
	    ./internal/adapters/mysql_agent_tool/... \
	    ./internal/adapters/mysql_http_tool/... \
	    ./internal/adapters/mysql_mcp_server/... \
	    ./internal/adapters/mysql_playground/... \
	    ./internal/adapters/mysql_prompt/... \
	    ./internal/adapters/mysql_provider/...
	docker compose -f docker-compose.test.yml down
