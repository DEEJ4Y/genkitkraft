import "@typespec/http";
import "../models/health.tsp";

using TypeSpec.Http;
using Api.Models;

namespace Api.Routes;

@tag("Health")
namespace Health {
  /** Liveness probe. Returns 200 if the server process is running. */
  @get
  @route("/livez")
  @summary("Liveness check")
  op liveness(): HealthCheckResponse;

  /** Readiness probe. Returns 200 if the server is ready to accept traffic. */
  @get
  @route("/readyz")
  @summary("Readiness check")
  op readiness(): HealthCheckResponse | {
    @statusCode statusCode: 503;
    @body body: HealthCheckResponse;
  };
}
