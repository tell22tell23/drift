package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sammanbajracharya/drift_cli/internal/cli"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			fmt.Printf("Error loading .env file: %v", err)
		}
	}

	app := cli.NewApp()
	if err := app.Run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}
