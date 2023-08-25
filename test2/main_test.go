package main

import (
        "golang.org/x/exp/slices"
	"testing"
	"time"
	"reflect"
	"strings"
)


func TestAddGetMessages(t *testing.T) {

	store, err := NewMessageStore("db_test1")
	if err != nil {
		t.Error("Error setting up data store")
	}
	defer store.DB.Close()


	messages := []Message{}
	num := 4

	timestampDesc := func (a Message, b Message) int {
		return strings.Compare(b.Timestamp, a.Timestamp)
	}

	for i := 0; i < num; i++ {
		now := time.Now().UTC().Add(time.Duration(i) * time.Second)
		timestamp := now.Format(time.RFC3339)
		messages = append(messages, Message{Timestamp: timestamp, Body: "Hello World " + timestamp})
	}
	if slices.IsSortedFunc(messages, timestampDesc) {
		t.Errorf("Messages should not be sorted in descending order yet")
	}

	err = store.AddMessages(messages)
	if err != nil {
		t.Error("Error putting data into DB")
	}

	got, err := store.GetMessages("now", num)
	if err != nil {
		t.Error("Error retrieving data from DB")
	}

	if !slices.IsSortedFunc(got, timestampDesc) {
		t.Errorf("Messages not sorted in descending order")
	}
	slices.SortFunc(messages, timestampDesc)
	if !reflect.DeepEqual(got, messages) {
		t.Errorf("Got %q, wanted %q", got, messages)
	}
}

func TestAddGetMessagesMillion(t *testing.T) {

	store, err := NewMessageStore("db_test2")
	if err != nil {
		t.Error("Error setting up data store")
	}
	defer store.DB.Close()

	messages := []Message{}
	num := 1000000

	for i := 0; i < num; i++ {
		now := time.Now().UTC().Add(-time.Duration(i) * time.Second)
		timestamp := now.Format(time.RFC3339)
		messages = append(messages, Message{Timestamp: timestamp, Body: "Hello World " + timestamp})
	}

	err = store.AddMessages(messages)
	if err != nil {
		t.Error("Error putting data into DB")
	}

	got, err := store.GetMessages("now", num)
	if err != nil {
		t.Error("Error retrieving data from DB")
	}

	if !slices.IsSortedFunc(got, func(a Message, b Message) int { return strings.Compare(b.Timestamp, a.Timestamp) }) {
		t.Errorf("Messages not sorted in descending order")
	}
	if !reflect.DeepEqual(got, messages) {
		t.Errorf("got %q, wanted %q", got, messages)
	}
}

func TestAddGetMessagesLimit(t *testing.T) {

	store, err := NewMessageStore("db_test3")
	if err != nil {
		t.Error("Error setting up data store")
	}
	defer store.DB.Close()

	messages := []Message{}
	num := 5

	for i := 0; i < num; i++ {
		now := time.Now().UTC().Add(-time.Duration(i) * time.Second)
		timestamp := now.Format(time.RFC3339)
		messages = append(messages, Message{Timestamp: timestamp, Body: "Hello World " + timestamp})
	}

	err = store.AddMessages(messages)
	if err != nil {
		t.Error("Error putting data into DB")
	}

	got, err := store.GetMessages("now", num - 2)
	if err != nil {
		t.Error("Error retrieving data from DB")
	}

	if !reflect.DeepEqual(got, messages[:3]) {
		t.Errorf("got %q, wanted %q", got, messages)
	}
}
