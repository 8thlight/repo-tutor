# 0130. Plain cursor movement

Status: accepted

## Context

Bubble Tea v2 redraws by sending only the cells that changed, and it picks shortcuts from the terminal type. For `xterm-256color` it moves the cursor backward with a backward tab (`ESC [ Z`). A terminal that does not implement that sequence, such as the one built into an IDE, leaves the cursor in the wrong column. Rows then overlap, text lands in the wrong place, and leftover cells from the previous page show through. The corruption appeared on pages with many styled rows, such as a rendered README, and not in kitty.

## Decision

Tell the renderer that the terminal is a `vt100`, so it uses only basic cursor movement and no backward tabs, repeat sequences, or scroll shortcuts. Detect the color profile from the real environment first and pass it explicitly, so colors do not change. Apply this to the workspace only. Plain mode does not use the renderer.

## Consequences

Redraws send somewhat more bytes. Rendering is the same in kitty and in terminals that lack the optional sequences. A test checks the substitution, and a pseudo-terminal run against a terminal emulator confirmed the corrupted rows disappear. If a future Bubble Tea release can turn off only the backward tab, that is a narrower fix and replaces this one.
