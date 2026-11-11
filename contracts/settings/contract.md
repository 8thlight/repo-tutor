# Settings contract

Status: draft. Component: `settings`. Issue: none yet.

The settings component turns the bytes of `settings.json` into the learner's settings, and the settings back into bytes. It also holds the defaults and the choices the Settings page offers. It is a pure function: it reads no file, keeps no state, and does not change its arguments. The `session` driver reads and writes the file, as [0220](../../docs/decisions/0220-settings-file.md) says. The choices come from [0230](../../docs/decisions/0230-settings-page.md), [0250](../../docs/decisions/0250-joined-tab-styles.md), [0300](../../docs/decisions/0300-course-flow.md), [0450](../../docs/decisions/0450-exam.md), and [0500](../../docs/decisions/0500-use-the-terminal-colors.md).

## Interface

- `Defaults() -> Values`
- `Choices() -> Choices`
- `Decode(data) -> (Values, Status)`
- `Encode(values) -> bytes`

`Values` holds `TabStyle` (text), `MarkdownOpensRendered` (true or false), `Highlight` (text), `QuizLength` (whole number), and `ExamLength` (whole number). `Status` holds `Unreadable` and `ReadOnly`, both true or false. `Choices` holds the offered tab styles, highlights, quiz lengths, and exam lengths, each as an ordered list.

`Decode` depends only on `data`. `Encode` depends only on `values`. Neither reports an error: a file that cannot be used gives the defaults and sets `Unreadable`.

## Rules

**Defaults.** The tab style is `none`, Markdown opens rendered, the highlight is `green`, the quiz length is 5, and the exam length is 10.

**Choices.** The lists below are offered in this order. An exam length of 0 means every topic read, and it is offered last.

| Setting | Choices, in order |
| -- | -- |
| Tab style | `none`, `block`, `slanted`, `rounded`, `joined_slanted`, `joined_rounded`, `joined_arrows` |
| Highlight | `none`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan` |
| Quiz length | 3, 5, 10 |
| Exam length | 10, 20, 30, 0 |

**One setting at a time.** `Decode` reads each field on its own. A field holds a setting when its name is present and its value is one of the choices for that setting, or `true` or `false` for the Markdown setting. Any other value, including `null`, an empty string, and a number that is not offered, gives that field its default. It does not change any other field.

**Exact values.** A value is compared exactly. `Block` is not `block`, and a value with spaces around it is not offered. A highlight from an earlier release, such as `teal`, `orange`, `dark_blue`, or `purple`, is not offered, so it gives `green`.

**Exact names.** A field name is compared exactly. `Tab_Style` and `QUIZ_LENGTH` are not the names of settings, so each is an unknown field and is ignored.

**A repeated field.** When a name appears more than once in the object, the last value is the one read, and the usual rule for a value then applies to it.

**Other fields.** A field the contract does not name is ignored, in the version the tool reads and in a newer one.

**An unreadable file.** When the bytes are not one JSON object, the result is the defaults with `Unreadable` set and `ReadOnly` clear. That covers an empty file, a truncated file, text that is not JSON, an array, a bare number, and text after the object.

**The version.** A file with no `version`, or a `version` of 1 or less, is read and may be written over. A file whose `version` is greater than 1 is read for the fields above, with the same per-field fallbacks, and has `ReadOnly` set. The tool never writes over it.

**Encoding.** `Encode` writes one JSON object. It holds `version` with the value 1, then `tab_style`, `markdown_opens_rendered`, `highlight`, `quiz_length`, and `exam_length`, always all of them and in that order. It is indented with two spaces, has one field to a line, and ends with a newline, so a learner who edits the file by hand sees a short, steady layout. Decoding the result gives the same values and a clear status.

## Side effects and data access

- Side effects: none.
- Data access: `Decode` creates `Values` and `Status` from `data` and does not change it. `Encode` creates bytes from `values`. `Defaults` and `Choices` create their results from nothing.

## Decisions to know

| Decision | Reversal |
| -- | -- |
| A bad value for one setting gives that setting's default and keeps the rest, instead of rejecting the file. 0220 decides this | hard |
| An unreadable file gives the defaults and a flag, and the driver shows the notice. 0220 decides this | hard |
| A file from a newer version is read and marked read-only, so a downgrade never loses fields | hard |
| Values and field names are matched exactly, with no change of case and no trimming | cheap |
| A repeated field takes the last value, because JSON does not define it | cheap |
| `null` counts as a value that is not offered, so it gives the default | cheap |
| The written file has a fixed layout: field order, two-space indent, and a final newline | cheap |
| An exam length of 0 means every topic read, and it is stored as 0 | cheap |
| A negative `version` counts as an older file, not a newer one | cheap |

## Properties

These hold for every input the contract tests use.

- **Pure.** `Decode` leaves its input bytes as they were. `Encode` does not change its argument.
- **No memory.** The same call gives the same result on a shared instance, whatever calls came before, in forward or reverse order.
- **Deterministic.** Two calls with the same input give identical values and identical bytes.
- **Round trip.** `Decode(Encode(values))` returns `values` with a clear status, for every value the cases name.
- **Declared data access.** `Decode` sets only the five fields of `Values` and the two of `Status`.

## Not covered

- A field of the wrong type, such as `{"quiz_length": "5"}` or `{"version": "1"}`. The current decoder treats the whole file as unreadable. A person chose to leave this open.
- The choice lists as shared package variables. A caller that changes one changes it for every later reader. A person chose to leave this open.
- A byte order mark before the object. The current decoder treats it as unreadable. No decision exists.
- Values passed to `Encode` that are not choices. The caller gives offered values, and the file written for others is not defined.
- A top-level `null`. The current decoder returns the defaults with a clear status. No decision exists.
- Spaces or line breaks around the object in the input. They are valid JSON and the current decoder accepts them. Cases do not use them.
- The notices the learner sees, the file location, and the temporary-file write. They belong to the `session` driver and the Settings page.

## Wrong implementations tried

Each row names a plausible wrong implementation and the tests that fail it. The wrong builds were scratch copies and are not committed.

| Wrong implementation | Caught by |
| -- | -- |
| Ignores the version | `version-2` and `version-100` cases |
| Marks version 1 read-only | `version-1` case |
| Marks every version but 1 read-only, including 0 and negative | `version-0` and `a-negative-version` cases |
| One unknown tab style resets every field | `one-unknown-value-leaves-the-other-fields-alone` |
| Accepts any positive quiz length | `quiz-length-4`, `-3`, and `100` cases |
| Rejects an exam length of 0 | `exam-length-0` case |
| An unreadable file gives zero values instead of defaults | every `unreadable` case |
| An unreadable file also sets read-only | every `unreadable` case |
| Leaves the exam length out of the file | The one-field-per-setting test |
| Writes version 2 | The one-field-per-setting test |
| Accepts any highlight string | `an-unknown-highlight` and the earlier-highlight cases |
| Defaults the highlight to none, or Markdown to source | The defaults test and the empty-object case |
| An unknown field makes the file unreadable | `an-unknown-field` cases |
| Keeps the last result between calls | The no-memory test |
| Matches values without regard to case | `a-value-is-matched-exactly` case |
| Offers the highlights in another order | The choices test |
| Changes its input bytes | The input-unchanged test |
| Writes camel case field names | The one-field-per-setting test and every decode case |
| Drops a false Markdown setting | The one-field-per-setting test |
| Copies the exam length into the quiz length | The one-field-per-setting test |
| Writes compact JSON, indents with four spaces, puts the version last, or leaves out the final newline | The by-hand layout test |
| Reads field names without regard to case | `a-field-name-is-matched-exactly` and `an-upper-case-field-name-is-an-unknown-field` cases |
| Takes the first value of a repeated field | The two repeated-field cases |
