package main

import (
	"fmt"
	"os"

	"github.com/ymgyt/costman/internal/di"
	"github.com/ymgyt/costman/internal/scenario"
)

const version = "0.0.2"

func main() {
	s := di.MustServices()
	if err := scenario.AWSDailyReport(s.AWS, s.Webhook); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
