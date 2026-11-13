# Themes

The theme gallery (Settings → Appearance, #111) ships the themes in
[`internal/modules/appearance/adapter/gateway/themes/`](../internal/modules/appearance/adapter/gateway/themes/).
They are embedded in the binary, so adding a theme is one JSON file.

## Format

```json
{
  "id": "nord",
  "name": "Nord",
  "author": "Arctic Ice Studio (MIT)",
  "scheme": "dark",
  "tokens": {
    "--color-bg": "#2e3440",
    "--color-surface": "#3b4252",
    "--color-text": "#eceff4",
    "--color-accent": "#88c0d0"
  }
}
```

- **id:** 2–40 characters `a-z`, `0-9`, `-`. It is also the file name (`nord.json`).
- **scheme:** `light` or `dark`. Applying the theme switches to that scheme, so native controls and scrollbars match.
- **tokens:** only the design tokens of [`web/src/shared/theme/tokens.css`](../web/src/shared/theme/tokens.css):
  - colors as `#rgb` or `#rrggbb`: `--color-bg`, `--color-surface`, `--color-surface-2`, `--color-border`, `--color-text`, `--color-text-muted`, `--color-accent`, `--color-accent-contrast`, `--color-danger`, `--color-success`, `--color-warning`;
  - radii as `0px`–`32px`: `--radius`, `--radius-small`.

  `--color-bg`, `--color-surface`, `--color-text` and `--color-accent` are required. Set all colors for a polished result.

Free CSS, layout rules, fonts, `url(…)` and unknown fields are refused, for built-in and imported themes alike. Updates to the app therefore cannot break a theme, and a shared theme file cannot load anything or hide parts of the page.

## Contributing a theme

1. Add `<id>.json` to the folder above.
   - Check the contrast: text on `--color-bg` and `--color-surface`, and `--color-accent-contrast` on `--color-accent`, should reach WCAG AA (4.5:1).
   - Credit the original palette and its license in `author`.
2. Run `go test ./internal/modules/appearance/...`. The tests load every file and fail on an invalid one.
3. Open a pull request with a screenshot of the dashboard in the new theme.

## Sharing your own

In the gallery, **⬇** downloads any theme as a file and **Import theme…** adds one (at most 16 KB, 50 imported themes).

Applying a theme replaces your light/dark choice, custom CSS and accent color. Your own values are kept, and **Restore my own style** brings them back.
