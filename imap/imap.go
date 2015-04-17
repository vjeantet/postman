package imap

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mxk/go-imap/imap"
)

const (
	IdleTimeout = 3 * time.Minute
)

var (
	DefaultLogger  = log.New(os.Stderr, "[imap ] ", log.LstdFlags)
	DefaultLogMask = imap.LogConn | imap.LogCmd
)

type ImapClient struct {
	client *imap.Client

	Host     string
	Port     uint
	Ssl      bool
	Username string
	Password string
}

func (c *ImapClient) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *ImapClient) Connect() error {
	var err error

	if c.Port == 993 || c.Ssl == true {
		c.client, err = imap.DialTLS(c.Addr(), &tls.Config{})
	} else {
		c.client, err = imap.Dial(c.Addr())
	}

	if err != nil {
		return fmt.Errorf("IMAP dial error! ", err)
	}

	if c.client.Caps["STARTTLS"] {
		_, err = imap.Wait(c.client.StartTLS(nil))
	}

	if err != nil {
		return fmt.Errorf("Could not stablish TLS encrypted connection. ", err)
	}

	if c.client.Caps["ID"] {
		_, err = imap.Wait(c.client.ID("name", "go-postman"))
	}

	c.client.SetLogMask(imap.LogConn)
	_, err = imap.Wait(c.client.Login(c.Username, c.Password))
	if err != nil {
		return fmt.Errorf("IMAP authentication failed! Invalid credentials.")
	}
	c.client.SetLogMask(imap.DefaultLogMask)

	return err
}

func (c *ImapClient) Disconnect() {
	imap.Wait(c.client.Logout(30 * time.Second))
	c.client.Close(true)
}

func (c *ImapClient) Select(mailbox string) error {
	_, err := imap.Wait(c.client.Select(mailbox, false))

	if err != nil {
		return fmt.Errorf("Failed to switch to mailbox %s", mailbox)
	}

	return err
}

func (c *ImapClient) Unseen(chMsg chan string) (err error) {
	var ids []uint32

	ids, err = c.query("UNSEEN")
	if err != nil {
		return err
	}

	// Récupérer les messages un par un
	// for _, id := range ids {
	// 	err = c.messagesForIds([]uint32{id}, chMsg)
	// }

	// Récupérer les messages par lot d'une taille équivalente à la capacité du channel
	size := cap(chMsg)
	if size == 0 {
		size = 1
	}
	for i := 0; i <= len(ids)/size; i++ {
		start := i * size
		end := start + size
		if end > len(ids) {
			end = len(ids)
		}
		err = c.messagesForIds(ids[start:end], chMsg)
	}

	if err != nil {
		return err
	}

	return err
}

func (c *ImapClient) Incoming(chMsg chan string) (err error) {
	err = c.waitForIncoming()
	if err != nil {
		return err
	}

	for _, resp := range c.client.Data {
		switch resp.Label {
		case "EXISTS":
			return c.Unseen(chMsg)
		case "FETCH":
			return c.Unseen(chMsg)
		}

	}

	return err
}

func (c *ImapClient) query(arguments ...string) ([]uint32, error) {
	args := []imap.Field{}
	for _, a := range arguments {
		args = append(args, a)
	}

	cmd, err := imap.Wait(c.client.Search(args...))
	if err != nil {
		return nil, fmt.Errorf("An error ocurred while searching for messages. ", err)
	}

	return cmd.Data[0].SearchResults(), nil
}

func (c *ImapClient) messagesForIds(ids []uint32, chMsg chan string) error {

	if len(ids) > 0 {
		set, _ := imap.NewSeqSet("")
		set.AddNum(ids...)

		cmd, err := imap.Wait(c.client.Fetch(set, "RFC822"))
		if err != nil {
			return fmt.Errorf("An error ocurred while fetching unread messages data. ", err)
		}

		for _, msg := range cmd.Data {
			attrs := msg.MessageInfo().Attrs
			chMsg <- imap.AsString(attrs["RFC822"])
		}
	}

	return nil
}

func (c *ImapClient) waitForIncoming() (err error) {
	_, err = c.client.Idle()
	if err != nil {
		return fmt.Errorf("Could not start IDLE process. ", err)
	}

	err = c.client.Recv(IdleTimeout)
	if err != nil && err != imap.ErrTimeout {
		return fmt.Errorf("Some error ocurred while IDLING: %q", err)
	}

	_, err = imap.Wait(c.client.IdleTerm())
	if err != nil {
		return fmt.Errorf("IDLE command termination failed for some reason. ", err)
	}

	return err
}

func init() {
	imap.DefaultLogger = DefaultLogger
	imap.DefaultLogMask = DefaultLogMask
}

func NewClient(host string, port uint, ssl bool, username string, password string) *ImapClient {
	return &ImapClient{
		Host:     host,
		Port:     port,
		Ssl:      ssl,
		Username: username,
		Password: password}
}
