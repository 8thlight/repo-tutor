# 0030. Question pacing and quiz goals

Status: accepted

## Context

Showing the next lesson and question alongside feedback makes it hard to tell which answer an assessment refers to. An open-ended quiz also gives the learner no visible finish line. The learner asked for an explicit next step and a question index with a goal.

## Decision

Stop after assessing an answer or explaining a skipped exercise. Request the next exercise only when the learner chooses Continue. Keep hints attached to the current question and allow follow-up questions before continuing. Save the review stage so restarting does not advance the course.

Start each quiz with a five-question goal. Generate one question at a time and show upcoming slots in the question index. Count attempted or skipped questions toward completion, with correctness recorded separately. After finishing, offer an explicit extension of five more questions. Topic changes leave the quiz; starting another quiz replaces its index while retaining overall topic outcomes.

Replace the lesson pane during model requests with a loading view using the existing spinner. Keep local browsing available. Highlight source with compiled Chroma lexers and retain a plain fallback. These choices preserve the self-contained binary and agent-driver boundaries.
