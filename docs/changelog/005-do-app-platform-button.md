# Add the "Deploy to DigitalOcean" one-click button and App Platform guide

**Date:** 2026-10-09
**Branch:** feat/do-app-platform-button
**Issue:** none

## Summary

The repo has a DigitalOcean App Platform template at `.do/deploy.template.yaml`. No page used it. This change adds the deploy button to the README, the home page, and the Installation page. It also adds a full App Platform guide. The template uses public default credentials, so each button has a trial-defaults warning.

## Changes

- `website/docs/deployment/app-platform.md`: Add the App Platform guide. It covers the components, deploy steps, first sign-in, the required update of `ENCRYPTION_KEY`, `AUTH_CREDENTIALS`, and `PUBLIC_API_KEY`, custom domain, scale up, limits, upgrade, and troubleshooting.
- `website/docs/deployment/digitalocean.md`, `docker.md`, `reverse-proxy.md`, `horizontal-scaling.md`: Move each page down one position in the sidebar. The App Platform page is first.
- `website/docs/deployment/digitalocean.md`, `docker.md`: Add a note that says when to use each deployment guide and links to the other two.
- `website/docs/getting-started/installation.md`: Add the button, a warning, and links to the App Platform and Droplet guides.
- `website/src/pages/index.tsx`, `index.module.css`: Add the button to the hero. Add a "Deploy in one click" section with a warning.
- `README.md`: Add the button and the trial-defaults warning below the intro.

## Notes

- The button image loads from `deploytodo.com`. The blue button on the blue hero has low contrast. The owner chose to keep it.
- The button works only when the change is on `main` and the repo is public. The button was not tested with a DigitalOcean account.
- Both `redis://` and `valkey://` are valid values for `CACHE_URL`. The template keeps `redis://valkey:6379`.
- The DigitalOcean documentation does not say that a deployment pulls a newer image for the `latest` tag. The Upgrade section tells the reader to pin a version tag.
- `npm run build` in `website` passed with no broken links.
