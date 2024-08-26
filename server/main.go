package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func newServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	log.Println("New incoming connection from client:", ws.RemoteAddr())

	s.conns[ws] = true
	defer func() {
		delete(s.conns, ws)
		ws.Close()
	}()

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Read error:", err)
			continue
		}
		msg := buf[:n]
		log.Println("Received:", string(msg))

		// Echo the message back to the client
		if _, err := ws.Write(msg); err != nil {
			fmt.Println("Write error:", err)
		}

		// Broadcast the message to all connected clients except the sender
		s.broadcast(msg, ws)
		log.Println("message sent:", string(msg))
	}
}

func (s *Server) broadcast(msg []byte, sender *websocket.Conn) {
	for ws := range s.conns {
		// Skip the sender
		if ws == sender {
			continue
		}
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(msg); err != nil {
				fmt.Println("Broadcast write error:", err)
			}
		}(ws)
	}
}

func main() {
	server := newServer()
	http.Handle("/ws", websocket.Handler(server.handleWS))

	fmt.Println("Server started on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		fmt.Println("ListenAndServe error:", err)
	}
}
