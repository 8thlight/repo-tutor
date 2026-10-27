package terminal

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

var noticeLifetime = 10 * time.Second

type noticeExpired struct{ generation int }

func (model interfaceModel) noticeSticky() bool {
	return model.errorText != "" || model.skipArmed || model.leaveArmed
}

func (model interfaceModel) expireNotice(before interfaceModel) (interfaceModel, tea.Cmd) {
	if model.notice == "" || model.notice == before.notice || model.noticeSticky() {
		return model, nil
	}
	model.noticeGeneration++
	generation := model.noticeGeneration
	return model, tea.Tick(noticeLifetime, func(time.Time) tea.Msg { return noticeExpired{generation} })
}
