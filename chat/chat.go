package main

import (
	"container/list"
	"log"
	"net/http"
	"time"

	"github.com/googollee/go-socket.io"
)

type Event struct {
	EvtType string
	User string
	Timestamp int
	Text string
}

type Subscription struct {
	Archive []Event
	New <-chan Event
}

func NewEvent(evtType, user, msg string) Event {
	return Event{evtType, user, int(time.Now().Unix()), msg}
}

var (
	subscribe = make(chan (chan<- Subscription), 10)
	unsubscribe = make(chan (<-chan Event), 10)
	publish = make(chan Event, 10)
)

func Subscribe() Subscription {
	c := make(chan Subscription)
	subscribe <- c
	return <- c
}

func (s Subscription) Cancel() {
	unsubscribe <- s.New

	for {
		select {
		case _, ok := <- s.New:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

func Join(user string) {
	publish <- NewEvent("join", user, "")
}

func Say(user, message string) {
	publish <- NewEvent("message", user, message)
}

func Leave(user string) {
	publish <- NewEvent("leave", user, "")
}

func Chatroom() {
	archive := list.New()
	subscribers := list.New()

	for {
		select {
		case c := <-subscribe: // when new user comes..
			var events []Event

			for e := archive.Front(); e != nil; e = e.Next() { // if there are events
				events = append(events, e.Value.(Event))
			}

			subscriber := make(chan Event, 10)
			subscribers.PushBack(subscriber)

			c <- Subscription{events, subscriber}
		
		case event := <- publish: // when new event comes
			// event is sent to all users
			for e := subscribers.Front(); e != nil; e = e.Next() {
				subscriber := e.Value.(chan Event)
				subscriber <- event
			}

			if archive.Len() >= 20 {
				archive.Remove(archive.Front())
			}
			archive.PushBack(event)

		case c := <-unsubscribe: // when user it out
			for e := subscribers.Front(); e != nil; e = e.Next() {
				subscriber := e.Value.(chan Event)

				if subscriber == c {
					subscribers.Remove(e)
					break
				}
			}
		}
	}
}

func main() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	go Chatroom()

	server.On("connection", func(so socketio.Socket) {
		s := Subscribe()
		Join(so.Id())

		for _, event := range s.Archive {
			so.Emit("event", event)
		}	

		newMessages := make(chan string)

		so.On("message", func(msg string) {
			newMessages <- msg
		})

		so.On("disconnection", func() {
			Leave(so.Id())
			s.Cancel()	
		})

		go func() {
			for {
				select {
				case event := <-s.New: // when event comes
					so.Emit("event", event) // send event data to web
				case msg := <- newMessages: // when web sends message
					Say(so.Id(), msg) // publish message
				}
			}
		}()
	})

	http.Handle("/socket.io/", server)

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8000", nil)
}