package tutor

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"repo-tutor/internal/repository"
)

const overviewFormat = 1

const overviewHeaderOpen = "<!-- repo-tutor overview "

const overviewHeaderClose = " -->\n\n"

type SavedOverview struct {
	Fingerprint string
	Goal        string
	Reply       Reply
}

type overviewHeader struct {
	Format      int                   `json:"format"`
	Fingerprint string                `json:"fingerprint"`
	Goal        string                `json:"goal"`
	Question    string                `json:"question"`
	Citations   []repository.Location `json:"citations"`
}

func GoalKey(goal string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(goal)))[:16]
}

func (overview SavedOverview) Fresh(fingerprint string) bool {
	return overview.Fingerprint == fingerprint
}

func EncodeOverview(overview SavedOverview) []byte {
	header, _ := json.Marshal(overviewHeader{overviewFormat, overview.Fingerprint, overview.Goal, overview.Reply.Question, overview.Reply.Citations})
	return []byte(overviewHeaderOpen + string(header) + overviewHeaderClose + overview.Reply.Explanation + "\n")
}

func DecodeOverview(data []byte) (SavedOverview, bool) {
	rest, found := bytes.CutPrefix(data, []byte(overviewHeaderOpen))
	if !found {
		return SavedOverview{}, false
	}
	headerText, body, found := bytes.Cut(rest, []byte(overviewHeaderClose))
	if !found {
		return SavedOverview{}, false
	}
	var header overviewHeader
	if err := json.Unmarshal(headerText, &header); err != nil || header.Format != overviewFormat || len(header.Citations) == 0 {
		return SavedOverview{}, false
	}
	explanation := string(bytes.TrimSuffix(body, []byte("\n")))
	if explanation == "" {
		return SavedOverview{}, false
	}
	return SavedOverview{header.Fingerprint, header.Goal, Reply{Explanation: explanation, Question: header.Question, Assessment: "none", Citations: header.Citations}}, true
}
