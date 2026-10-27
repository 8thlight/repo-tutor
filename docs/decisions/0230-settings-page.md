# 0230. Settings page

Status: accepted

## Context

ADR 0220 stores a learner's settings in a file. Nothing lets the learner change them yet. The learner is someone mid-lesson who wants the workspace to look right, then wants to get back. They are not configuring software, so the page must be short, show what each choice does before it is chosen, and never leave them unsure whether a change took effect or was saved.

Two choices have a risk the learner cannot see. The slanted and rounded tab styles draw with Powerline glyphs, and a terminal without a patched font shows boxes. The page must let the learner find that out before committing, and must make undoing it trivial.

## Decision

`F3` opens the Settings page and, pressed again on it, returns to the origin, as 0110 and 0200 describe. It highlights no tab. The Home page lists Settings before About, with the line "Look and behavior. F3 opens it from anywhere." Help lists F3. The tab line hints become `Tab Next  F1 Help  F2 About  F3 Settings`, and they shorten in this order: `F1 Help  F2 About  F3 Settings`, then `F1 Help  F2 About`, then `F1 Help`.

The page opens with one muted line that names its keys, "↑↓ move · Space chooses · F3 or Esc closes", and a blank line. It then shows a single list of choices under muted headings. Up and Down move between choices, skipping headings. Space selects the focused choice. Enter does nothing here, because Enter opens things on other pages and a stray Enter must not change a setting. The selected choice in each group shows `(•)` and the others show `( )`, so the state never rests on color alone. The focused row uses the same `›` marker as other lists. A choice applies and saves at once, and the notice reads "Saved". There is no Apply button and no Cancel; changing back is one more selection. `F3` and Esc close the page.

The groups, in order:

- **Tab style** lists `None`, `Block`, `Slanted`, and `Rounded`. Each row shows a sample of the bar drawn in that style, built by the same function that draws the real bar, so a sample cannot differ from the result. The real tab bar above the page also changes at once. The Slanted and Rounded rows carry the line "Needs a Powerline or Nerd Font in your terminal." at all times, not only when selected, so a learner reads it before choosing. A learner who sees boxes in the sample picks Block and is back where they started.
- **Markdown files open** lists `Rendered` and `Source`, and sets the view a Markdown file opens in. The page does not mention `v`, which belongs to the file view. The per-file choice from `v` stays for the session and wins over the setting.
- **Highlight color** is not shown until the branding set holds more than one color. The field exists in the file now, so the page can add the group without a format change.

Tab styles draw as follows. None shows plain labels with the active label bold and in the accent color, and the rule from 0260 marks it. Block is the fill from 0210. Slanted and Rounded use the same fills as Block and add a glyph on each side of every tab, colored as the fill on the terminal's own background. Slanted uses U+E0BA before the label and U+E0B8 after it. Rounded uses U+E0B6 before and U+E0B4 after. With no color, every style falls back to the bracketed active tab from 0210.

Settings load once at startup. A file that could not be read, or that comes from a newer repo-tutor, shows its notice from 0220 once at startup and again as a line at the top of the page. In the read-only case, choosing a row shows the notice and changes nothing. With `NO_COLOR` set, the page says "Color is off, so tab style does not apply." and Tab style rows are shown but do not change the bar. A failed write keeps the change for the running session and shows "Could not save. This change lasts until you quit."

At widths under 80 columns the samples move under their labels so no row wraps mid-sample.

## Consequences

A learner sees every tab style before choosing and can undo a bad one in a keystroke. The page is also one more list that moves with the arrow keys like the rest of the workspace, and it adds one new key, Space, to choose. Powerline glyph codes are fixed here so a rewrite draws the same tabs. Two limits are accepted: the tool cannot detect a missing font, so it relies on the sample and the standing note, and the page shows a Markdown setting that affects only files opened afterward.
