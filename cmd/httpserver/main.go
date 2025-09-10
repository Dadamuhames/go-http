package main

import (
	"flag"
	"fmt"
	"go-http/internal/headers"
	"go-http/internal/request"
	"go-http/internal/response"
	"go-http/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const defaultPort = 42069

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
	port := flag.Int("port", defaultPort, "port")

	flag.Parse()

	server, err := server.Serve(*port, func(res response.ResponseWriter, req *request.Request) {
		target := req.RequestLine.RequestTarget

		switch {
		case target == "/yourproblem":
			hdrs := headers.NewHeaders()

			hdrs.Set("Content-Type", "text/html")

			res.Send(response.HTTP_STATUS_BAD_REQUEST, *hdrs, []byte(badRequest()))

		case target == "/myproblem":
			hdrs := headers.NewHeaders()

			hdrs.Set("Content-Type", "text/html")

			res.Send(response.HTTP_STATUS_INTERNAL_SERVER_ERROR, *hdrs, []byte(serverError()))

		case strings.HasPrefix(target, "/httpbin/stream/"):
			num := strings.Replace(target, "/httpbin/stream/", "", 1)

			resp, err := http.Get(fmt.Sprintf("https://httpbin.org/stream/%s", num))

			if err != nil {
				res.SendBodyWithDefaultHeaders(
					response.HTTP_STATUS_INTERNAL_SERVER_ERROR,
					[]byte("Internal server error"),
				)
				return
			}

			res.SendFromStream(response.HTTP_STATUS_OK, resp.Body)

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
