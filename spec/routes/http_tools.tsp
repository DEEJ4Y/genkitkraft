import "@typespec/http";
import "@typespec/openapi";
import "../models/http_tool.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("HttpTools")
namespace HttpTools {
  /** List all HTTP tools with pagination. */
  @get
  @route("/api/v1/http-tools")
  @summary("List HTTP tools")
  @operationId("listHttpTools")
  op listHttpTools(@query limit?: int32 = 20, @query offset?: int32 = 0): HttpToolListResponse;

  /** Create a new HTTP tool. */
  @post
  @route("/api/v1/http-tools")
  @summary("Create HTTP tool")
  @operationId("createHttpTool")
  op createHttpTool(@body body: CreateHttpToolRequest): {
    @statusCode statusCode: 201;
    @body body: HttpToolResponse;
  };

  /** Get a single HTTP tool by ID. */
  @get
  @route("/api/v1/http-tools/{id}")
  @summary("Get HTTP tool")
  @operationId("getHttpTool")
  op getHttpTool(@path id: string): HttpToolResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Update an existing HTTP tool. */
  @put
  @route("/api/v1/http-tools/{id}")
  @summary("Update HTTP tool")
  @operationId("updateHttpTool")
  op updateHttpTool(@path id: string, @body body: UpdateHttpToolRequest): HttpToolResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Delete an HTTP tool. */
  @delete
  @route("/api/v1/http-tools/{id}")
  @summary("Delete HTTP tool")
  @operationId("deleteHttpTool")
  op deleteHttpTool(@path id: string): {
    @statusCode statusCode: 204;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
