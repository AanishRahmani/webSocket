package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

func main() {

	origin := "http://localhost/"
	url := "ws://localhost:3000/ws"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	var wg sync.WaitGroup // sync goroutines

	sentMessages := make(chan string, 10) // Channel to track messages sent by the client

	//send messages to the server
	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(os.Stdin)
		for {

			message, _ := reader.ReadString('\n')
			message = strings.TrimSpace(message)
			_, err := ws.Write([]byte(message))
			if err != nil {
				log.Println("Error sending message:", err)
				return
			}

			sentMessages <- message
		}
	}()

	// Goroutine to receive messages from the server
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var msg = make([]byte, 512)
			n, err := ws.Read(msg)
			if err != nil {
				log.Println("Error reading message:", err)
				return
			}
			receivedMessage := strings.TrimSpace(string(msg[:n]))

			// Check if the received message was sent by this client
			select {
			case sent := <-sentMessages:
				if receivedMessage == sent {
					continue // Skip printing the sent message
				}
			default:
			}

			fmt.Printf("%s\n", receivedMessage)
		}
	}()

	// Wait for both goroutines to finish (they won't, as they run indefinitely)
	wg.Wait()
}
