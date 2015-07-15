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
	RoomAuth  string
	RoomName  string
	RoomColor string
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

	sendHipChat(mime, hnd)

	return nil
}

func sendHipChat(mime *enmime.MIMEBody, hnd *HipChatHandler) {

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

	c := hipchat.NewClient(hnd.RoomAuth)

	//If specify html, need to determine/format the escape characters
	notifRq := &hipchat.NotificationRequest{Color: hnd.RoomColor, Message: message, MessageFormat: "text"}

	_, err := c.Room.Notification(hnd.RoomName, notifRq)
	if err != nil {
		panic(err)
	}

}

//Describe the handler
func (hnd *HipChatHandler) Describe() string {
	return "HipChat Handler"
}

//NewHipChatHandler create the handler
func NewHipChatHandler(roomAuth string, roomName string, roomColor string) *HipChatHandler {
	return &HipChatHandler{
		RoomAuth:  roomAuth,
		RoomName:  roomName,
		RoomColor: roomColor}
}
