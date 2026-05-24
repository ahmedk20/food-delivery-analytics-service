package main

import (
	"log"

	"github.com/quickbite/analytics-service/lib/boot"
)

func main() {
	if err := boot.Run(); err != nil {
		log.Fatal(err)
	}
}
