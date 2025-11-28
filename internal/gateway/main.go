package main

import (
	"codev42/internal/gateway/routes"
)

func main() {
	conns, router := routes.SetupRoutes()
	router.Run(":8080")

	for _, conn := range conns {
		defer conn.Close()
	}
}
