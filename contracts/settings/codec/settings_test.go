package codec_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	contract "repo-tutor/contracts/settings/codec"
)

type fileValues struct {
	TabStyle              string `json:"tab_style"`
	MarkdownOpensRendered bool   `json:"markdown_opens_rendered"`
	Highlight             string `json:"highlight"`
	QuizLength            int    `json:"quiz_length"`
	ExamLength            int    `json:"exam_length"`
}

func (file fileValues) values() contract.Values {
	return contract.Values{TabStyle: file.TabStyle, MarkdownOpensRendered: file.MarkdownOpensRendered, Highlight: file.Highlight, QuizLength: file.QuizLength, ExamLength: file.ExamLength}
}

type decodeCase struct {
	Name   string  `json:"name"`
	Source string  `json:"source"`
	Input  *string `json:"input"`
	Expect *struct {
		Values *fileValues `json:"values"`
		Status *struct {
			Unreadable bool `json:"unreadable"`
			ReadOnly   bool `json:"read_only"`
		} `json:"status"`
	} `json:"expect"`
}

func (one decodeCase) status() contract.Status {
	return contract.Status{Unreadable: one.Expect.Status.Unreadable, ReadOnly: one.Expect.Status.ReadOnly}
}

type encodeCase struct {
	Name   string     `json:"name"`
	Source string     `json:"source"`
	Values fileValues `json:"values"`
}

type manifest struct {
	Defaults struct {
		Source string     `json:"source"`
		Values fileValues `json:"values"`
	} `json:"defaults"`
	Choices struct {
		Source     string   `json:"source"`
		TabStyle   []string `json:"tab_style"`
		Highlight  []string `json:"highlight"`
		QuizLength []int    `json:"quiz_length"`
		ExamLength []int    `json:"exam_length"`
	} `json:"choices"`
	Decode []decodeCase `json:"decode"`
	Encode []encodeCase `json:"encode"`
}

func load(t *testing.T) manifest {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var loaded manifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loaded); err != nil {
		t.Fatal(err)
	}
	return loaded
}

func TestEveryCaseExpectsSomething(t *testing.T) {
	loaded := load(t)
	names := map[string]bool{}
	for _, one := range loaded.Decode {
		if one.Name == "" || names[one.Name] || one.Source == "" || one.Input == nil || one.Expect == nil || one.Expect.Values == nil || one.Expect.Status == nil {
			t.Fatalf("decode case %q is incomplete or repeated", one.Name)
		}
		names[one.Name] = true
	}
	for _, one := range loaded.Encode {
		if one.Name == "" || names["encode:"+one.Name] || one.Source == "" {
			t.Fatalf("encode case %q is incomplete or repeated", one.Name)
		}
		names["encode:"+one.Name] = true
	}
}

func TestEveryOfferedValueHasACaseThatAcceptsIt(t *testing.T) {
	loaded := load(t)
	accepts := func(field string, matches func(contract.Values) bool) bool {
		for _, one := range loaded.Decode {
			if strings.Contains(*one.Input, `"`+field+`"`) && matches(one.Expect.Values.values()) && !one.Expect.Status.Unreadable {
				return true
			}
		}
		return false
	}
	for _, style := range loaded.Choices.TabStyle {
		if !accepts("tab_style", func(values contract.Values) bool { return values.TabStyle == style }) {
			t.Errorf("no case accepts tab style %q", style)
		}
	}
	for _, color := range loaded.Choices.Highlight {
		if !accepts("highlight", func(values contract.Values) bool { return values.Highlight == color }) {
			t.Errorf("no case accepts highlight %q", color)
		}
	}
	for _, length := range loaded.Choices.QuizLength {
		if !accepts("quiz_length", func(values contract.Values) bool { return values.QuizLength == length }) {
			t.Errorf("no case accepts quiz length %d", length)
		}
	}
	for _, length := range loaded.Choices.ExamLength {
		if !accepts("exam_length", func(values contract.Values) bool { return values.ExamLength == length }) {
			t.Errorf("no case accepts exam length %d", length)
		}
	}
}

func TestDefaults(t *testing.T) {
	loaded := load(t)
	if got := newCodec().Defaults(); got != loaded.Defaults.Values.values() {
		t.Fatalf("got %+v, want %+v", got, loaded.Defaults.Values.values())
	}
}

func TestChoicesAreOfferedInOrder(t *testing.T) {
	loaded := load(t)
	got := newCodec().Choices()
	want := contract.Choices{TabStyles: loaded.Choices.TabStyle, Highlights: loaded.Choices.Highlight, QuizLengths: loaded.Choices.QuizLength, ExamLengths: loaded.Choices.ExamLength}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestDecode(t *testing.T) {
	for _, one := range load(t).Decode {
		t.Run(one.Name, func(t *testing.T) {
			values, status := newCodec().Decode([]byte(*one.Input))
			if values != one.Expect.Values.values() || status != one.status() {
				t.Fatalf("input %q: got %+v %+v, want %+v %+v", *one.Input, values, status, one.Expect.Values.values(), one.status())
			}
		})
	}
}

func fileOf(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("the encoded settings are not a JSON object: %v\n%s", err, data)
	}
	return decoded
}

func TestEncodeWritesTheVersionAndOneFieldPerSetting(t *testing.T) {
	for _, one := range load(t).Encode {
		t.Run(one.Name, func(t *testing.T) {
			decoded := fileOf(t, newCodec().Encode(one.Values.values()))
			want := map[string]any{
				"version": float64(1), "tab_style": one.Values.TabStyle, "markdown_opens_rendered": one.Values.MarkdownOpensRendered,
				"highlight": one.Values.Highlight, "quiz_length": float64(one.Values.QuizLength), "exam_length": float64(one.Values.ExamLength),
			}
			if !reflect.DeepEqual(decoded, want) {
				t.Fatalf("got %v, want %v", decoded, want)
			}
		})
	}
}

func TestEncodeWritesAFileThatIsEasyToEditByHand(t *testing.T) {
	loaded := load(t)
	golden := map[string]string{"the-defaults": "testdata/defaults.settings.json", "every-field-changed": "testdata/changed.settings.json"}
	for _, one := range loaded.Encode {
		name, found := golden[one.Name]
		if !found {
			continue
		}
		want, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if got := newCodec().Encode(one.Values.values()); !bytes.Equal(got, want) {
			t.Fatalf("%s: got\n%s\nwant\n%s", one.Name, got, want)
		}
		delete(golden, one.Name)
	}
	if len(golden) != 0 {
		t.Fatalf("no encode case for %v", golden)
	}
}

func TestEncodedSettingsDecodeToThemselves(t *testing.T) {
	loaded := load(t)
	codec := newCodec()
	var all []contract.Values
	for _, one := range loaded.Encode {
		all = append(all, one.Values.values())
	}
	for _, one := range loaded.Decode {
		all = append(all, one.Expect.Values.values())
	}
	for _, want := range all {
		if got, status := codec.Decode(codec.Encode(want)); got != want || status != (contract.Status{}) {
			t.Errorf("%+v came back as %+v %+v", want, got, status)
		}
	}
}

func TestDecodeLeavesItsInputAlone(t *testing.T) {
	codec := newCodec()
	for _, one := range load(t).Decode {
		input := []byte(*one.Input)
		before := slices.Clone(input)
		codec.Decode(input)
		if !bytes.Equal(input, before) {
			t.Fatalf("%s: the input changed from %q to %q", one.Name, before, input)
		}
	}
}

func TestTheSameCallGivesTheSameResultWhateverCameBefore(t *testing.T) {
	loaded := load(t)
	run := func(codec contract.Codec, order []decodeCase) map[string][2]any {
		results := map[string][2]any{}
		for _, one := range order {
			values, status := codec.Decode([]byte(*one.Input))
			results[one.Name] = [2]any{values, status}
		}
		return results
	}
	reversed := slices.Clone(loaded.Decode)
	slices.Reverse(reversed)
	shared := newCodec()
	forward := run(shared, loaded.Decode)
	backward := run(shared, reversed)
	fresh := run(newCodec(), loaded.Decode)
	if !reflect.DeepEqual(forward, backward) || !reflect.DeepEqual(forward, fresh) {
		t.Fatal("results depend on which calls came before")
	}
}

func TestTheSameInputGivesIdenticalBytesAndValues(t *testing.T) {
	codec := newCodec()
	for _, one := range load(t).Encode {
		if first, second := codec.Encode(one.Values.values()), codec.Encode(one.Values.values()); !bytes.Equal(first, second) {
			t.Fatalf("%s: two calls gave different bytes", one.Name)
		}
	}
	if codec.Defaults() != codec.Defaults() {
		t.Fatal("the defaults differ between calls")
	}
}
