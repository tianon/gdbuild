package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Lshortfile)

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s some/pkg.dsc|some/dir [...]\n", os.Args[0])
	}

	for _, arg := range os.Args[1:] {
		dsc, img := buildBin(arg)
		fmt.Printf("\n- %q built in %q\n\n", dsc.Source, img)
	}
}
