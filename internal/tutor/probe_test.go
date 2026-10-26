package tutor

import "testing"

func probeCourse() []Topic {
	return []Topic{{Overview, "Overview"}, {"a", "A"}, {"b", "B"}, {"c", "C"}, {"d", "D"}}
}

func TestProbeMarksCorrectTopicsHeldAndSkipsThem(t *testing.T) {
	course := probeCourse()
	progress := map[string]Progress{}
	probe, progress := RecordProbe(nil, course, progress, "probe", "a", Reply{Question: "What does A do?"})
	if probe.Asked != 1 || len(probe.Topics) != 3 || probe.Topics[2] != "c" {
		t.Fatalf("probe should cover the first three topics after the overview: %+v", probe)
	}
	probe, progress = RecordProbe(probe, course, progress, "answer", "a", Reply{Assessment: "none", Question: "What does A do?"})
	if probe.Answered != 0 {
		t.Fatal("a hint must not answer a probe question")
	}
	before := progress
	probe, progress = RecordProbe(probe, course, progress, "answer", "a", Reply{Assessment: "correct"})
	if probe.Answered != 1 || !progress["a"].Held || before["a"].Held {
		t.Fatal("a correct answer should mark the topic held without mutating earlier progress")
	}
	probe, progress = RecordProbe(probe, course, progress, "probe", "b", Reply{Question: "What does B do?"})
	probe, progress = RecordProbe(probe, course, progress, "answer", "b", Reply{Assessment: "partial"})
	if probe.Answered != 2 || progress["b"].Held {
		t.Fatal("a partial answer should leave the topic to be taught")
	}
	if next, found := NextUnheld(course, progress, Overview); !found || next.ID != "b" {
		t.Fatalf("next topic should skip the held topic, got %+v", next)
	}
	if first, found := FirstUnheld(course, progress); !found || first.ID != "b" {
		t.Fatalf("first unheld topic should be b, got %+v", first)
	}
}

func TestIncorrectAnswerClearsHeldAndLessonEndsProbe(t *testing.T) {
	progress := Record(map[string]Progress{"a": {Held: true}}, "a", "answer", Reply{Assessment: "incorrect"})
	if progress["a"].Held {
		t.Fatal("an incorrect answer should clear the held mark")
	}
	probe := &Probe{Topics: []string{"a"}, Asked: 1}
	if left, _ := RecordProbe(probe, probeCourse(), progress, "lesson", "a", Reply{}); left != nil {
		t.Fatal("changing topic should leave the probe")
	}
}

func TestRestartedProbeSkipsHeldTopics(t *testing.T) {
	progress := map[string]Progress{"a": {Held: true}}
	probe, _ := RecordProbe(&Probe{Topics: []string{"a", "b", "c"}, Asked: 3, Answered: 3}, probeCourse(), progress, "probe", "b", Reply{Question: "B?"})
	if probe.Asked != 1 || probe.Topics[0] != "b" || probe.Answered != 0 {
		t.Fatalf("a new probe should start over without held topics: %+v", probe)
	}
}
