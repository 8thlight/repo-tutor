package settings_test

import (
	"testing"

	"repo-tutor/internal/settings"
)

func TestAnEmptyObjectGivesDefaults(t *testing.T) {
	got, status := settings.Decode([]byte(`{}`))
	if got != settings.Defaults() || status != (settings.Status{}) {
		t.Fatalf("got %+v %+v", got, status)
	}
}

func TestKnownValuesAreRead(t *testing.T) {
	got, _ := settings.Decode([]byte(`{"version":1,"tab_style":"rounded","markdown_opens_rendered":false}`))
	if got.TabStyle != "rounded" || got.MarkdownOpensRendered {
		t.Fatalf("got %+v", got)
	}
}

func TestAnUnknownValueUsesThatFieldsDefault(t *testing.T) {
	got, status := settings.Decode([]byte(`{"tab_style":"zigzag","highlight":"pink","markdown_opens_rendered":false}`))
	if got.TabStyle != settings.TabStyleBlock || got.Highlight != "green" || got.MarkdownOpensRendered || status.Unreadable {
		t.Fatalf("got %+v %+v", got, status)
	}
}

func TestUnknownFieldsAreIgnored(t *testing.T) {
	got, status := settings.Decode([]byte(`{"version":1,"font":"mono"}`))
	if got != settings.Defaults() || status != (settings.Status{}) {
		t.Fatalf("got %+v %+v", got, status)
	}
}

func TestInvalidJSONIsUnreadableAndGivesDefaults(t *testing.T) {
	got, status := settings.Decode([]byte(`{"tab_style":`))
	if got != settings.Defaults() || !status.Unreadable {
		t.Fatalf("got %+v %+v", got, status)
	}
}

func TestANewerVersionIsReadButReadOnly(t *testing.T) {
	got, status := settings.Decode([]byte(`{"version":2,"tab_style":"none"}`))
	if got.TabStyle != "none" || !status.ReadOnly {
		t.Fatalf("got %+v %+v", got, status)
	}
}

func TestEncodedSettingsDecodeToThemselves(t *testing.T) {
	want := settings.Settings{TabStyle: "slanted", MarkdownOpensRendered: false, Highlight: "dark_blue"}
	got, status := settings.Decode(settings.Encode(want))
	if got != want || status != (settings.Status{}) {
		t.Fatalf("got %+v %+v", got, status)
	}
}

func TestAnEarlierTealValueFallsBackToGreen(t *testing.T) {
	got, status := settings.Decode([]byte(`{"version":1,"highlight":"teal"}`))
	if got.Highlight != "green" || status.Unreadable {
		t.Fatalf("got %+v %+v", got, status)
	}
}

func TestEveryJoinedStyleIsAKnownValue(t *testing.T) {
	for _, style := range []string{"joined_slanted", "joined_rounded", "joined_arrows"} {
		got, _ := settings.Decode([]byte(`{"tab_style":"` + style + `"}`))
		if got.TabStyle != style {
			t.Fatalf("%s should be accepted, got %s", style, got.TabStyle)
		}
	}
}
