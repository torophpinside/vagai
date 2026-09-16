package main

import (
	"fmt"
	"os"

	"github.com/anomalyco/vagai-api/internal/middleware"
)

func main() {
	token, err := middleware.GenerateToken(1, 1, "torophpinside@vagai.ia", "owner")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(token)
}