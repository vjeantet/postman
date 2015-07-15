package handler

import (
	"bytes"
	"fmt"
	"log"
	"net/mail"

	"github.com/tbruyelle/hipchat-go/hipchat"
	"github.com/vjeantet/go.enmime"
)

//HipChatHandler struct
type HipChatHandler struct {
}

//Deliver handles hipchat delivery
func (hnd *HipChatHandler) Deliver(message string) error {
	mailMessage, _ := mail.ReadMessage(bytes.NewBufferString(message))
	mime, _ := enmime.ParseMIMEBody(mailMessage)
	// 	s := `
	// De    : %s
	// Sujet : %s
	// Text  : %d chars
	// Html  : %d chars
	// Inlines      : %d
	// Attachements : %d
	// Others       : %d`
	// 	log.Printf(s,
	// 		mime.GetHeader("From"),
	// 		mime.GetHeader("Subject"),
	// 		len(mime.Text),
	// 		len(mime.Html),
	// 		len(mime.Inlines),
	// 		len(mime.Attachments),
	// 		len(mime.OtherParts),
	// 	)

	sendHipChat(mime)

	return nil
}

func sendHipChat(mime *enmime.MIMEBody) {

	roomAuth := "IEV0ODUXMpDXprgA0aeCCIfOhqzQGg04dT2S882N"
	roomName := "testroom5"
	roomColor := "green"

	s := `
De    : %s
Sujet : %s
Text  : %d chars
Html  : %d chars
Inlines      : %d
Attachements : %d
Others       : %d`

	message := fmt.Sprintf(s,
		mime.GetHeader("From"),
		mime.GetHeader("Subject"),
		len(mime.Text),
		len(mime.Html),
		len(mime.Inlines),
		len(mime.Attachments),
		len(mime.OtherParts),
	)

	log.Println(message)

	c := hipchat.NewClient(roomAuth)

	//If specify html, need to determine/format the escape characters
	notifRq := &hipchat.NotificationRequest{Color: roomColor, Message: message, MessageFormat: "text"}

	_, err := c.Room.Notification(roomName, notifRq)
	if err != nil {
		panic(err)
	}

}

//Describe the handler
func (hnd *HipChatHandler) Describe() string {
	return "HipChat Handler"
}

//NewHipChatHandler create the handler
func NewHipChatHandler() *HipChatHandler {
	return &HipChatHandler{}
}
