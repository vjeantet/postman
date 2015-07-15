package watch

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/etrepat/postman/handler"
	"github.com/etrepat/postman/imap"
	"github.com/etrepat/postman/version"
)

const (
	DELIVERY_MODE_POSTBACK = "postback"
	DELIVERY_MODE_LOGGER   = "logger"
	DELIVERY_MODE_SMART    = "smart"
)

var (
	DefaultLogger  = log.New(os.Stdout, "[watch] ", log.LstdFlags)
	DELIVERY_MODES = map[string]bool{
		DELIVERY_MODE_POSTBACK: true,
		DELIVERY_MODE_LOGGER:   true,
		DELIVERY_MODE_SMART:    true}
)

type Flags struct {
	Host          string
	Port          uint
	Ssl           bool
	Username      string
	Password      string
	Mailbox       string
	Mode          string
	PostbackUrl   string
	PostEncoded   bool
	PostParamName string
}

type Watch struct {
	mailbox  string
	handlers []handler.MessageHandler
	client   *imap.ImapClient
	logger   *log.Logger
	chMsgs   chan string
	done     chan bool
	wg       sync.WaitGroup
}

func (w *Watch) Mailbox() string {
	return w.mailbox
}

func (w *Watch) SetMailbox(value string) {
	w.mailbox = value
}

func (w *Watch) SetLogger(logger *log.Logger) {
	w.logger = logger
}

func (w *Watch) Logger() *log.Logger {
	return w.logger
}

func (w *Watch) AddHandler(handler handler.MessageHandler) {
	w.handlers = append(w.handlers, handler)
}

func (w *Watch) Handlers() []handler.MessageHandler {
	return w.handlers
}

func (w *Watch) Start() {
	w.logger.Println("Starting ", version.VersionShort())

	w.chMsgs = make(chan string, 3)
	w.done = make(chan bool)

	w.wg.Add(1)
	go w.handleIncoming()

	w.logger.Printf("Handling incoming messages with:")
	for i := 0; i < len(w.handlers); i++ {
		w.logger.Printf("> %s", w.handlers[i].Describe())
	}
	w.wg.Add(1)
	err := w.monitorMailbox()
	if err != nil {
		w.logger.Fatalln(err)
	}
}

func (w *Watch) Stop() {
	close(w.done)
	log.Printf("Waiting for termination ==> maximum %d minutes", imap.IdleTimeout/time.Minute)
	w.wg.Wait()

	// Stop close imap connection only when the program enter a waiting state

}

func (w *Watch) handleIncoming() {
	var err error
	var wg sync.WaitGroup
	for message := range w.chMsgs {

		wg.Add(1)
		go func(m string) {
			for _, handler := range w.handlers {
				err = handler.Deliver(m)
				if err != nil {
					w.logger.Println(err)
				} else {
					w.logger.Println("Delivered successfully")
				}
			}
			wg.Done()
		}(message)
	}
	wg.Wait()
	log.Printf("quitting handleIncomming")
	w.wg.Done()
}

func (w *Watch) monitorMailbox() error {
	defer w.wg.Done()

	var err error

	w.logger.Printf("Initiating connection to %s", w.client.Addr())
	err = w.client.Connect()
	if err != nil {
		return err
	}

	defer log.Println("Disconnected from IMAP Server " + w.client.Addr())
	defer w.client.Disconnect()

	w.logger.Printf("Switching to %s", w.mailbox)
	err = w.client.Select(w.mailbox)
	if err != nil {
		return err
	}

	w.logger.Printf("Checking for new (unseen) messages")
	err = w.client.Unseen(w.chMsgs)
	if err != nil {
		return err
	}

L:
	for {
		select {
		case <-w.done:
			close(w.chMsgs)
			log.Printf("closing w.chMsgs")
			break L
		default:
		}
		w.logger.Printf("Waiting for new messages")
		w.client.Incoming(w.chMsgs)
	}

	return nil
}

func NewFlags() *Flags {
	return &Flags{}
}

func New(flags *Flags, handlers ...handler.MessageHandler) *Watch {
	watch := &Watch{
		mailbox: flags.Mailbox,
		client:  imap.NewClient(flags.Host, flags.Port, flags.Ssl, flags.Username, flags.Password),
		logger:  DefaultLogger}

	if len(handlers) != 0 {
		for _, hnd := range handlers {
			watch.AddHandler(hnd)
		}
	} else {
		switch flags.Mode {
		case DELIVERY_MODE_POSTBACK:
			watch.AddHandler(handler.New(handler.POSTBACK_HANDLER, flags.PostbackUrl, flags.PostEncoded, flags.PostParamName))

		case DELIVERY_MODE_LOGGER:
			watch.AddHandler(handler.New(handler.LOGGER_HANDLER, DefaultLogger))
		case DELIVERY_MODE_SMART:
			watch.AddHandler(handler.New(handler.SMART_HANDLER))
		}
	}

	return watch
}

func DeliveryModeValid(mode string) bool {
	return DELIVERY_MODES[mode]
}

func ValidDeliveryModes() []string {
	modes := make([]string, len(DELIVERY_MODES))
	i := 0

	for k, _ := range DELIVERY_MODES {
		modes[i] = k
		i++
	}

	return modes
}
