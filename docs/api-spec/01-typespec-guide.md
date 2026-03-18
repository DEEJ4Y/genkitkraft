# TypeSpec — Generic Writing Guide

A general-purpose reference for writing TypeSpec API definitions, independent of any specific project.

---

## Table of Contents

1. [What is TypeSpec?](#1-what-is-typespec)
2. [Project Setup](#2-project-setup)
3. [Imports & Using](#3-imports--using)
4. [Namespaces](#4-namespaces)
5. [Scalars & Built-in Types](#5-scalars--built-in-types)
6. [Models](#6-models)
7. [Enums & Unions](#7-enums--unions)
8. [Operations](#8-operations)
9. [HTTP & REST Decorators](#9-http--rest-decorators)
10. [Generics (Templates)](#10-generics-templates)
11. [Interfaces](#11-interfaces)
12. [Decorators](#12-decorators)
13. [Aliases](#13-aliases)
14. [Versioning](#14-versioning)
15. [OpenAPI Extensions](#15-openapi-extensions)
16. [Quick Reference](#16-quick-reference)

---

## 1. What is TypeSpec?

TypeSpec is a language for describing APIs. You write `.tsp` files that compile to OpenAPI (Swagger), JSON Schema, or other output formats. It provides type safety, composition, and reusability that raw OpenAPI YAML/JSON lacks.

---

## 2. Project Setup

### Initialize a project

```bash
npm install -g @typespec/compiler
tsp init                    # Interactive scaffolding
tsp install                 # Install dependencies
tsp compile .               # Compile to output format
```

### tspconfig.yaml

```yaml
emit:
  - "@typespec/openapi3"
options:
  "@typespec/openapi3":
    emitter-output-dir: "{output-dir}/openapi3"
    openapi-versions:
      - 3.1.0
    file-type: "json" # or "yaml"
```

### Common dependencies (package.json)

```json
{
  "dependencies": {
    "@typespec/compiler": "^0.64.0",
    "@typespec/http": "^0.64.0",
    "@typespec/openapi": "^0.64.0",
    "@typespec/openapi3": "^0.64.0",
    "@typespec/rest": "^0.64.0",
    "@typespec/versioning": "^0.64.0"
  }
}
```

---

## 3. Imports & Using

### `import` — load files and libraries

`import` brings `.tsp` files or installed library packages into scope. It does **not** make names directly available — it only loads the declarations. You still need `using` to reference them without fully-qualified names.

#### Importing installed libraries

```typespec
import "@typespec/http";          // HTTP protocol decorators & models
import "@typespec/openapi";       // OpenAPI-specific decorators
import "@typespec/openapi3";      // OpenAPI 3.x emitter support
import "@typespec/rest";          // REST conventions
import "@typespec/versioning";    // API versioning support
```

#### Importing local files

```typespec
import "./models.tsp";                // File in the same directory
import "./sub-feature/main.tsp";      // File in a subdirectory
import "../../common";                // A directory (loads its main.tsp)
import "../shared/errors.tsp";       // Relative path to another module
```

**Key rules:**

- Paths are relative to the current file.
- Importing a **directory** (e.g., `import "./common"`) automatically loads `main.tsp` inside it.
- Importing a **file** requires the `.tsp` extension.
- Importing a **library** (e.g., `import "@typespec/http"`) loads the package from `node_modules`.
- Import order does not matter — all imports are resolved before compilation.
- Circular imports are not allowed.

### `using` — bring namespaces into scope

After importing, use `using` to reference types and decorators without their full namespace prefix.

```typespec
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.Http;        // Now you can write @get instead of @TypeSpec.Http.get
using OpenAPI;              // Now you can write @extension instead of @OpenAPI.extension
```

#### Using your own namespaces

```typespec
import "../../common";

using Common;               // Access Common.SingleResponse as SingleResponse
using Common.V3;            // Access nested namespace members directly
```

#### Without `using`

You can always reference things by their fully-qualified name:

```typespec
import "@typespec/http";

// Without `using TypeSpec.Http`:
@TypeSpec.Http.get
@TypeSpec.Http.route("/users")
op getUsers(): TypeSpec.Http.Body<User[]>;

// With `using TypeSpec.Http`:
@get
@route("/users")
op getUsers(): Body<User[]>;
```

### Import vs Using — summary

| Concept  | What it does                                       | Analogy                               |
| -------- | -------------------------------------------------- | ------------------------------------- |
| `import` | Loads a file/library so its declarations exist     | `#include` / `require`                |
| `using`  | Brings a namespace into scope for shorthand access | `using namespace` / `from X import *` |

### Common import + using combinations

```typespec
// Typical file header for an HTTP API module
import "@typespec/http";
import "@typespec/openapi";
import "../../common";
import "./models.tsp";

using TypeSpec.Http;
using OpenAPI;
using Common;
```

### Re-exporting from a main.tsp

A `main.tsp` file can serve as a barrel that imports sub-modules to compose a full API:

```typespec
// squadcast/main.tsp
import "./users/main.tsp";
import "./teams/main.tsp";
import "./services/main.tsp";
import "./incidents/main.tsp";

@service(#{ title: "My API" })
namespace MyApi;
```

All imported namespaces become part of the compiled output without needing explicit `using`.

---

## 4. Namespaces

Namespaces group related types and operations. Everything declared inside a namespace belongs to it.

```typespec
namespace MyApi;

model User {
  id: string;
}
```

### Nested namespaces

```typespec
namespace MyApi.V1.Users;
```

### Namespace blocks

```typespec
namespace MyApi {
  namespace V1 {
    model User { id: string; }
  }
}
```

### File-level namespace

A bare `namespace` at the top of a file scopes everything in that file:

```typespec
namespace MyApi.V1.Users;

// Everything below belongs to MyApi.V1.Users
model User { ... }
op getUser(): ...;
```

---

## 5. Scalars & Built-in Types

### Primitive types

| Type                 | Description                    |
| -------------------- | ------------------------------ |
| `string`             | UTF-8 string                   |
| `boolean`            | true / false                   |
| `integer`            | Arbitrary-precision integer    |
| `float`              | Floating-point number          |
| `int32`, `int64`     | Sized integers                 |
| `float32`, `float64` | Sized floats                   |
| `bytes`              | Binary data                    |
| `plainDate`          | Date without time (YYYY-MM-DD) |
| `plainTime`          | Time without date (HH:MM:SS)   |
| `utcDateTime`        | Full UTC date-time             |
| `duration`           | ISO 8601 duration              |
| `url`                | URL string                     |

### Literal types

```typespec
model Config {
  mode: "fast" | "slow";           // String literal union
  timeout: 5 | 10 | 30;           // Numeric literal union
  enabled: true;                    // Boolean literal
}
```

### Arrays

```typespec
tags: string[];
users: User[];
```

### Records (maps)

```typespec
metadata: Record<string>;          // { [key: string]: string }
permissions: Record<boolean>;      // { [key: string]: boolean }
```

### Nullable

```typespec
deletedAt: utcDateTime | null;     // Required but can be null
notes?: string | null;             // Optional and nullable
```

---

## 6. Models

### Basic model

```typespec
model User {
  id: string;
  name: string;
  email: string;
  role?: string;                    // Optional (? suffix)
  createdAt: utcDateTime;
}
```

### Inheritance (`extends`)

The child model includes all parent fields and can add more:

```typespec
model AdminUser extends User {
  permissions: string[];
}
```

### Copy pattern (`is`)

Creates a model with the same shape but no type relationship:

```typespec
model CreateUserInput is User {}    // Same fields, independent type
```

### Spread (`...`)

Inline-mix another model's fields:

```typespec
model AuditFields {
  createdAt: utcDateTime;
  updatedAt: utcDateTime;
}

model User {
  id: string;
  name: string;
  ...AuditFields;                   // Adds createdAt and updatedAt
}
```

### Nested inline objects

```typespec
model User {
  id: string;
  address: {
    street: string;
    city: string;
    zip: string;
  };
}
```

### Default values

```typespec
model PaginationParams {
  pageSize?: int32 = 20;
  page?: int32 = 1;
}
```

---

## 7. Enums & Unions

### Enums

```typespec
enum Color {
  Red,
  Green,
  Blue,
}

// With explicit values
enum Status {
  Active: "active",
  Inactive: "inactive",
  Pending: "pending",
}

// Names with special characters
enum Permission {
  `read-users`,
  `write-users`,
  `delete-users`,
}
```

### Unions (named)

```typespec
union Priority {
  low: "low",
  medium: "medium",
  high: "high",
  critical: "critical",
}
```

### Inline unions

```typespec
model Task {
  status: "open" | "in_progress" | "done";
  priority: 1 | 2 | 3 | 4 | 5;
}
```

### Discriminated unions

```typespec
@discriminator("kind")
union Shape {
  circle: Circle,
  square: Square,
}

model Circle {
  kind: "circle";
  radius: float64;
}

model Square {
  kind: "square";
  side: float64;
}
```

### Alias unions (for combining models)

```typespec
alias ApiError =
  | BadRequestError
  | UnauthorizedError
  | NotFoundError
  | InternalError;
```

---

## 8. Operations

Operations define API endpoints.

### Basic operation

```typespec
op getUser(@path id: string): User;
```

### With full HTTP decoration

```typespec
@get
@route("/users/{id}")
@summary("Get a user by ID")
@tag("Users")
op getUser(@path id: string): User;
```

### Parameters

```typespec
op searchUsers(
  @path orgId: string,              // Path parameter
  @query search?: string,           // Query parameter
  @query("page_size") pageSize?: int32,  // Query with custom name
  @header Authorization: string,    // HTTP header
  @body body: SearchRequest,        // Request body
): UserList;
```

### Return types

```typespec
// Single return
op getUser(): User;

// Union return (success + errors)
op getUser(): User | NotFoundError;

// Inline response with status code
op createUser(@body body: CreateUserRequest): {
  @statusCode statusCode: 201;
  @body body: User;
} | BadRequestError;
```

---

## 9. HTTP & REST Decorators

Requires `import "@typespec/http"` and `using TypeSpec.Http`.

### Method decorators

```typespec
@get       // GET
@post      // POST
@put       // PUT
@patch     // PATCH
@delete    // DELETE
@head      // HEAD
```

### Routing

```typespec
@route("/api/v1")
namespace MyApi.V1;

@route("/users")
op listUsers(): User[];

@route("/users/{id}")
op getUser(@path id: string): User;
```

Routes compose: namespace route + operation route = full path.

### Parameter decorators

| Decorator     | Maps to               |
| ------------- | --------------------- |
| `@path`       | URL path segment      |
| `@query`      | URL query string      |
| `@header`     | HTTP header           |
| `@body`       | Request/response body |
| `@statusCode` | HTTP status code      |

### Query parameter options

```typespec
@query("custom_name") myParam: string,          // Rename in URL
@query(#{ explode: true }) filters: string[],   // ?filters=a&filters=b
```

### Response body wrapper

```typespec
op getUser(): Body<{ data: User }>;   // Wraps response in Body<>
```

### Service metadata

```typespec
@service(#{ title: "My API" })
@info(#{ version: "2.0.0" })
@server("https://api.example.com", "Production")
@useAuth(BearerAuth)
namespace MyApi;
```

### Tags

```typespec
@tag("Users")
op getAllUsers(): User[];
```

---

## 10. Generics (Templates)

TypeSpec supports generic (template) models and operations.

### Generic models

```typespec
model ApiResponse<T> {
  data: T;
  success: boolean;
}

model PaginatedResponse<T> {
  data: T[];
  total: int32;
  page: int32;
  pageSize: int32;
}

// Usage
op getUser(): ApiResponse<User>;
op listUsers(): PaginatedResponse<User>;
```

### Generic with defaults

```typespec
model ErrorResponse<Code = 500> {
  @statusCode statusCode: Code;
  message: string;
}
```

### Generic with constraints

```typespec
model KeyedResponse<T extends { id: string }> {
  item: T;
  key: string;
}
```

---

## 11. Interfaces

Interfaces define a contract of operations that can be implemented.

```typespec
interface CRUD<T, TCreate, TUpdate> {
  list(): T[];
  read(@path id: string): T;
  create(@body body: TCreate): T;
  update(@path id: string, @body body: TUpdate): T;
  delete(@path id: string): void;
}
```

### Extending interfaces

```typespec
interface UserOps extends CRUD<User, CreateUser, UpdateUser> {
  @route("/me")
  getMe(): User;
}
```

---

## 12. Decorators

Decorators annotate declarations with metadata.

### Built-in decorators

| Decorator               | Target              | Purpose                        |
| ----------------------- | ------------------- | ------------------------------ |
| `@doc("...")`           | Any                 | Documentation string           |
| `@summary("...")`       | Operation           | Short summary                  |
| `@tag("...")`           | Operation/Namespace | OpenAPI tag grouping           |
| `@key("fieldName")`     | Model property      | Mark as identifier             |
| `@minLength(n)`         | String property     | Minimum length                 |
| `@maxLength(n)`         | String property     | Maximum length                 |
| `@minValue(n)`          | Numeric property    | Minimum value                  |
| `@maxValue(n)`          | Numeric property    | Maximum value                  |
| `@pattern("regex")`     | String property     | Regex constraint               |
| `@format("...")`        | String property     | Format hint (email, uri, etc.) |
| `@secret`               | String property     | Marks as sensitive             |
| `@deprecated("reason")` | Any                 | Mark as deprecated             |
| `@visibility("read")`   | Property            | Control read/write visibility  |
| `@example(value)`       | Any                 | Provide example value          |
| `@encode("...")`        | Scalar              | Encoding format                |

### JSDoc-style documentation

```typespec
/**
 * Retrieves a user by their unique identifier.
 *
 * Requires `read` scope on the access token.
 */
@get
op getUser(@path id: string): User;
```

---

## 13. Aliases

Aliases create shorthand names for types.

```typespec
alias UserId = string;
alias UserList = User[];

alias CommonErrors =
  | BadRequestError
  | UnauthorizedError
  | InternalError;

// Use in operations
op getUser(): User | CommonErrors;
```

---

## 14. Versioning

Requires `import "@typespec/versioning"` and `using TypeSpec.Versioning`.

### Define versions

```typespec
@versioned(Versions)
namespace MyApi;

enum Versions {
  v1: "v1",
  v2: "v2",
}
```

### Version-conditional members

```typespec
model User {
  id: string;
  name: string;

  @added(Versions.v2)
  avatar?: string;              // Only exists in v2+

  @removed(Versions.v2)
  legacyField: string;          // Removed in v2
}
```

### Co-existing version namespaces (alternative approach)

Instead of using the `@versioned` system, you can simply use separate namespaces:

```typespec
namespace MyApi.V1.Users { ... }
namespace MyApi.V2.Users { ... }
```

---

## 15. OpenAPI Extensions

Add custom `x-*` properties to the OpenAPI output.

```typespec
import "@typespec/openapi";
using OpenAPI;

@extension("x-internal", true)
op internalEndpoint(): void;

@extension(
  "x-speakeasy-pagination",
  #{
    type: "cursor",
    inputs: #[
      #{ name: "cursor", in: "parameters", type: "cursor" },
    ],
    outputs: #{ nextCursor: "$.pageInfo.nextCursor" },
  }
)
@get
op listItems(): ItemList;
```

**Note:** The `#{ }` and `#[ ]` syntax creates object/array values for extensions.

---

## 16. Quick Reference

### Minimal API file

```typespec
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.Http;
using OpenAPI;

@service(#{ title: "My API" })
@server("https://api.example.com", "Production")
@useAuth(BearerAuth)
namespace MyApi;

model Item {
  id: string;
  name: string;
  createdAt: utcDateTime;
}

@tag("Items")
@route("/items")
@get
@summary("List all items")
op listItems(): Body<{ data: Item[] }>;

@tag("Items")
@route("/items/{id}")
@get
@summary("Get item by ID")
op getItem(@path id: string): Body<{ data: Item }>;

@tag("Items")
@route("/items")
@post
@summary("Create an item")
op createItem(@body body: { name: string }): {
  @statusCode statusCode: 201;
  @body body: { data: Item };
};

@tag("Items")
@route("/items/{id}")
@delete
@summary("Delete an item")
op deleteItem(@path id: string): {
  @statusCode statusCode: 204;
};
```

### Common patterns cheat sheet

```typespec
// Optional field
name?: string;

// Nullable field
value: string | null;

// Optional + nullable
value?: string | null;

// Array
items: Item[];

// Map / dictionary
metadata: Record<string>;

// Literal union
status: "active" | "inactive";

// Spread
model B { ...A; extraField: string; }

// Inheritance
model B extends A { extraField: string; }

// Generic
model Wrapper<T> { data: T; }

// Alias
alias Errors = Error1 | Error2 | Error3;
```
