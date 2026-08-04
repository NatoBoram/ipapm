package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/kubo"
)

func syncAll(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Source, suite string, next apt.InRelease,
) error {
	log.Println("all")
	return nil
}

func syncDiff(
	ctx context.Context, kubo *kubo.Client, client *apt.Client,
	source apt.Source, suite string, next apt.InRelease,
	previous apt.InRelease,
) error {
	diff, err := previous.Diff(next)
	if err != nil {
		return fmt.Errorf("couldn't get the diff between previous and next InRelease files: %s", err)
	}

	log.Printf(
		"Diff: %d added, %d changed, %d removed",
		len(diff.Added), len(diff.Changed), len(diff.Removed),
	)

	return nil
}
