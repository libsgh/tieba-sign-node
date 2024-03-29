package TiebaSign

import "os"

const MainServer string = "http://localhost:8081"

var DB_TYPE string = getEnv("DB_TYPE", "sqlite")
var PORT string = getEnv("PORT", "8088")
var NODE_NAME string = getEnv("NODE_NAME", "http://localhost:8088")

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
