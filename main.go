package main

import (
	"context"
	"fmt"

	application "github.com/thiagobgarc/orders-api/app"
)

func main() {
	app := application.New()

	err := app.Start(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}
