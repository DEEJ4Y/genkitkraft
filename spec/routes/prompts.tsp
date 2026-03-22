import "@typespec/http";
import "@typespec/openapi";
import "../models/prompt.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Prompts")
namespace Prompts {
  /** List all prompts with pagination. */
  @get
  @route("/api/v1/prompts")
  @summary("List prompts")
  @operationId("listPrompts")
  op listPrompts(@query limit?: int32 = 20, @query offset?: int32 = 0): PromptListResponse;

  /** Create a new prompt. */
  @post
  @route("/api/v1/prompts")
  @summary("Create prompt")
  @operationId("createPrompt")
  op createPrompt(@body body: CreatePromptRequest): {
    @statusCode statusCode: 201;
    @body body: PromptResponse;
  };

  /** Get a single prompt by ID. */
  @get
  @route("/api/v1/prompts/{id}")
  @summary("Get prompt")
  @operationId("getPrompt")
  op getPrompt(@path id: string): PromptResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Update an existing prompt. */
  @put
  @route("/api/v1/prompts/{id}")
  @summary("Update prompt")
  @operationId("updatePrompt")
  op updatePrompt(@path id: string, @body body: UpdatePromptRequest): PromptResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Delete a prompt. */
  @delete
  @route("/api/v1/prompts/{id}")
  @summary("Delete prompt")
  @operationId("deletePrompt")
  op deletePrompt(@path id: string): {
    @statusCode statusCode: 204;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
