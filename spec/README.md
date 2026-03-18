# API Spec

This directory contains the API specifications for the project, defined using TypeSpec. The specifications are used to generate OpenAPI documentation, which can in turn be used to generate server stubs, and client SDKs, ensuring consistency across implementations.

Refer to the [TypeSpec Guide](docs/api-spec/01-typespec-guide.md) for more details on how to write and maintain these specifications.

## Requirements

- Node.js
- npm

## Setup

```bash
npm install
```

## Usage

```bash
# Compile once
npm run build

# Watch mode
npm run dev
```

Output is written to `tsp-output/schema/openapi.1.0.yaml`.

## Project Structure

- `main.tsp`: The main entry point for the TypeSpec compiler.
- `models/`: Contains TypeSpec definitions for data models used in the API.
- `routes/`: Contains TypeSpec definitions for API routes using the data models defined in `models/`.
- `common/`: Contains shared TypeSpec definitions and utilities used across models and routes.
