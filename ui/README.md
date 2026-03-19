# GenKitKraft UI

Next.js 14 (Pages Router) frontend for GenKitKraft, built with [Mantine](https://mantine.dev/) and statically exported for embedding in the Go binary.

## Getting Started

```bash
npm install
npm run dev
```

The dev server runs at `http://localhost:3000` and proxies `/api/*` requests to the Go backend at `http://localhost:8080`.

## Scripts

| Command | Description |
|---|---|
| `npm run dev` | Start development server |
| `npm run build` | Build static export to `dist/` |
| `npm run start` | Start production server |
| `npm run generate:api` | Regenerate TypeScript types from the OpenAPI spec |

## Project Structure

```
ui/
├── components/       # React components
│   ├── AppLayout     # Sidebar shell (Mantine AppShell)
│   ├── LoginDialog   # Authentication form
│   ├── ProviderCard  # LLM provider status card
│   └── ProviderForm  # Provider create/edit modal
├── lib/
│   ├── api/          # OpenAPI fetch client & generated types
│   └── auth.tsx      # Auth context, provider, and gate
├── pages/
│   ├── _app.tsx      # App entry (providers, layout)
│   ├── index.tsx     # Dashboard
│   └── settings.tsx  # LLM provider configuration
└── dist/             # Static build output (git-ignored)
```

## Key Libraries

- **Mantine v8** — UI components and AppShell layout
- **TanStack React Query** — Server state management
- **openapi-fetch / openapi-typescript** — Type-safe API client generated from the TypeSpec definition

## API Type Generation

When the backend API spec changes, regenerate the TypeScript types:

```bash
npm run generate:api
```

This reads from `../spec/tsp-output/schema/openapi.yaml` and writes to `lib/api/schema.d.ts`.
