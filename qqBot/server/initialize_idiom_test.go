package server

import (
	"log"
	"testing"
)

func TestInitializeIdiom(t *testing.T) {
	t.Run("test Initializing the corpus", func(t *testing.T) {
		NewIdiomMap()
		for i := 0; i < 5; i++ {
			nextIdiom := FindNextIdiom("锦上添花")
			if nextIdiom != "" {
				log.Printf("Randomly return the eligible words as '%s'", nextIdiom)
				continue
			}
			log.Println("There is no four-character idiom ending in the word '花'")
		}
	})
}
