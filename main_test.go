package main

import (
	"github.com/google/uuid"
	"log"
	"testing"
)

func TestTest(t *testing.T) {
	id := uuid.New()
	var id2 uuid.UUID
	log.Println(id.ID() == 0, id2.ID() == 0)
}
