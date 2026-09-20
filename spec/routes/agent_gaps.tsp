import "@typespec/http";
import "@typespec/openapi";
import "../models/agent_gaps.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Agent Gaps")
namespace AgentGaps {
  /** List an agent's self-reported gaps. */
  @get
  @route("/api/v1/agents/{agentId}/gaps")
  @summary("List agent gaps")
  @operationId("listGaps")
  op listGaps(@path agentId: string, @query limit?: int32 = 20, @query offset?: int32 = 0): GapListResponse;

  /** Get a single gap by ID. */
  @get
  @route("/api/v1/agents/{agentId}/gaps/{gapId}")
  @summary("Get agent gap")
  @operationId("getGap")
  op getGap(@path agentId: string, @path gapId: string): GapResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Update a gap's review status (resolve, dismiss, or reopen). */
  @put
  @route("/api/v1/agents/{agentId}/gaps/{gapId}")
  @summary("Update agent gap")
  @operationId("updateGap")
  op updateGap(@path agentId: string, @path gapId: string, @body body: UpdateGapRequest): GapResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
