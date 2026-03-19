import "@typespec/http";
import "@typespec/openapi";
import "../models/auth.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Auth")
namespace Auth {
  /** Check whether authentication is required for this instance. */
  @get
  @route("/api/auth/status")
  @summary("Auth status")
  @operationId("getAuthStatus")
  op status(): AuthStatusResponse;

  /** Authenticate with username and password. Sets an HttpOnly session cookie on success. */
  @post
  @route("/api/auth/login")
  @summary("Login")
  @operationId("login")
  op login(@body body: LoginRequest): LoginResponse | {
    @statusCode statusCode: 401;
    @body body: ErrorResponse;
  } | {
    @statusCode statusCode: 429;
    @body body: ErrorResponse;
  };

  /** End the current session. Clears the session cookie. */
  @post
  @route("/api/auth/logout")
  @summary("Logout")
  @operationId("logout")
  op logout(): LogoutResponse | {
    @statusCode statusCode: 401;
    @body body: ErrorResponse;
  };

  /** Get the currently authenticated user. */
  @get
  @route("/api/auth/me")
  @summary("Current user")
  @operationId("getMe")
  op me(): MeResponse | {
    @statusCode statusCode: 401;
    @body body: ErrorResponse;
  };
}
