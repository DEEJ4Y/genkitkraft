import "@typespec/http";
import "@typespec/openapi";
import "../models/built_in_tool.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Built-in Tools")
namespace BuiltInTools {
  /** List all available built-in tools. */
  @get
  @route("/api/v1/built-in-tools")
  @summary("List built-in tools")
  @operationId("listBuiltInTools")
  op listBuiltInTools(): BuiltInToolListResponse;
}
