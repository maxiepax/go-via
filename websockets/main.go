package websockets

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
)

type LogServer struct {
	Hook *hook

	// subscriberMessageBuffer controls the max number
	// of messages that can be queued for a subscriber
	// before it is kicked.
	//
	// Defaults to 16.
	subscriberMessageBuffer int

	subscribersMu sync.Mutex
	subscribers   map[*subscriber]struct{}

	historyMu sync.Mutex
	history   [][]byte
}

type subscriber struct {
	msgs      chan []byte
	closeSlow func()
}

type hook struct {
	formatter logrus.Formatter
	ls        *LogServer
}

func (hook *hook) Fire(entry *logrus.Entry) error {
	json, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	ls := hook.ls

	ls.subscribersMu.Lock()
	defer ls.subscribersMu.Unlock()

	for s := range ls.subscribers {
		select {
		case s.msgs <- json:
		default:
			go s.closeSlow()
		}
	}

	ls.historyMu.Lock()
	if len(ls.history) < 50 {
		ls.history = append(ls.history, json)
	} else {
		ls.history = append(ls.history[1:], json)
	}
	ls.historyMu.Unlock()

	return nil
}

// Levels define on which log levels this hook would trigger
func (hook *hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

func NewLogServer() *LogServer {
	ls := &LogServer{
		subscriberMessageBuffer: 16,
		subscribers:             make(map[*subscriber]struct{}),
		history:                 make([][]byte, 0),
	}
	ls.Hook = &hook{
		formatter: &logrus.JSONFormatter{},
		ls:        ls,
	}

	return ls
}

func (ls *LogServer) Handle(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not accept websocket")
		return
	}
	defer func() {
		err := conn.Close(websocket.StatusInternalError, "")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Warn("could not close websocket connection")
		}
	}()

	err = ls.subscribe(c.Request.Context(), conn)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("websocket was closed unexpected")
		return
	}
}

func (ls *LogServer) subscribe(ctx context.Context, c *websocket.Conn) error {
	ctx = c.CloseRead(ctx)

	s := &subscriber{
		msgs: make(chan []byte, ls.subscriberMessageBuffer),
		closeSlow: func() {
			err := c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err": err,
				}).Warn("could not close websocket connection")
			}
		},
	}

	ls.historyMu.Lock()
	history := ls.history
	ls.historyMu.Unlock()

	for _, msg := range history {
		err := writeTimeout(ctx, time.Second*5, c, msg)
		if err != nil {
			return err
		}
	}

	ls.addSubscriber(s)
	defer ls.deleteSubscriber(s)

	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

func (ls *LogServer) addSubscriber(s *subscriber) {
	ls.subscribersMu.Lock()
	ls.subscribers[s] = struct{}{}
	ls.subscribersMu.Unlock()
}

func (ls *LogServer) deleteSubscriber(s *subscriber) {
	ls.subscribersMu.Lock()
	delete(ls.subscribers, s)
	ls.subscribersMu.Unlock()
}
