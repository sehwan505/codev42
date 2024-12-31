package main

import (
	"github.com/sehwan505/codev42/routes"
)

func main() {
	r, err := routes.SetupRouter()
	if err != nil {
		return
	}
	r.Run(":8080")
}
