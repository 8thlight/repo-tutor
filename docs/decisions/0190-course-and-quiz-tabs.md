# 0190. Course and Quiz tabs

Status: accepted

## Context

ADR 0180 named a Learn tab, then folded the Quiz into it. Every page in the tool serves learning, so Learn named the whole tool and told the learner nothing about where a page lived. Sources and Hotspots were also given positions they did not earn. Sources lists files cited in a lesson, so it only makes sense from a lesson. Hotspots is a ranked list with no next step; its data is better used inside lessons, which a later record will decide.

## Decision

The tab bar has four tabs: Home, Course, Quiz, and Progress. The tab names an activity, as in 0050.

- Course covers the lesson, Topics, Sources, Hotspots, and source files. Sources and Hotspots are sub-pages of the course and have no tab. Ctrl+K, Ctrl+S, and Ctrl+O open them, and Course stays highlighted.
- Quiz covers the quiz question and its index. Ctrl+G and Ctrl+B open them.
- Progress is unchanged. Help, About, and command output belong to no tab.

Tab and Shift+Tab step through Home, Course, Quiz, and Progress. Course opens the lesson and Quiz opens the question index.

A source file remembers the list that opened it, as in 0180.

The Home page names its first entry Course and drops the Hotspots entry. Hotspots is reached with Ctrl+O until a later record replaces it.

## Consequences

Sources is a place a lesson leads to and not a place to travel to. The Hotspots page has one key and no tab, so it is easy to miss; that is accepted because it is to be replaced. This supersedes 0180 and restores the Quiz tab that 0180 removed.
