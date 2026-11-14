# Accessibility

trisse-home aims for **WCAG 2.2 level AA** (#113). Everything can be used with a keyboard and a screen reader.

## What is checked automatically

- **axe-core in the Playwright suite** (`web/e2e/a11y.ts`) runs the WCAG 2.0–2.2 A/AA rules on:
  - login, dashboard and edit mode;
  - the widget picker and the shortcuts overview;
  - **every settings section**;
  - the dashboard in dark mode.

  A new violation fails CI.

- **Contrast unit tests:**
  - The light and dark design tokens must reach 4.5:1 for text: normal and muted text on all surfaces, accent and status colours on surfaces, and text on accent (`web/src/shared/theme/tokens.test.ts`).
  - The same check runs for every built-in theme of the gallery (`TestBuiltinThemesMeetWCAGAA`).
- Components may only use colour tokens that exist. An undefined token would fall back to a colour with poor contrast, and a test catches it.

## Keyboard

| Keys                                | Action                                                               |
| ----------------------------------- | -------------------------------------------------------------------- |
| `Ctrl+K` or `/`                     | Open the quick search                                                |
| `1` … `9`                           | Switch to tab 1 to 9                                                 |
| `e`                                 | Start or finish edit mode                                            |
| `?`                                 | Show all shortcuts                                                   |
| `Ctrl+Z`, `Ctrl+Shift+Z` / `Ctrl+Y` | Undo and redo in edit mode                                           |
| Arrow keys / `Shift` + arrow keys   | Move / resize the focused widget in edit mode                        |
| Arrow keys, `Home`, `End`           | Navigate the widget picker                                           |
| `Esc`                               | Close dialogs, menus and the launcher; focus returns to where it was |

Shortcuts are ignored while you type in a field or a dialog is open.

The `?` overview in the app is the reference; this table can lag behind it. Dialogs are native `<dialog>` elements: focus stays inside and `Esc` closes them. Every focusable element shows a visible focus ring.

## Screen readers

- Widgets have headings, buttons have names that include their visible text (WCAG 2.5.3), and decorative icons are hidden.
- One shared **polite live region** announces a widget whose updates start failing ("Weather: update failed").
  - It says this once, when the status changes, not on every failed poll.
  - The same message is not repeated within five minutes, so polling widgets cannot flood it.
- Form errors use `role="alert"`.

## Motion

With _reduce motion_ switched on in the operating system:

- animations and transitions are shortened to nothing;
- smooth scrolling is off;
- the rotating placeholder of the search field stays still.

## Known limitations

- **Not yet audited manually with NVDA, JAWS or VoiceOver.** Automated checks find roughly a third of WCAG problems. Findings from a manual audit will be tracked as sub-issues of #113.
- Drag & drop in edit mode has a keyboard alternative (arrow keys), but no announcement of the new position yet.
- Charts (sparklines, heatmaps, radar) show their key values as text. The radar image itself has no text alternative beyond its title.
- Custom CSS and imported themes can lower the contrast. Only the built-in themes are checked.

## Reporting a problem

Please open an issue with the label `accessibility`. Include the browser, the assistive technology and the steps to reproduce.
