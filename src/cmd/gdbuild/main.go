package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/resolver"
)

func main() {
	log.SetFlags(log.Lshortfile)

	if len(os.Args) != 2 {
		log.Fatalf("usage: %s something.dsc\n", os.Args[0])
	}

	dsc, err := control.ParseDscFile(os.Args[1])
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// TODO parse this information from an image?  optional commandline parameters?
	suite := "unstable"
	arch := "amd64"
	index, err := resolver.GetBinaryIndex(
		"http://httpredir.debian.org/debian",
		suite,
		"main",
		arch,
	)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	depArch, err := dependency.ParseArch(arch)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	buildEssential, err := dependency.Parse("build-essential")
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	allCan := true
	allPossi := append(
		buildEssential.GetPossibilities(*depArch),
		dsc.BuildDepends.GetPossibilities(*depArch)...,
	)
	allBins := []control.BinaryIndex{}
	for _, possi := range allPossi {
		can, why, bins := index.ExplainSatisfies(*depArch, possi)
		if !can {
			log.Printf("%s: %s\n", possi.Name, why)
			allCan = false
		} else {
			// TODO more smarts for which dep out of bins to use
			allBins = append(allBins, bins[0])
		}
	}

	if !allCan {
		log.Fatalf("Unsatisfied possi; exiting.\n")
	}

	dockerfile := fmt.Sprintf("FROM debian:%s\n", suite)

	dockerfile += `
RUN apt-get update && apt-get install -y \
` // --no-install-recommends
	for _, pkg := range allBins {
		dockerfile += fmt.Sprintf("\t\t%s=%s \\\n", pkg.Package, pkg.Version)
	}
	dockerfile += "\t&& rm -rf /var/lib/apt/lists/*\n"

	dockerfile += "\nWORKDIR /usr/src/pkg\n"

	origVersion := dsc.Version
	origVersion.Revision = ""
	origPrefix := fmt.Sprintf("%s_%s.orig", dsc.Source, origVersion)
	dockerfile += fmt.Sprintf(`
COPY %s*.tar.* /usr/src/
RUN origPrefix=%q \
	&& set -ex \
	&& tar -xf "../$origPrefix".tar.* --strip-components=1 \
	&& for orig in "../$origPrefix-"*.tar.*; do \
		targetDir="$(basename "$orig")"; \
		targetDir="${targetDir#$origPrefix-}" \
		targetDir="${targetDir%%.tar.*}"; \
		mkdir -p "$targetDir"; \
		tar -xf "$orig" --strip-components=1 -C "$targetDir"; \
	done
`, origPrefix, origPrefix)
	dockerfile += fmt.Sprintf("ADD %s_%s.debian.tar.* /usr/src/pkg/\n", dsc.Source, dsc.Version)

	dockerfile += `
RUN chown -R nobody:nogroup ..
USER nobody:nogroup
RUN dpkg-buildpackage -uc -us
`

	files, err := filepath.Glob(fmt.Sprintf("%s_*.tar.*", dsc.Source))
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	err = dockerBuild(fmt.Sprintf("debian/pkg-%s", dsc.Source), dockerfile, files...)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
