namespace Api.Models;

/** Response indicating whether authentication is required. */
model AuthStatusResponse {
  /** True if AUTH_CREDENTIALS is configured and non-empty. */
  required: boolean;
}

/** Request body for the login endpoint. */
model LoginRequest {
  /** Username to authenticate with. */
  username: string;

  /** Plaintext password. */
  password: string;
}

/** Response after successful login. */
model LoginResponse {
  /** Authenticated username. */
  username: string;
}

/** Response after successful logout. */
model LogoutResponse {
  ok: boolean;
}

/** Current authenticated user info. */
model MeResponse {
  /** Username of the currently authenticated user. */
  username: string;
}

/** Standard error response. */
model ErrorResponse {
  /** Human-readable error description. */
  error: string;
}
