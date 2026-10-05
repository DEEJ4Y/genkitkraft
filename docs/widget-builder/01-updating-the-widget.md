# Updating the Chat Widget Builder

How to change `widget-builder/` or bump the `navigableai-chat-widget` version without shipping a stale or broken builder.

## How the pieces fit

- `widget-builder/` is a React package built with tsup into `widget-builder/dist/`. It designs configs for the [AI chat widget](https://github.com/techorionai/ai-chat-widget) (`navigableai-chat-widget`) and generates embed code.
- Two apps consume it:
  - `ui/`: the dashboard (Agent → **Widget** tab).
  - `website/`: the docs site (`/widget-builder`).
- Both depend on it with `"genkitkraft-widget-builder": "file:../widget-builder"` and `install-links=true` (`ui/.npmrc`, `website/.npmrc`). They get a **copy** in `node_modules/genkitkraft-widget-builder`, not a symlink. A symlink would load a second copy of React.
- The live preview (`src/Preview.tsx`) injects the real widget into the page. The widget renders its UI in an iframe that is **hosted and versioned**: `https://chat.techorionai.com/builds/<version>/app/`. That code is not in this repo and can't be patched here.
- The `Dockerfile` builds `widget-builder` first, then `ui`.

Because of the copy, rebuilding `widget-builder` changes nothing in `ui/` or `website/` until the copy is refreshed. Bundler caches then hide the new copy as well. Skipping either step leads to testing old code.

## A. After changing `widget-builder/src`

1. In `widget-builder/`:
   ```bash
   npm test
   npm run typecheck
   npm run build        # tsup -> dist/, what the apps consume
   ```
2. In each consumer you want to test (`ui/`, `website/`), refresh the copy. A bare `npm install` does nothing because the lockfile hasn't changed, so remove the old copy first:
   ```bash
   rm -rf node_modules/genkitkraft-widget-builder && npm install
   ```
3. Clear the bundler caches, then restart the dev server:
   ```bash
   # ui/
   rm -rf .next/cache dist/cache
   # website/
   rm -rf node_modules/.cache/webpack && npx docusaurus clear
   ```
4. Check the consumer really has the new code. Pick a symbol you just added:
   ```bash
   grep -c <newSymbol> node_modules/genkitkraft-widget-builder/dist/index.js
   cmp ../widget-builder/dist/index.js node_modules/genkitkraft-widget-builder/dist/index.js && echo identical
   ```

`node_modules` is not tracked by git, so the refresh is a local step. Nothing to commit for it.

On Windows the Docusaurus dev server may listen on IPv6 only. Use `http://localhost:<port>/genkitkraft/widget-builder`, not `127.0.0.1`.

## B. Bumping `navigableai-chat-widget`

### Places to change together

| What | Where |
| --- | --- |
| Dependency | `widget-builder/package.json` and `widget-builder/package-lock.json` (`npm install navigableai-chat-widget@<version> --save-exact` in `widget-builder/`) |
| Version constant | `WIDGET_VERSION` in `widget-builder/src/constants.ts` (the CDN URL is derived from it) |
| Hard-coded versions in user docs | `website/docs/guides/chat-widget.md` (the `npm install navigableai-chat-widget@…` line and the CDN `<script>` URL) |
| Provider source | `widget-builder/src/providerSource.ts` and the copy shown in `website/docs/guides/chat-widget.md` must stay identical |

Then follow procedure A.

### Re-check the workarounds for the hosted app

The builder carries workarounds for bugs in the hosted iframe app. For a new version, test whether each is still needed, and remove the ones that aren't, along with their tests.

| Workaround | Where | Why it exists |
| --- | --- | --- |
| Drop incomplete primary colors | `isValidPrimaryColor` in `src/colors.ts`, applied by `normalizeConfig` in `src/normalize.ts` | The color input reports every keystroke. The iframe throws on values like `#` or `#ff` and stays blank, and the bad value was persisted in `localStorage`. |
| Send `grape` and `dark` as hex | `toWidgetPrimaryColor` in `src/colors.ts`, applied by `normalizeConfig` | The iframe runs every primary color through chroma-js, which only knows CSS color names. `grape` and `dark` throw `unknown hex color`. The "Minimal" preset uses `dark`. |
| Same config shape on update as on init | `buildWidgetConfig` in `src/widgetConfig.ts`, used by `src/Preview.tsx` | `actionsMap` is always an object, welcome actions without a matching action are dropped, and `chatProvider` is present. `injectAiChatWidget` does this on first load. |
| Send `override_config` only after the iframe `load` event | `src/Preview.tsx` | Events sent before the iframe has loaded are lost or race the `init` → `set_config` handshake. |
| Apply the color scheme with `toggleColorScheme`, not only `override_config` | `push` in `src/Preview.tsx`, `getColorScheme` in `src/widgetConfig.ts` | The hosted app's message listener compares the requested scheme with the scheme it had at mount (a stale value), so switching back to that scheme is skipped. Choosing Dark or the Dark preset silently did nothing when the iframe's saved scheme was dark. `toggleColorScheme` applies unconditionally. |

`normalizeConfig` is shared by the live preview and the generated snippets, so the workarounds also protect people who paste the embed code.

### New or changed config

Read `node_modules/navigableai-chat-widget/README.md` and `src/types.ts` in `widget-builder/` for added, renamed or removed options. Update the form in `src/sections.tsx`, the presets in `src/presets.ts`, and the user docs in `website/docs/guides/chat-widget.md`.

## C. Verify before merging

Unit tests don't cover the hosted iframe, so also check in a browser.

1. `npm test`, `npm run typecheck` and `npm run build` pass in `widget-builder/`.
2. Refresh the consumer copy (procedure A) and open the builder from `website/` or `ui/`.
3. With the browser console open (including the iframe context), confirm the preview stays up with no errors after each of these:
   - Pick **every** name in the Primary color dropdown.
   - Switch the default color scheme Light → Dark → Light, and apply the Dark preset then Default. Read `data-mantine-color-scheme` on the iframe's `<html>` after each. Repeat after a reload with dark saved.
   - Type `#ff0000` into Custom primary color one character at a time.
   - Apply each preset.
   - Edit a handful of other fields (agent name, logos, message colors, actions, "Start expanded").
   - Reload the page. A saved config must not crash the preview.
4. Open the generated npm and CDN snippets and check they reflect the config, including the `grape`/`dark` hex conversion.

To automate step 3, Playwright works against the real hosted iframe (it needs network access). Install it in a scratch directory, not in the repo, and read the iframe's text after each change. A blank iframe means it crashed.
