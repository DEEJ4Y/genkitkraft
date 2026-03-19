.PHONY: generate generate-spec generate-go generate-ts build

generate: generate-spec generate-go generate-ts

generate-spec:
	cd spec && npx tsp compile .

generate-go:
	go generate ./internal/api/gen/...

generate-ts:
	cd ui && npm run generate:api

build:
	go build ./cmd/server/...
