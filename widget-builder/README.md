# widget-builder

Chat widget builder UI, shared by the dashboard (`ui/`, Agent → **Widget** tab) and the docs website (`website/`, `/widget-builder`). It designs configs for the [AI chat widget](https://github.com/techorionai/ai-chat-widget) (`navigableai-chat-widget`, pinned in `src/constants.ts`) and generates embed code.

## Layout

| File | Purpose |
| --- | --- |
| `src/state.ts` | Config state: `setIn`/`unsetIn` (prune empty values), persistence, share links |
| `src/snippets.ts`, `src/providerSource.ts` | npm/CDN snippets, the `GenkitkraftChatProvider` source (TS and JS), backend example |
| `src/sections.tsx`, `src/fields.tsx` | Form sections and path-bound inputs |
| `src/Preview.tsx` | Injects the real widget into the page with a mock provider |
| `src/WidgetBuilder.tsx` | The assembled builder |

The provider source in `src/providerSource.ts` is also shown in `website/docs/guides/chat-widget.md`. Keep them identical.

## Develop

```bash
npm install
npm test          # vitest
npm run build     # tsup -> dist/ (what the apps consume)
```

`ui/` and `website/` depend on this package with `file:../widget-builder` and `install-links=true`, so they get a **copy**, not a symlink (a symlink would load a second copy of React). After changing the package:

```bash
npm run build
cd ../ui        # or ../website
rm -rf node_modules/genkitkraft-widget-builder && npm install
rm -rf .next/cache dist/cache    # ui only: Next caches the old copy
npx docusaurus clear             # website only
```

Peer dependencies (`react`, `@mantine/core`, `@mantine/hooks`, `@tabler/icons-react`) come from the consuming app.
