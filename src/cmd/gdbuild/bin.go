package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/resolver"
)

func binSatPossi(depArch *dependency.Arch, bin control.BinaryIndex, possi dependency.Possibility) bool {
	return !possi.Substvar &&
		(possi.Architectures == nil || possi.Architectures.Matches(depArch)) &&
		possi.Name == bin.Package &&
		(possi.Arch == nil || possi.Arch.Is(&bin.Architecture)) &&
		(possi.Version == nil || possi.Version.SatisfiedBy(bin.Version))
}

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
	bins := map[string]control.BinaryIndex{}
	binsSlice := []string{} // ugh Go
RelLoop:
	for _, rel := range append(buildEssential.Relations, dsc.BuildDepends.Relations...) {
		canRel := false
		for _, possi := range rel.Possibilities {
			if possi.Substvar {
				continue
			}
			if bin, ok := bins[possi.Name]; ok && binSatPossi(depArch, bin, possi) {
				continue RelLoop
			}
		}
	PossiLoop:
		for _, possi := range rel.Possibilities {
			if possi.Substvar {
				continue
			}
			entries, ok := map[string][]control.BinaryIndex(*index)[possi.Name]
			if !ok {
				continue
			}
			for _, bin := range entries {
				if binSatPossi(depArch, bin, possi) {
					if existBin, ok := bins[bin.Package]; ok {
						log.Printf("uh oh, already chose %s=%s but want %s=%s for %q\n", existBin.Package, existBin.Version, bin.Package, bin.Version, possi)
						continue PossiLoop
					}
					bins[bin.Package] = bin
					binsSlice = append(binsSlice, bin.Package)
					canRel = true
					break PossiLoop
				}
			}
		}
		if !canRel {
			log.Printf("warning: unable to satisfy %q\n", rel)
			allCan = false
		}
	}
	sort.Strings(binsSlice)

	if !allCan {
		//log.Fatalf("Unsatisfied possi; exiting.\n")
		log.Println()
		log.Println("WARNING: Unsatisfied possi!")
		log.Println()
	}

	dockerfile := fmt.Sprintf("FROM debian:%s\n", suite)
	// TODO allow this to instead be "FROM scratch\nADD some-chroot-tarball.tar.* /\n"

	dockerfile += `
RUN apt-get update && apt-get install -y --no-install-recommends \
` // --no-install-recommends
	for _, pkg := range binsSlice {
		bin := bins[pkg]
		dockerfile += fmt.Sprintf("\t\t%s=%s \\\n", bin.Package, bin.Version)
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
RUN dpkg-source -x %q pkg
RUN (cd pkg && dpkg-buildpackage -uc -us -d) && mkdir .out && ln %q_* .out/
`, ".in/"+filepath.Base(dsc.Filename), dsc.Source)

	err = dockerBuild(img, dockerfile, files...)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	return *dsc, img
}
