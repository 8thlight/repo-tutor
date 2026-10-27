package terminal

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"repo-tutor/internal/repository"
	"repo-tutor/internal/tutor"
)

func hotspotFixture() interfaceModel {
	model := interfaceFixture()
	model.application.snapshot.Hotspots = []repository.Hotspot{{Path: "main.go", Line: 2, Language: "go", Commits: 3, Branches: 4, Lines: 10, Score: 20}}
	model.application.course = tutor.Course(model.application.snapshot)
	return model
}

func TestChoosingTheHotspotsTopicInTheListStartsItsLesson(t *testing.T) {
	model := hotspotFixture()
	started, _ := model.openTopic(tutor.HotspotsTopic)
	opened := started.(interfaceModel)
	defer opened.cancel()
	if !opened.busy || opened.intent.topic != tutor.HotspotsTopic || opened.intent.action != "lesson" {
		t.Fatalf("the topic should start its lesson: %+v", opened.intent)
	}
}

func TestATextLineThatStartsWithASlashIsAnOrdinaryQuestion(t *testing.T) {
	model := readingFixture()
	model.input.SetValue("/topics what is this?")
	started, _ := model.submit()
	asked := started.(interfaceModel)
	defer asked.cancel()
	if !asked.busy || asked.intent.action != "question" || asked.intent.message != "/topics what is this?" {
		t.Fatalf("slash text should be sent as a question: %+v", asked.intent)
	}
	if asked.page == "topics" {
		t.Fatal("the interface no longer parses commands")
	}
}

func TestControlOOpensNothingAnymore(t *testing.T) {
	model := hotspotFixture()
	next, _ := press(model, 'o', tea.ModCtrl)
	if next.page != "lesson" {
		t.Fatalf("Ctrl+O should no longer open a page, got %s", next.page)
	}
}

func TestTheTopicsListOffersTheHotspotsTopic(t *testing.T) {
	model := hotspotFixture()
	opened, _ := model.openPicker("topics", "")
	found := false
	for _, item := range opened.(interfaceModel).picker.Items() {
		if entry, ok := item.(browserItem); ok && entry.value == tutor.HotspotsTopic {
			found = true
		}
	}
	if !found {
		t.Fatal("the Topics list should include Where change happens")
	}
}
