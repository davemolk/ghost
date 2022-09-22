package main

import (
	"log"
	"os"
)

// writeJSON takes in a byte slice and file name and writes
// the contents to a .txt file.
func (g *ghost) writeJSON(name string, data []byte) {
	f, err := os.Create(name)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Println(err)
		return
	}
	err = f.Sync()
	if err != nil {
		log.Println(err)
	}
}