namespace Api.Models;

/** Status of a health check probe. */
enum HealthStatus {
  Up: "up",
  Down: "down",
}

/** Response body for health check endpoints. */
model HealthCheckResponse {
  /** Current status of the probe. */
  status: HealthStatus;

  /** Human-readable description of the check result. */
  message: string;
}
