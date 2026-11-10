package about

type Access string

const (
	ReadOnly  Access = "read only"
	ReadWrite Access = "read and write"
)

type Place struct {
	Name   string
	Path   string
	Holds  string
	Access Access
}

const AgentHistoryNote = "Your agent CLI keeps its own conversation history. repo-tutor does not read or write it."
