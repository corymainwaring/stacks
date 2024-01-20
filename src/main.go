package main

import (
	"log"
	"os"
	"sort"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	f, err := os.OpenFile(".stack", os.O_RDWR, 666)
	if err != nil {
		log.Println(err)
	}
	s := LoadStack(f)
	for _, entries := range s.Sections {
		sort.Sort(ByLeastDependent(entries))
	}
	/*
		err = f.Truncate(0)
		if err != nil {
			log.Println(err)
		}
		_, err = f.Seek(0, 0)
		if err != nil {
			log.Println(err)
		}
		fmt.Fprint(f, s)
	*/
	log.Println(s)
}
