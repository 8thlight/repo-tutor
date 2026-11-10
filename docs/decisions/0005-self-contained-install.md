---
category: architecture
status: accepted
approval: person
partly-superseded-by: ['0630']
---

# 0005. Self-contained binary

A desktop app now joins the terminal app, and the build rule is in [0630](0630-a-desktop-app-joins-the-terminal-app.md). The terminal app stays one Go binary.

## Decision

The tool ships as one Go binary. It needs no runtime beyond the agent CLI the learner already has installed.
