package notify

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"github.com/tmnhs/common/logger"
)

const (
	NotifyTypeMail    = 1
	NotifyTypeWebHook = 2
)

var _defaultMail *Mail

type Mail struct {
	Port     int
	From     string
	Host     string
	Secret   string
	Nickname string
}

func (mail *Mail) SendMsg(msg *Message) {
	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(_defaultMail.From, _defaultMail.Nickname)) //这种方式可以添加别名，即“XX官方”
	m.SetHeader("To", msg.To...)
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/html", msg.Body)

	d := gomail.NewDialer(_defaultMail.Host, _defaultMail.Port, _defaultMail.From, _defaultMail.Secret)
	if err := d.DialAndSend(m); err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("smtp send msg[%+v] err: %s", msg, err.Error()))
	}
}
