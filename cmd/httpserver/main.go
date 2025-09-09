package main

import (
	"go-http/internal/headers"
	"go-http/internal/request"
	"go-http/internal/response"
	"go-http/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func badRequest() string {
	return `<html>
    <head>
    <title>400 Bad Request</title>
    </head>
    <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
    </body>
    </html>`
}

func serverError() string {
	return `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
}

func main() {
	server, err := server.Serve(port, func(res response.ResponseWriter, req *request.Request) {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			hdrs := headers.NewHeaders()

			hdrs.Set("Content-Type", "text/html")

			res.Send(response.HTTP_STATUS_BAD_REQUEST, *hdrs, []byte(badRequest()))

		case "/myproblem":
			hdrs := headers.NewHeaders()

			hdrs.Set("Content-Type", "text/html")

			res.Send(response.HTTP_STATUS_INTERNAL_SERVER_ERROR, *hdrs, []byte(serverError()))

		default:
			res.SendEmptyResponse(response.HTTP_STATUS_OK)
		}
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
