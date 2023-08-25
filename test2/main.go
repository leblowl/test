package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/linxGnu/grocksdb"
	"log"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{}

type Message struct {
	Timestamp string `json:"timestamp"`
	Body      string `json:"body"`
}

// These events are sent to the main thread when messages have been
// added to the message store
type UpdateEvent struct {
	Messages []Message
}

// The message store uses RocksDB as it's underlying data store. Data
// is ordered by key on disk.
type MessageStore struct {
	DB           *grocksdb.DB
	ReadOptions  *grocksdb.ReadOptions
	WriteOptions *grocksdb.WriteOptions
}

// Create a new message store
func NewMessageStore(file string) (*MessageStore, error) {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	db, err := grocksdb.OpenDb(opts, "/app/"+file)

	if err != nil {
		log.Println("Error opening DB", err)
		return nil, err
	}

	ro := grocksdb.NewDefaultReadOptions()
	wo := grocksdb.NewDefaultWriteOptions()

	return &MessageStore{DB: db, ReadOptions: ro, WriteOptions: wo}, nil
}

// Add messages to message store
// TODO: Add combined timestamp/user key
func (ms *MessageStore) AddMessages(messages []Message) error {
	// TODO: batch puts with DB.Write
	for _, message := range messages {
		jsonData, err := json.Marshal(message)
		if err != nil {
			return err
		}
		// FIXME: Probably a better way to get a byte buffer for an integer
		err = ms.DB.Put(ms.WriteOptions, []byte(message.Timestamp), []byte(jsonData))
		if err != nil {
			return err
		}
	}
	return nil
}

// Get num messages from position pos from the data store
func (ms *MessageStore) GetMessages(fromTimestamp string, num int) ([]Message, error) {
	messages := []Message{}

	it := ms.DB.NewIterator(ms.ReadOptions)
	defer it.Close()

	if fromTimestamp == "now" {
		it.SeekToLast()
	} else {
		it.SeekForPrev([]byte(fromTimestamp))
		if it.Valid() {
			it.Prev()
		}
	}

	for i := 0; i < num && it.Valid(); it.Prev() {
		var message Message
		// TODO: Free value slice here?
		err := json.Unmarshal(it.Value().Data()[:], &message)
		if err != nil {
			log.Println("Failed to unmarshal DB value", err)
			return nil, err
		}
		messages = append(messages, message)
		i = i + 1
	}

	if err := it.Err(); err != nil {
		log.Println("err", err)
		return nil, err
	}
	return messages, nil
}

func main() {

	log.Println("Setting up data store...")
	store, err := NewMessageStore("db")
	if err != nil {
		log.Println("Error setting up data store")
		os.Exit(1)
	}

	// This is the event channel for updates that are potentially
	// pushed to the client
	updateEvents := make(chan UpdateEvent)

	// Simulate recieving a huge number of messages in bulk
	go func() {
		// load a hundred thousand messages
		num := 100000

		for i := 0; i < num; i++ {
			past := time.Now().UTC().Add(-1 * time.Hour).Add(-time.Duration(i) * time.Second)
			timestamp := past.Format(time.RFC3339)
			messages := []Message{Message{Timestamp: timestamp, Body: "Hello World " + timestamp}}
			err = store.AddMessages(messages)
			if err != nil {
				log.Println("[New Messages] Error putting data into DB", err)
			}
		}
		log.Println("Loaded historical messages", num)
	}()

	// Simulate recieving new messages
	go func() {
		var now time.Time

		for {
			now = time.Now().UTC()
			timestamp := now.Format(time.RFC3339)
			messages := []Message{Message{Timestamp: timestamp, Body: "Hello World " + timestamp}}
			err = store.AddMessages(messages)
			if err != nil {
				log.Println("[New Messages] Error putting data into DB", err)
			}
			log.Println("Latest timestamp", now.Format(time.RFC3339))
			// TODO: Add mechanism to only push to this channel when someone is listening
			updateEvents <- UpdateEvent{Messages: messages}
			time.Sleep(time.Second * 1)
		}
	}()

	// Simulate recieving individual late messages
	go func() {
		for {
			past := time.Now().UTC().Add(-10 * time.Second)
			timestamp := past.Format(time.RFC3339)
			messages := []Message{Message{Timestamp: timestamp, Body: "Hello World " + timestamp}}
			err = store.AddMessages(messages)
			if err != nil {
				log.Println("[New Messages] Error putting data into DB", err)
			}
			// TODO: Add mechanism to only push to this channel when someone is listening
			updateEvents <- UpdateEvent{Messages: messages}
			time.Sleep(time.Second * 5)
		}
	}()

	// Simulate recieving chunk of late messages
	go func() {
		for {
			messages := []Message{}
			for i := 1; i <= 10; i++ {
				past := time.Now().UTC().Add(time.Duration(-10*i) * time.Second)
				timestamp := past.Format(time.RFC3339)
				messages = append(messages, Message{Timestamp: timestamp, Body: "Hello World " + timestamp})
			}
			err = store.AddMessages(messages)
			if err != nil {
				log.Println("[New Messages] Error putting data into DB", err)
			}
			// TODO: Add mechanism to only push to this channel when someone is listening
			updateEvents <- UpdateEvent{Messages: messages}
			time.Sleep(time.Second * 20)
		}
	}()

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Websocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		// Earliest timestamp for session. All messages recieved
		// later than this timestamp are pushed to the client.
		sessionTimestamp := ""
		readRequests := make(chan string)

		// Handle message read requests from the client
		go func() {
			// Assume message type is always websocket.TextMessage
			// and assume the only message is a read request for next
			// set of messages (for now). It would be easy to adjust this
			// so the client can retrieve a specific range of messages.
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println("Websocket read failed:", err)
					// TODO: Return properly with HTTP response
					os.Exit(1)
				}
				readRequests <- string(msg)
			}
		}()

		for {
			// Handle read request from client or update event from other
			// go routines
			select {
			case _ = <-readRequests:
				log.Println("Recieved read request")
				fromTimestamp := sessionTimestamp
				if sessionTimestamp == "" {
					fromTimestamp = "now"
				}
				messages, err := store.GetMessages(fromTimestamp, 10)
				if err != nil {
					log.Println("Failed getting messages")
				}

				// Update earliest timestamp
				if len(messages) > 0 {
					sessionTimestamp = messages[len(messages)-1].Timestamp
					log.Println("Session timestamp:", sessionTimestamp)
				}

				jsonData, err := json.Marshal(messages)
				if err != nil {
					log.Println("Failed to marshal JSON data", err)
				}

				err = conn.WriteMessage(websocket.TextMessage, []byte(jsonData))
				if err != nil {
					log.Println("Websocket write failed:", err)
					break
				}
			case evt := <-updateEvents:
				pushMessages := []Message{}

				for _, msg := range evt.Messages {
					// If message is later than the session timestamp, then it's within
					// their view, push to client.
					if sessionTimestamp != "" && msg.Timestamp > sessionTimestamp {
						pushMessages = append(pushMessages, msg)
					}
				}

				if len(pushMessages) > 0 {
					jsonData, err := json.Marshal(pushMessages)
					if err != nil {
						log.Println("Failed to marshal JSON data", err)
					}

					err = conn.WriteMessage(websocket.TextMessage, []byte(jsonData))
					if err != nil {
						log.Println("Websocket write failed:", err)
						break
					}
					log.Println("Push event", pushMessages)
				}
			}
		}
	})

	log.Println("Starting web server on port 8080")
	http.ListenAndServe(":8080", nil)
}
