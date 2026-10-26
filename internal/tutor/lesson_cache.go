package tutor

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"repo-tutor/internal/repository"
)

const lessonFormat = 1

const lessonHeaderOpen = "<!-- repo-tutor lesson "

const lessonHeaderClose = " -->\n\n"

type SavedLesson struct {
	Fingerprint string
	Goal        string
	Topic       string
	Reply       Reply
}

type lessonHeader struct {
	Format      int                   `json:"format"`
	Fingerprint string                `json:"fingerprint"`
	Goal        string                `json:"goal"`
	Topic       string                `json:"topic"`
	Question    string                `json:"question"`
	Citations   []repository.Location `json:"citations"`
}

func GoalKey(goal string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(goal)))[:16]
}

func LessonKey(goal, topic string) string {
	return GoalKey(goal) + "-" + fmt.Sprintf("%x", sha256.Sum256([]byte(topic)))[:12]
}

func (lesson SavedLesson) Fresh(fingerprint string) bool {
	return lesson.Fingerprint == fingerprint
}

func EncodeLesson(lesson SavedLesson) []byte {
	header, _ := json.Marshal(lessonHeader{lessonFormat, lesson.Fingerprint, lesson.Goal, lesson.Topic, lesson.Reply.Question, lesson.Reply.Citations})
	return []byte(lessonHeaderOpen + string(header) + lessonHeaderClose + lesson.Reply.Explanation + "\n")
}

func DecodeLesson(data []byte) (SavedLesson, bool) {
	rest, found := bytes.CutPrefix(data, []byte(lessonHeaderOpen))
	if !found {
		return SavedLesson{}, false
	}
	headerText, body, found := bytes.Cut(rest, []byte(lessonHeaderClose))
	if !found {
		return SavedLesson{}, false
	}
	var header lessonHeader
	if err := json.Unmarshal(headerText, &header); err != nil || header.Format != lessonFormat || len(header.Citations) == 0 {
		return SavedLesson{}, false
	}
	explanation := string(bytes.TrimSuffix(body, []byte("\n")))
	if explanation == "" {
		return SavedLesson{}, false
	}
	return SavedLesson{header.Fingerprint, header.Goal, header.Topic, Reply{Explanation: explanation, Question: header.Question, Assessment: "none", Citations: header.Citations}}, true
}
