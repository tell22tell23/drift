package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/sammanbajracharya/drift/internal/cli"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Printf("Error loading .env file: %v", err)
	}

	app := cli.NewApp()
	if err := app.Run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}
