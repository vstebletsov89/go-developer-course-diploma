package main

import (
	"go-developer-course-diploma/internal/configs"
	"go-developer-course-diploma/internal/gophermart"
	"log"
)

func main() {
	cfg, err := configs.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := gophermart.RunApp(cfg); err != nil {
		log.Fatal(err)
	}
}
