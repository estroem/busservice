package main

func GetConfig(key string) string {
	rabbitmq_username := "default_user_8Umc7dr4LB1yLDzbY0T"
	rabbitmq_password := "ZC8Z-05HvbqlYBlhh44fep9j36ESp_Xf"

	if key == "rabbitmq_username" && rabbitmq_username != "%rabbitmq_username%" {
		return rabbitmq_username
	} else if key == "rabbitmq_password" && rabbitmq_password != "%rabbitmq_password%" {
		return rabbitmq_password
	}

	return ""
}
