# 0290. Hotspots as a lesson

Status: accepted. `/hotspots` is removed from the full-screen interface under [0320](0320-no-slash-commands-in-the-interface.md), and the lesson is in the Topics list.

## Context

The Hotspots page lists files ranked by recent change and branch points. A ranked list gives a learner no next step: they see a file name and a score and must decide what to do with it. The data is already computed, and the tutor already receives the top ten files as reading hints with every request. The tutor can turn that data into a lesson, which is what a learner came for.

## Decision

The course gains a topic, "Where change happens", with the id `hotspots`. It comes last in the course, after the foundations, because it assumes the learner knows the packages. It exists only when the repository has at least one hotspot.

When this topic is the subject of a request, the source excerpts start with the five highest hotspots, each at the line where its hint points, so the lesson cites the files the hints name. The prompt gains one instruction for this topic: explain what the hints show about where change concentrates, trace what the top file is responsible for in the sources, end with one comprehension question, and say that hints are reading priorities and not defects. The caveats about the Git window and the branch estimates stay in the prompt as before.

The Hotspots page goes away. That removes `Ctrl+O`, the hotspots list, and the rule that returns a source file to the hotspots list. Source files now return only to Sources. The Course tab covers the lesson, Topics, Sources, and source files. `/hotspots` in the full-screen interface opens the topic's lesson, as `/topic` does for any topic. The `scan` command and the plain interface keep their printed list of hotspots, since they are text output and not a page.

## Consequences

A learner reads about the files that matter and is asked a question about them, and the files they cited appear in Sources. The ranked list is no longer a screen, so a learner who wanted only the list uses `repo-tutor scan`. The topic is the last in the course, so a learner who follows the recommended order reaches it after the foundations. The page, its key, and the footer caveat for it are removed with no replacement, since the lesson carries the caveat. This changes the Hotspots sub-page in 0190 and the Ctrl+O key in 0230, and no other rule there.
