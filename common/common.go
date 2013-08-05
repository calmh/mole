package common

import (
	"log"
)

func bugOn(e error) {
	if e != nil {
		log.Fatal("BUG:", e)
	}
}
