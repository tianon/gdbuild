package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"

	"pault.ag/go/debian/changelog"
	"pault.ag/go/debian/control"
)

var tarballDirs = []string{
	"../tarballs",
	"..",
}

func buildSrc(dir string) (control.Control, changelog.ChangelogEntry, string) {
	con, err := control.ParseControlFile(filepath.Join(dir, "debian/control"))
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	chg, err := changelog.ParseFileOne(filepath.Join(dir, "debian/changelog"))
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	img := fmt.Sprintf("gdbuild/src:%s_%s", con.Source.Source, scrubForDockerTag(chg.Version.String()))

	dockerfile := "FROM debian:unstable\n"
	// TODO allow this to instead be "FROM scratch\nADD some-chroot-tarball.tar.* /\n"

	dockerfile += `
RUN apt-get update && apt-get install -y --no-install-recommends \
		dpkg-dev \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src

`
	files := []string{}

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

		files = append(files, origs...)
		files = append(files, filepath.Join(dir, "debian"))

		dockerfile += "COPY"
		for _, f := range origs {
			dockerfile += " " + filepath.Base(f)
		}
		dockerfile += " /usr/src/.out/\n"
		dockerfile += fmt.Sprintf("RUN ln -s .out/%q.tar.* .out/%q-*.tar.* ./\n", origBase, origBase)

		dockerfile += "\n# origtargz --unpack\n"
		re := regexp.MustCompile(fmt.Sprintf(`^%s(?:-(.*))?\.tar\..*$`, regexp.QuoteMeta(origBase)))
		dockerfile += "RUN set -ex"
		for _, f := range origs {
			orig := filepath.Base(f)
			targetDir := "pkg"
			matches := re.FindStringSubmatch(orig)
			if matches != nil && matches[1] != "" {
				targetDir += "/" + matches[1]
			}
			dockerfile += fmt.Sprintf(" \\\n\t&& mkdir %q && tar -xC %q -f %q --strip-components=1", targetDir, targetDir, orig)
		}
		dockerfile += "\n"

		dockerfile += "\nCOPY debian /usr/src/pkg/debian\n"
	} else {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		files = append(files, absDir)
		dockerfile += fmt.Sprintf("COPY %s /usr/src/pkg\n", filepath.Base(absDir))
	}

	outVer := chg.Version
	outVer.Epoch = 0
	pkgVer := con.Source.Source + "_" + outVer.String()
	links := fmt.Sprintf("%q %q", pkgVer+".dsc", pkgVer+"_source.changes")
	if chg.Version.IsNative() {
		links += fmt.Sprintf(" %q.*", pkgVer+".tar")
	} else {
		links += fmt.Sprintf(" %q.*", pkgVer+".debian.tar")
	}
	dockerfile += fmt.Sprintf(`
# work around overlayfs issues (data inconsistency issues; see https://github.com/docker/docker/issues/10180)
VOLUME /usr/src/pkg
# rm: cannot remove 'pkg/.pc/xyz.patch': Directory not empty

RUN (cd pkg && dpkg-buildpackage -uc -us -S -nc) && mkdir -p .out && ln %s .out/
`, links)

	err = dockerBuild(img, dockerfile, files...)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	return *con, *chg, img
}
