package main

func GetConfig(key string) string {
	rabbitmq_username := "%rabbitmq_username%"
	rabbitmq_password := "%rabbitmq_password%"

	if key == "rabbitmq_username" && rabbitmq_username != "%rabbitmq_username%" {
		return rabbitmq_username
	} else if key == "rabbitmq_password" && rabbitmq_password != "%rabbitmq_password%" {
		return rabbitmq_password
	}

	return ""
}
