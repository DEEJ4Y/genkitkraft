import "@typespec/http";
import "@typespec/openapi";
import "../models/mcp_server.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("McpServers")
namespace McpServers {
  /** List all MCP server configurations with pagination. */
  @get
  @route("/api/v1/mcp-servers")
  @summary("List MCP servers")
  @operationId("listMcpServers")
  op listMcpServers(@query limit?: int32 = 20, @query offset?: int32 = 0): McpServerListResponse;

  /** Create a new MCP server configuration. */
  @post
  @route("/api/v1/mcp-servers")
  @summary("Create MCP server")
  @operationId("createMcpServer")
  op createMcpServer(@body body: CreateMcpServerRequest): {
    @statusCode statusCode: 201;
    @body body: McpServerResponse;
  };

  /** Get a single MCP server configuration by ID. */
  @get
  @route("/api/v1/mcp-servers/{id}")
  @summary("Get MCP server")
  @operationId("getMcpServer")
  op getMcpServer(@path id: string): McpServerResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Update an existing MCP server configuration. */
  @put
  @route("/api/v1/mcp-servers/{id}")
  @summary("Update MCP server")
  @operationId("updateMcpServer")
  op updateMcpServer(@path id: string, @body body: UpdateMcpServerRequest): McpServerResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Delete an MCP server configuration. */
  @delete
  @route("/api/v1/mcp-servers/{id}")
  @summary("Delete MCP server")
  @operationId("deleteMcpServer")
  op deleteMcpServer(@path id: string): {
    @statusCode statusCode: 204;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** List tools available from an MCP server by connecting to it. */
  @get
  @route("/api/v1/mcp-servers/{id}/tools")
  @summary("List MCP server tools")
  @operationId("listMcpServerTools")
  op listMcpServerTools(@path id: string): McpServerToolListResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
