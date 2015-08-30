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

	// TODO flags, etc

	for _, arg := range os.Args[1:] {
		fi, err := os.Stat(arg)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		if fi.IsDir() {
			con, chg, img := buildSrc(arg)
			fmt.Printf("\n- %q (%q) source DSC built in %q\n\n", con.Source.Source, chg.Version, img)
		} else {
			dsc, img := buildBin(arg)
			fmt.Printf("\n- %q (%q) built in %q\n\n", dsc.Source, dsc.Version, img)
		}
	}
}
