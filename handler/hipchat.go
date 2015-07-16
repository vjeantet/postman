package handler

import (
	"bytes"
	"fmt"
	"log"
	"net/mail"
	"strings"

	"github.com/jaytaylor/html2text"
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

	return sendHipChat(mime, hnd)
}

//short truncate message, note not unicode compliant
func short(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i])
	}
	return s
}

//sendHipChat transform email to message, log and send to hipchat room
func sendHipChat(mime *enmime.MIMEBody, hnd *HipChatHandler) error {

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

	//log general message information
	log.Println(message)

	messageFormat := "text"

	s = `
  From     : %s
  Subject  : %s
  %s`

	message = fmt.Sprintf(s,
		mime.GetHeader("From"),
		mime.GetHeader("Subject"),
		formatMessage(mime.Text),
	)

	//need to truncate message to 10000, supported by hipchat api
	message = short(message, 10000)

	//log what sending to hipchat
	log.Println(message)

	c := hipchat.NewClient(hnd.RoomAuth)

	//If specify html, need to determine/format the escape characters
	notifRq := &hipchat.NotificationRequest{Color: hnd.RoomColor, Message: message, MessageFormat: messageFormat}

	_, err := c.Room.Notification(hnd.RoomName, notifRq)
	if err != nil {
		log.Println("failed to send to hipchat: " + err.Error())
		return err
	}

	return nil
}

//formatMessage make into text if contains html
func formatMessage(message string) string {

	isHTML := strings.Contains(message, "<html>")

	if isHTML {
		text, err := html2text.FromString(message)
		if err != nil {
			log.Println("failed to convert to text " + err.Error())
		}
		return text
	}

	return message
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
