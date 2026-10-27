# 0240. Highlight colors

Status: accepted

## Context

The workspace uses one accent color for the active tab, headings, the focus marker, the spinner, and the composer border. It is currently teal, which is not part of the 8th Light brand. The brand palette has seven accent colors: red `#F32735`, orange `#FF8300`, yellow `#FFD100`, green `#93D500`, blue `#0071CE`, dark blue `#004987`, and purple `#A81B8D`. No one of these reads at 4.5 to 1 against both a dark and a light terminal background. Orange, yellow, and green are too light for a white background. Blue, dark blue, and purple are too dark for a black one. Red is borderline on both.

## Decision

The learner chooses the highlight from the seven brand colors. Green is the default, since green is the color associated with learning. The settings file field `highlight` holds `green`, `red`, `orange`, `yellow`, `blue`, `dark_blue`, or `purple`. A value of `teal` from an earlier file is unknown, so it falls back to green under 0220.

The tool shifts a chosen color to stay readable. It does this at draw time and does not store a shifted value.

1. Take the terminal's reported background, or the fallback near-black or near-white that 0210 uses.
2. If the brand color already has a contrast ratio of at least 4.6 against that background, use it unchanged.
3. Otherwise convert the color to hue, saturation, and lightness. Keep hue and saturation. Move lightness toward lighter on a dark background and darker on a light one, by the smallest amount that reaches 4.6. Contrast rises steadily with lightness, so the smallest amount is found by search.
4. The margin of 0.1 over 4.5 covers rounding to eight-bit channels.

The same shifted color is used as text and as the tab fill. The tab text is the terminal background, so the ratio that protects text also protects the tab. The muted color and the other palette entries are unchanged.

The Settings page gains a Highlight color group after Markdown files open, in the brand's order. Each row shows a swatch drawn in the shifted color, so the learner sees what they will get on their terminal, then the name. The group carries the line "Colors shift to stay readable on your terminal's background." The page keeps the focused row in view when its list is taller than the screen. With `NO_COLOR` the page says "Color is off, so highlight color does not apply."

## Consequences

The accent matches the brand, and green on a dark terminal is the brand green unchanged. On a light terminal the same choice becomes a darker green, and yellow becomes a dark gold. That is the intended result and not a fault. On a dark terminal, blue and dark blue lighten toward each other and look alike; both stay, because they differ on a light terminal. Teal is gone. The shift is a pure function with no stored state, so a terminal that changes its background gets a correct color on the next report.
