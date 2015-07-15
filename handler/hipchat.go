package handler

import (
	"bytes"
	"log"
	"net/mail"

	"github.com/vjeantet/go.enmime"
)

type HipChatHandler struct {
}

func (hnd *HipChatHandler) Deliver(message string) error {
	mailMessage, _ := mail.ReadMessage(bytes.NewBufferString(message))
	mime, _ := enmime.ParseMIMEBody(mailMessage)
	s := `
De    : %s
Sujet : %s
Text  : %d chars
Html  : %d chars
Inlines      : %d
Attachements : %d
Others       : %d`
	log.Printf(s,
		mime.GetHeader("From"),
		mime.GetHeader("Subject"),
		len(mime.Text),
		len(mime.Html),
		len(mime.Inlines),
		len(mime.Attachments),
		len(mime.OtherParts),
	)

	return nil
}

func (hnd *HipChatHandler) Describe() string {
	return "HipChat Handler"
}

func NewHipChatHandler() *HipChatHandler {
	return &HipChatHandler{}
}
