package main

import (
	"fmt"

	"github.com/devopsext/trends-back/trends"
)

// This is just for testing purposes, maybe we'll use cobra later
func main() {
	fmt.Println(trends.GetTrends())
}
