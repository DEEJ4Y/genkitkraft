import "@typespec/http";
import "@typespec/openapi";
import "../models/playground.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Playground")
namespace Playground {
  /** List all playground sessions for an agent. */
  @get
  @route("/api/v1/agents/{agentId}/playground/sessions")
  @summary("List playground sessions")
  @operationId("listPlaygroundSessions")
  op listPlaygroundSessions(@path agentId: string): PlaygroundSessionListResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Create a new playground session for an agent. */
  @post
  @route("/api/v1/agents/{agentId}/playground/sessions")
  @summary("Create playground session")
  @operationId("createPlaygroundSession")
  op createPlaygroundSession(@path agentId: string, @body body: CreatePlaygroundSessionRequest): {
    @statusCode statusCode: 201;
    @body body: PlaygroundSessionResponse;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Delete a playground session. */
  @delete
  @route("/api/v1/agents/{agentId}/playground/sessions/{sessionId}")
  @summary("Delete playground session")
  @operationId("deletePlaygroundSession")
  op deletePlaygroundSession(@path agentId: string, @path sessionId: string): {
    @statusCode statusCode: 204;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Get all messages in a playground session. */
  @get
  @route("/api/v1/agents/{agentId}/playground/sessions/{sessionId}/messages")
  @summary("List playground messages")
  @operationId("listPlaygroundMessages")
  op listPlaygroundMessages(@path agentId: string, @path sessionId: string): PlaygroundMessageListResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Send a message and receive a streaming response (SSE). */
  @post
  @route("/api/v1/agents/{agentId}/playground/chat")
  @summary("Playground chat")
  @operationId("playgroundChat")
  op playgroundChat(@path agentId: string, @body body: PlaygroundChatRequest): PlaygroundMessageResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
