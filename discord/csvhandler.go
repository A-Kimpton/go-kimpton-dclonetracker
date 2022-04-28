package discord

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sync"
)

var (
	filePath = flag.String("db", "./defaultdb.csv", "The database file containing channelids")
	lock sync.RWMutex // lock for file
)

type DiscordMessage struct {
	ChannelID string // Universally unique
	MessageID string // Unique in channel - if 0, use for alerting
}

func GetAllMessages() (dm []DiscordMessage) {

	// Get lock
	lock.RLock()
	defer lock.RUnlock()

	// Placeholder
	dm = make([]DiscordMessage, 0)

	f, err := os.Open(*filePath)
    if err != nil {
        fmt.Printf("Could not read file %s\n", *filePath)
    }
    defer f.Close()

    csvReader := csv.NewReader(f)
    records, err := csvReader.ReadAll()
    if err != nil {
        fmt.Printf("Unable to parse file %s\n", *filePath)
    }

	for _, record := range(records) {
		newMessage := DiscordMessage{
			ChannelID: record[0],
			MessageID: record[1],
		}
		dm = append(dm, newMessage)
	}

    return dm
}

func AddMessage(newMessage DiscordMessage) () {
	// Get a copy of the messages
	messages := GetAllMessages()
	// Check if a message already exists
	for _, m := range(messages) {
		if m.ChannelID == newMessage.ChannelID {
			// Delete if msgID is both 0 or both not 0
			if m.MessageID == newMessage.MessageID || (m.MessageID != "0" && newMessage.MessageID != "0") {
				RemoveMessage(m) // delete it
			}			
		}
	}
	// Reobtain the list
	messages = GetAllMessages()
	messages = append(messages, newMessage)

	// Get lock
	lock.Lock()
	defer lock.Unlock()

	// Make file (clears it if exists)
	f, err := os.Create(*filePath)
    if err != nil {
        fmt.Printf("failed to open file %s", err)
    }
	defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    for _, record := range messages {
		line := []string{record.ChannelID, record.MessageID}
		if err := w.Write(line); err != nil {
            fmt.Printf("error writing record to file %s", err)
        }
    }
}

func RemoveMessage(m1 DiscordMessage) () {
	// Get a copy of the messages and remove one
	messages := GetAllMessages()
	for i, m2 := range messages {
		if m1.ChannelID == m2.ChannelID && m1.MessageID == m2.MessageID {
			messages = append(messages[:i], messages[i+1:]...)
			break
		}
	}

	// Get lock
	lock.Lock()
	defer lock.Unlock()

	// Make file (clears it if exists)
	f, err := os.Create(*filePath)
    if err != nil {
        fmt.Printf("failed to open file %s", err)
    }
	defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    for _, record := range messages {
		line := []string{record.ChannelID, record.MessageID}
		if err := w.Write(line); err != nil {
            fmt.Printf("error writing record to file %s", err)
        }
    }
}

func RemoveAllMessage(m1 DiscordMessage) () {
	// Get a copy of the messages and remove one
	messages := GetAllMessages()
	messagesToKeep := make([]DiscordMessage,0)
	for _, m2 := range messages {
		if m1.ChannelID != m2.ChannelID {
			messagesToKeep = append(messagesToKeep, m2)
		}
	}
	messages = messagesToKeep

	// Get lock
	lock.Lock()
	defer lock.Unlock()

	// Make file (clears it if exists)
	f, err := os.Create(*filePath)
    if err != nil {
        fmt.Printf("failed to open file %s", err)
    }
	defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    for _, record := range messages {
		line := []string{record.ChannelID, record.MessageID}
		if err := w.Write(line); err != nil {
            fmt.Printf("error writing record to file %s", err)
        }
    }
}


