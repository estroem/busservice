package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

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
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
	data.Num = r.GetMessage()
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
