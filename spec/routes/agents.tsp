import "@typespec/http";
import "@typespec/openapi";
import "../models/agent.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Agents")
namespace Agents {
  /** List all agents with pagination. */
  @get
  @route("/api/v1/agents")
  @summary("List agents")
  @operationId("listAgents")
  op listAgents(@query limit?: int32 = 20, @query offset?: int32 = 0): AgentListResponse;

  /** Create a new agent. */
  @post
  @route("/api/v1/agents")
  @summary("Create agent")
  @operationId("createAgent")
  op createAgent(@body body: CreateAgentRequest): {
    @statusCode statusCode: 201;
    @body body: AgentResponse;
  };

  /** Get a single agent by ID. */
  @get
  @route("/api/v1/agents/{id}")
  @summary("Get agent")
  @operationId("getAgent")
  op getAgent(@path id: string): AgentResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Update an existing agent. */
  @put
  @route("/api/v1/agents/{id}")
  @summary("Update agent")
  @operationId("updateAgent")
  op updateAgent(@path id: string, @body body: UpdateAgentRequest): AgentResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Delete an agent. */
  @delete
  @route("/api/v1/agents/{id}")
  @summary("Delete agent")
  @operationId("deleteAgent")
  op deleteAgent(@path id: string): {
    @statusCode statusCode: 204;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
