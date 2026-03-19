package gen

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi-types.yaml ../../../spec/tsp-output/schema/openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi-server.yaml ../../../spec/tsp-output/schema/openapi.yaml
