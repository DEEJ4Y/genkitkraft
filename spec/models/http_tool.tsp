namespace Api.Models;

/** HTTP method for an HTTP tool. */
enum HttpMethod {
  GET: "GET",
  POST: "POST",
  PUT: "PUT",
  DELETE: "DELETE",
  PATCH: "PATCH",
  HEAD: "HEAD",
  OPTIONS: "OPTIONS",
}

/** A key-value pair representing an HTTP header. Values may contain Go template expressions. */
model HttpToolHeader {
  /** Header name (e.g. Content-Type, Authorization). */
  name: string;

  /** Header value. Supports Go template expressions (e.g. Bearer {{.token}}). */
  value: string;
}

/** An HTTP tool that can be attached to agents for LLM function calling. */
model HttpToolResponse {
  /** Unique HTTP tool ID. */
  id: string;

  /** Display name for this tool. The LLM uses this as the function name. */
  name: string;

  /** Description of what this tool does. The LLM uses this to decide when to call the tool. */
  description: string;

  /** HTTP method to use when calling the endpoint. */
  method: HttpMethod;

  /** URL of the API endpoint. Supports Go template expressions (e.g. https://api.example.com/{{.resource}}). */
  url: string;

  /** HTTP headers to include in the request. */
  headers: HttpToolHeader[];

  /** Request body template. Supports Go template expressions. */
  bodyTemplate: string;

  /** JSON Schema describing the input parameters the LLM must provide when calling this tool. */
  inputSchema: string;

  /** When this tool was created. */
  createdAt: utcDateTime;

  /** When this tool was last updated. */
  updatedAt: utcDateTime;
}

/** Request to create a new HTTP tool. */
model CreateHttpToolRequest {
  /** Display name for this tool. */
  name: string;

  /** Description of what this tool does. */
  description?: string;

  /** HTTP method to use. */
  method: HttpMethod;

  /** URL of the API endpoint. Supports Go template expressions. */
  url: string;

  /** HTTP headers to include. */
  headers?: HttpToolHeader[];

  /** Request body template. Supports Go template expressions. */
  bodyTemplate?: string;

  /** JSON Schema describing the input parameters for LLM function calling. */
  inputSchema?: string;
}

/** Request to update an existing HTTP tool. */
model UpdateHttpToolRequest {
  /** Updated display name. */
  name?: string;

  /** Updated description. */
  description?: string;

  /** Updated HTTP method. */
  method?: HttpMethod;

  /** Updated URL. Supports Go template expressions. */
  url?: string;

  /** Updated HTTP headers. */
  headers?: HttpToolHeader[];

  /** Updated request body template. */
  bodyTemplate?: string;

  /** Updated JSON Schema for input parameters. */
  inputSchema?: string;
}

/** Paginated list of HTTP tools. */
model HttpToolListResponse {
  /** Array of HTTP tools. */
  httpTools: HttpToolResponse[];

  /** Total number of HTTP tools. */
  total: int32;

  /** Number of HTTP tools per page. */
  limit: int32;

  /** Number of HTTP tools skipped. */
  offset: int32;
}
