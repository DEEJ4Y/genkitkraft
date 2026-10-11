# Fix the hero buttons on mobile

**Date:** 2026-10-11
**Branch:** fix/hero-buttons-mobile
**Issue:** none

## Summary

The hero on the home page has three buttons. On a narrow screen, the third button was cut off. The button row could not wrap, and the hero hides overflow. This change lets the row wrap onto a new line.

## Changes

- `website/src/pages/index.module.css`: Add `flex-wrap: wrap` and `gap: 0.75rem` to `.buttons`. Remove `margin-left` from `.mcpHeroButton` and `.deployHeroButton`. The `gap` replaces it.

## Notes

- On a wide screen, the three buttons stay on one line.
- `npm run build` in `website` passed.
- The layout was not checked in a browser at 360px, 414px, and 1200px widths.
