package server

import (
	"fmt"
	"go-http/internal/request"
	"go-http/internal/response"
	"log"
	"net"
)

type Handler func(w response.ResponseWriter, req *request.Request)

type Server struct {
	closed  bool
	handler Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := &Server{
		closed:  false,
		handler: handler,
	}

	go server.listen(ln)

	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}

func (s *Server) listen(ln net.Listener) {
	log.Println("Start listening")

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Fatal("TCP connection error:", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	log.Println("handling connection")

	req, err := request.RequestFromReader(conn)

	responseWriter := response.NewResponseWriter(conn)

	if err != nil {
		responseWriter.SendEmptyResponse(response.HTTP_STATUS_BAD_REQUEST)
		return
	}

	s.handler(*responseWriter, req)

	conn.Close()
}
