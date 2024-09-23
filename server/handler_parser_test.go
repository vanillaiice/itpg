package server

import (
	"bytes"
	"os"
	"testing"
)

func parseHandlersTest(t *testing.T) {
	handlersFile, err := os.ReadFile("handlers.json")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = parseHandlers(bytes.NewReader(handlersFile)); err != nil {
		t.Error(err)
	}
}
