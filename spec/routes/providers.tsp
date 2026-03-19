import "@typespec/http";
import "@typespec/openapi";
import "../models/provider.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Api.Models;

namespace Api.Routes;

@tag("Settings")
namespace Settings {
  /** List all configured providers. */
  @get
  @route("/api/v1/settings/providers")
  @summary("List providers")
  @operationId("listProviders")
  op listProviders(): ProviderListResponse;

  /** Create a new provider configuration. */
  @post
  @route("/api/v1/settings/providers")
  @summary("Create provider")
  @operationId("createProvider")
  op createProvider(@body body: CreateProviderRequest): {
    @statusCode statusCode: 201;
    @body body: ProviderResponse;
  } | {
    @statusCode statusCode: 409;
    @body body: ErrorResponse;
  };

  /** Get a single provider by ID. */
  @get
  @route("/api/v1/settings/providers/{id}")
  @summary("Get provider")
  @operationId("getProvider")
  op getProvider(@path id: string): ProviderResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Update an existing provider configuration. */
  @put
  @route("/api/v1/settings/providers/{id}")
  @summary("Update provider")
  @operationId("updateProvider")
  op updateProvider(@path id: string, @body body: UpdateProviderRequest): ProviderResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  } | {
    @statusCode statusCode: 409;
    @body body: ErrorResponse;
  };

  /** Delete a provider configuration. */
  @delete
  @route("/api/v1/settings/providers/{id}")
  @summary("Delete provider")
  @operationId("deleteProvider")
  op deleteProvider(@path id: string): {
    @statusCode statusCode: 204;
  } | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };

  /** Test provider connectivity by making a lightweight API call. */
  @post
  @route("/api/v1/settings/providers/{id}/test")
  @summary("Test provider connectivity")
  @operationId("testProvider")
  op testProvider(@path id: string): TestProviderResponse | {
    @statusCode statusCode: 404;
    @body body: ErrorResponse;
  };
}
