import "@typespec/http";
import "@typespec/openapi";
import "../models/deploy.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Deploy")
namespace Deploy {
  /** Stateless OpenAI-compatible chat completions endpoint. The caller provides the full message history on every request. The agent's configured system prompt is injected server-side. */
  @post
  @route("/api/v1/agents/{agentId}/deploy/chat/completions")
  @summary("Deploy chat completions")
  @operationId("deployChatCompletions")
  op deployChatCompletions(@path agentId: string, @body body: DeployChatCompletionsRequest): DeployChatCompletionResponse | {
    @statusCode statusCode: 400;
    @body body: ErrorResponse;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: ErrorResponse;
  };
}
