package main

import (
	"%MODULE_NAME%/internal/server"
)

func main() {
	rabbitmq_username := GetConfig("rabbitmq_username")
	rabbitmq_password := GetConfig("rabbitmq_password")

	server.CreateChannel(rabbitmq_username, rabbitmq_password)
	defer server.CloseConnection()
}
