package notify

import (
	"encoding/json"
	"fmt"
	"github.com/tmnhs/common/httpclient"
	"github.com/tmnhs/common/logger"
)

type WebHook struct {
	Kind string
	Url  string
}

var _defaultWebHook *WebHook

func (w *WebHook) SendMsg(msg *Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	_, err = httpclient.PostJson(_defaultWebHook.Url, string(b), 0)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("web hook api send msg[%+v] err: %s", msg, err.Error()))
	}
}
