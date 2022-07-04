package collection

import (
	"fmt"
	"log"
	"os"
)

// TODO
func NewCollection(cname string) (col *Collection, err error) {
	log.Fatal("TODO: CODE ME")
	return nil, nil
	// create collection dir
}

// returns error if collection dirpath does not exist
func GetCollection(cname string) (col *Collection, err error) {
	c := &Collection{Name: cname}
	if _, err := os.Stat(c.dirpath()); os.IsNotExist(err) {
		// c.dirpath() does not exist => collection does not exist
		return nil, fmt.Errorf("collection '%s' does not exist", cname)
	}
	return c, nil
}

// TODO
func DeleteCollection(cname string) (col *Collection, err error) {
	log.Fatal("TODO: CODE ME")
	return nil, nil
	// delete collection dir
}
