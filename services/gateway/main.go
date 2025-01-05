package main

import (
	"codev42/services/gateway/routes"
)

func main() {
	conn, router := routes.SetupRoutes()
	router.Run(":8080")

	defer conn.Close()
}
