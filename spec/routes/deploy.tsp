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

  /** Create a new stateful chat session for the given agent. */
  @post
  @route("/api/v1/agents/{agentId}/deploy/sessions")
  @summary("Create deploy session")
  @operationId("createDeploySession")
  op createDeploySession(@path agentId: string, @body body: CreateDeploySessionRequest): {
    @statusCode statusCode: 201;
    @body body: DeploySessionResponse;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: ErrorResponse;
  };

  /** Get a deploy session by ID. */
  @get
  @route("/api/v1/agents/{agentId}/deploy/sessions/{sessionId}")
  @summary("Get deploy session")
  @operationId("getDeploySession")
  op getDeploySession(@path agentId: string, @path sessionId: string): DeploySessionResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: ErrorResponse;
  };

  /** Delete a deploy session and all its messages. */
  @delete
  @route("/api/v1/agents/{agentId}/deploy/sessions/{sessionId}")
  @summary("Delete deploy session")
  @operationId("deleteDeploySession")
  op deleteDeploySession(@path agentId: string, @path sessionId: string): {
    @statusCode statusCode: 204;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  } | {
    @statusCode statusCode: 401;
    @body body: ErrorResponse;
  };

  /** Stateful OpenAI-compatible chat completions endpoint. Only the last user message is used; full history is loaded from the session. */
  @post
  @route("/api/v1/agents/{agentId}/deploy/sessions/{sessionId}/chat/completions")
  @summary("Deploy session chat completions")
  @operationId("deploySessionChatCompletions")
  op deploySessionChatCompletions(
    @path agentId: string,
    @path sessionId: string,
    @body body: DeployChatCompletionsRequest,
  ): DeployChatCompletionResponse | {
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
