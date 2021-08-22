package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	a := App{}
	a.Init(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"))

	a.Run(":8000")
}
