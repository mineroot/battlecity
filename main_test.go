package main

import (
	"log"
	"testing"
)

func TestTest(t *testing.T) {
	a := [2][2]bool{{true, true}, {true, true}}
	b := [2][2]bool{{true, true}, {true, true}}
	log.Println(a == b)
}
