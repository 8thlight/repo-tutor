package codec_test

import (
	contract "repo-tutor/contracts/settings/codec"
	implementation "repo-tutor/internal/settings/codec"
)

type adapter struct{}

func newCodec() contract.Codec { return adapter{} }

func (adapter) Defaults() contract.Values { return values(implementation.Defaults()) }

func (adapter) Choices() contract.Choices {
	return contract.Choices{
		TabStyles:   implementation.TabStyles,
		Highlights:  implementation.Highlights,
		QuizLengths: implementation.QuizLengths,
		ExamLengths: implementation.ExamLengths,
	}
}

func (adapter) Decode(data []byte) (contract.Values, contract.Status) {
	decoded, status := implementation.Decode(data)
	return values(decoded), contract.Status{Unreadable: status.Unreadable, ReadOnly: status.ReadOnly}
}

func (adapter) Encode(wanted contract.Values) []byte {
	return implementation.Encode(implementation.Settings{
		TabStyle: wanted.TabStyle, MarkdownOpensRendered: wanted.MarkdownOpensRendered,
		Highlight: wanted.Highlight, QuizLength: wanted.QuizLength, ExamLength: wanted.ExamLength,
	})
}

func values(decoded implementation.Settings) contract.Values {
	return contract.Values{
		TabStyle: decoded.TabStyle, MarkdownOpensRendered: decoded.MarkdownOpensRendered,
		Highlight: decoded.Highlight, QuizLength: decoded.QuizLength, ExamLength: decoded.ExamLength,
	}
}
