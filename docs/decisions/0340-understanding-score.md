# 0340. Understanding score

Status: accepted. An exam gives a separate score under [0450](0450-exam.md).

## Context

Progress listed counts of lessons and answers per topic. A count says what a learner did and not how well they understand. A lifetime ratio of correct to attempted answers punishes an early miss forever and says nothing about a learner who has since learned the topic. The quiz, with its fixed length, is already the tool's measure of understanding.

## Decision

A topic's understanding is its latest finished quiz.

- **Score.** A correct answer earns 1, a partial answer earns one half, and an incorrect or skipped answer earns 0. The score is the sum divided by the quiz goal, shown as a whole percentage. The goal is the quiz length, so the length setting from 0300 sets the size of the sample.
- **Latest quiz.** A quiz is finished when every question has an answer or a skip. The finished quiz replaces the topic's earlier score, so a learner who reviews and retakes can raise it. A quiz in progress does not change the stored score. A finished quiz that is extended is finished again when its new questions are answered, and a re-answered skip updates the score.
- **No quiz, no score.** A topic with no finished quiz has no score and counts as 0% in the course figure. Reading is shown on its own as "read" and is not blended into the percentage, so a row says what is missing. A topic marked known by the probe counts as 100% until a quiz replaces it.
- **Bands.** A score of 80% or more is "solid", 60% to 79% is "developing", and under 60% is "needs review". The bands sit next to the number because the assessments come from the model and a bare figure would claim more precision than the evidence has.
- **Course figures.** Progress and the sidebar show two numbers. Coverage is the topics read out of all topics. Understanding is the mean of the scores over the topics read, where a read topic with no quiz counts as 0%. A mean over every topic would sit near zero for most of a large course and say nothing true.
- **Review.** "Review a weak topic" on Home names the read topic with the lowest finished quiz under 60%. Plain mode's review command uses the same rule.

A topic stores the credit in half points and the quiz goal it was earned against, so the figure survives a change to the quiz length setting. The session version rises to 5.

## Consequences

A learner sees a number that moves when they study and quiz, and the number says which topics lack a quiz. A learner who reads and never quizzes has coverage and no understanding, which is accurate. A three-question quiz moves the score in steps of about 17 points, so the length setting changes how coarse the figure is. This replaces the counts shown on Progress and the answer tally in the sidebar. The lesson and answer counts still exist for the recommendation rules.
