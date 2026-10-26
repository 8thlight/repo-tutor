# 0170. Tab groups

Status: accepted

## Context

The tab bar listed seven pages: Home, Learn, Topics, Quiz, Sources, Hotspots, and Progress. They were not peers. Topics is how a learner chooses what to learn, so it is a way of moving around Learn. Sources and Hotspots both look at the repository itself and have nothing to do with the lesson on screen once the learner leaves the cited files. A bar of seven names made the learner decide where a thing lived before they could start.

## Decision

The tab bar names the five activities: Home, Learn, Quiz, Explore, and Progress. This keeps the rule from 0050 that the highlighted tab names the activity and not the page.

- Learn covers the lesson and the Topics list. Ctrl+K opens Topics with Learn still highlighted.
- Quiz covers the quiz question and the question index, as before.
- Explore covers Sources, source files, and Hotspots. Ctrl+S opens Sources, Ctrl+O opens Hotspots, and each page names the other's key in its title. Tab opens Explore on Sources.
- Progress is unchanged.
- Help, About, and command output belong to no tab, so no tab is highlighted.

Tab and Shift+Tab step through the five tabs in that order. Choosing a tab opens its main page: the lesson for Learn, the question index for Quiz, Sources for Explore.

The Home page replaces its Browse entry with two: Topics, which opens the Topics list, and Explore, which opens Sources. The other rules in 0080 stand.

## Consequences

The tab bar is shorter and the grouping matches what the learner is doing. Topics is one key press from Learn and is no longer a place to travel to. A learner who looks for Hotspots under its own tab must learn that it sits under Explore, which the title and Home entry state. Number keys on Home shift because the entries changed.
