import "@typespec/http";
import "@typespec/openapi";
import "../models/agent_tools.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Agent Tools")
namespace AgentTools {
  /** Get an agent's tool configuration. */
  @get
  @route("/api/v1/agents/{agentId}/tools")
  @summary("Get agent tools")
  @operationId("getAgentTools")
  op getAgentTools(@path agentId: string): AgentToolConfigResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Update an agent's tool configuration (replaces all tool assignments). */
  @put
  @route("/api/v1/agents/{agentId}/tools")
  @summary("Update agent tools")
  @operationId("updateAgentTools")
  op updateAgentTools(@path agentId: string, @body body: UpdateAgentToolConfigRequest): AgentToolConfigResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
