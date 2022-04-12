package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"

	pb "hworld-client/grpc"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "hworld-server-service.default.svc.cluster.local:8080", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func getNumbers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, struct {
		Num int `json:"num"`
	}{Num: 1})
}

type data struct {
	Num string
}

func talkToServer(data *data) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := pb.NewStreamServiceClient(conn)
	in := &pb.Request{Id: 1}
	stream, err := client.FetchResponse(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	done := make(chan bool)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}
			data.Num = resp.Result
		}
	}()
}

func listen(data *data) {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": data.Num,
		})
	})

	router.Run(":80")
}

func main() {
	flag.Parse()
	data := data{}
	talkToServer(&data)
	listen(&data)
}
