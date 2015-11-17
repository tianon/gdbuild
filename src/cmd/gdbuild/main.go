package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

			dir, err := dockerCpTmp(img, "/usr/src/.out")
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
			defer os.RemoveAll(dir)
			outVer := chg.Version
			outVer.Epoch = 0
			arg = filepath.Join(dir, ".out", fmt.Sprintf("%s_%s.dsc", con.Source.Source, outVer))
		}

		dsc, img := buildBin(arg)
		fmt.Printf("\n- %q (%q) built in %q\n\n", dsc.Source, dsc.Version, img)

		if testsuite, ok := dsc.Values["Testsuite"]; ok && testsuite == "autopkgtest" {
			testImg := autopkgtest(img, dsc)
			fmt.Printf("\n- tests run in %s\n\n", testImg)
		}
	}
}
