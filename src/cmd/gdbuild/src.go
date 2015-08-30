package main

import (
	"fmt"
	"log"
	"path/filepath"

	"pault.ag/go/debian/changelog"
	"pault.ag/go/debian/control"
)

var tarballDirs = []string{
	"../tarballs",
	"..",
}

func buildSrc(dir string) (control.Control, string) {
	con, err := control.ParseControlFile(filepath.Join(dir, "debian/control"))
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	chg, err := changelog.ParseFileOne(filepath.Join(dir, "debian/changelog"))
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	if !chg.Version.IsNative() {
		origBase := fmt.Sprintf("%s_%s.orig", con.Source.Source, chg.Version.Version)
		origs := []string{}
		for _, tarballDir := range tarballDirs {
			if !filepath.IsAbs(tarballDir) {
				tarballDir = filepath.Join(dir, tarballDir)
			}
			tarballs, err := filepath.Glob(filepath.Join(tarballDir, origBase+".tar.*"))
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
			if len(tarballs) > 0 {
				if len(tarballs) > 1 {
					log.Fatalf("error: found multiple base orig tarballs: %v\n", tarballs)
				}
				orig := tarballs[0]
				origs = append(origs, orig)
				tarballs, err = filepath.Glob(filepath.Join(tarballDir, origBase+"-*.tar.*"))
				if err != nil {
					log.Fatalf("error: %v\n", err)
				}
				origs = append(origs, tarballs...)
				break
			}
		}
		if len(origs) < 1 {
			log.Fatalf("error: unable to find orig tarball(s); searched for %s in %v\n", origBase+".tar.*", tarballDirs)
		}
		log.Fatalf("ORIG TARBALLS, BABY: %v\n", origs)
	}

	log.Fatal("TODO build the source dsc")

	return *con, ""
}
