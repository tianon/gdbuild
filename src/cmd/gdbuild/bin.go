package main

import (
	"fmt"
	"log"
	"path/filepath"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/resolver"
)

func buildBin(dscFile string) (control.DSC, string) {
	dscDir := filepath.Dir(dscFile)
	dsc, err := control.ParseDscFile(dscFile)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	if err := dsc.Validate(); err != nil {
		log.Fatalf("error, validation failed: %v\n", err)
	}

	dscMd5, err := md5sum(dscFile)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	img := fmt.Sprintf("gdbuild/bin:%s_%s", dsc.Source, dscMd5)

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

	buildEssential, err := dependency.Parse("build-essential, dpkg-dev, fakeroot")
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
	// TODO allow this to instead be "FROM scratch\nADD some-chroot-tarball.tar.* /\n"

	dockerfile += `
RUN apt-get update && apt-get install -y --no-install-recommends \
` // --no-install-recommends
	for _, pkg := range allBins {
		dockerfile += fmt.Sprintf("\t\t%s=%s \\\n", pkg.Package, pkg.Version)
	}
	dockerfile += "\t&& rm -rf /var/lib/apt/lists/*\n"

	files := []string{dsc.Filename}
	for _, f := range dsc.Files {
		files = append(files, filepath.Join(dscDir, f.Filename))
	}

	dockerfile += "COPY"
	for _, f := range files {
		dockerfile += " " + filepath.Base(f)
	}
	dockerfile += " /usr/src/.in/\n"

	dockerfile += fmt.Sprintf(`
WORKDIR /usr/src
RUN chown -R nobody:nogroup .
USER nobody:nogroup
RUN dpkg-source -x %q %q
RUN cd %q && dpkg-buildpackage -uc -us
`, ".in/"+filepath.Base(dsc.Filename), dsc.Source, dsc.Source)

	err = dockerBuild(img, dockerfile, files...)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	return *dsc, img
}