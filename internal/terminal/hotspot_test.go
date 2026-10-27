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

func TestSlashHotspotsOpensTheHotspotsLesson(t *testing.T) {
	model := hotspotFixture()
	model.input.SetValue("/hotspots")
	next, command := model.submit()
	opened := next.(interfaceModel)
	if command == nil && opened.page != "lesson" {
		t.Fatalf("/hotspots should open the lesson, got page %s", opened.page)
	}
	if opened.intent.topic != tutor.HotspotsTopic && opened.application.state.Topic != tutor.HotspotsTopic {
		t.Fatalf("the lesson should be about the hotspots topic: %+v", opened.application.state.Topic)
	}
	if opened.cancel != nil {
		opened.cancel()
	}
}

func TestSlashHotspotsExplainsWhenThereAreNone(t *testing.T) {
	model := interfaceFixture()
	model.application.snapshot.Hotspots = nil
	model.application.course = tutor.Course(model.application.snapshot)
	model.input.SetValue("/hotspots")
	next, _ := model.submit()
	if next.(interfaceModel).notice != "This repository has no hotspots to teach." {
		t.Fatalf("got %q", next.(interfaceModel).notice)
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
