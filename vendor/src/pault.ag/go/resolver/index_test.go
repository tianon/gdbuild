package resolver_test

import (
	"log"
	"strings"
	"testing"

	"pault.ag/go/debian/dependency"
	"pault.ag/go/resolver"
)

// Test Binary Index {{{
var testBinaryIndex = `Package: android-tools-fastboot
Source: android-tools
Version: 4.2.2+git20130529-5.1
Installed-Size: 184
Maintainer: Android tools Maintainer <android-tools-devel@lists.alioth.debian.org>
Architecture: amd64
Depends: libc6 (>= 2.14), libselinux1 (>= 2.0.65), zlib1g (>= 1:1.2.3.4)
Description: Android Fastboot protocol CLI tool
Homepage: http://developer.android.com/guide/developing/tools/adb.html
Description-md5: 56b9309fa4fb2f92a313a815c7d7b5d3
Section: devel
Priority: extra
Filename: pool/main/a/android-tools/android-tools-fastboot_4.2.2+git20130529-5.1_amd64.deb
Size: 56272
MD5sum: cd858b3257b250747822ebeea6c69f4a
SHA1: 9d45825f07b2bc52edc787ba78966db0d4a48e69
SHA256: c094b7e53eb030957cdfab865f68c817d65bf6a1345b10d2982af38d042c3e84

Package: android-tools-fsutils
Source: android-tools
Version: 4.2.2+git20130529-5.1
Installed-Size: 504
Maintainer: Android tools Maintainer <android-tools-devel@lists.alioth.debian.org>
Architecture: amd64
Depends: python:any, libc6 (>= 2.14), libselinux1 (>= 2.0.65), zlib1g (>= 1:1.2.3.4)
Description: Android ext4 utilities with sparse support
Homepage: http://developer.android.com/guide/developing/tools/adb.html
Description-md5: 23135bc652e7b302961741f9bcff8397
Section: devel
Priority: extra
Filename: pool/main/a/android-tools/android-tools-fsutils_4.2.2+git20130529-5.1_amd64.deb
Size: 71900
MD5sum: 996732fc455acdcf4682de4f80a2dc95
SHA1: 5c2320913cc7cc46305390d8b3a7ef51f0a174ef
SHA256: 270ad759d1fef9cedf894c42b5f559d7386aa1ec4de4cc3880eb44fe8c53c833

Package: androidsdk-ddms
Source: androidsdk-tools
Version: 22.2+git20130830~92d25d6-1
Installed-Size: 211
Maintainer: Debian Java Maintainers <pkg-java-maintainers@lists.alioth.debian.org>
Architecture: all
Depends: libandroidsdk-swtmenubar-java (= 22.2+git20130830~92d25d6-1), libandroidsdk-ddmlib-java (= 22.2+git20130830~92d25d6-1), libandroidsdk-ddmuilib-java (= 22.2+git20130830~92d25d6-1), libandroidsdk-sdkstats-java (= 22.2+git20130830~92d25d6-1), eclipse-rcp
Description: Graphical debugging tool for Android
Homepage: http://developer.android.com/tools/help/index.html
Description-md5: a2f559d2abf6ebb1d25bc3929d5aa2b0
Section: java
Priority: extra
Filename: pool/main/a/androidsdk-tools/androidsdk-ddms_22.2+git20130830~92d25d6-1_all.deb
Size: 132048
MD5sum: fde05f3552457e91a415c99ab2a2a514
SHA1: 82b05c97163ccfbbb10a52a5514882412a13ee43
SHA256: fa53e4f50349c5c9b564b8dc1da86c503b0baf56ab95a4ef6e204b6f77bfe70c
`

// }}}

/*
 *
 */

func isok(t *testing.T, err error) {
	if err != nil {
		log.Printf("Error! Error is not nil!\n")
		t.FailNow()
	}
}

func notok(t *testing.T, err error) {
	if err == nil {
		log.Printf("Error! Error is  nil!\n")
		t.FailNow()
	}
}

func assert(t *testing.T, expr bool) {
	if !expr {
		log.Printf("Assertion failed!")
		t.FailNow()
	}
}

/*
 *
 */

func TestResolverBasics(t *testing.T) {
	arch, err := dependency.ParseArch("amd64")
	isok(t, err)

	candidates, err := resolver.ReadFromBinaryIndex(
		strings.NewReader(testBinaryIndex),
	)
	isok(t, err)
	assert(t, len(*candidates) == 3)

	dep, err := dependency.Parse("baz")
	isok(t, err)
	possi := dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == false)

	dep, err = dependency.Parse("android-tools-fsutils")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == true)
}

func TestResolverVersion(t *testing.T) {
	arch, err := dependency.ParseArch("amd64")
	isok(t, err)

	candidates, err := resolver.ReadFromBinaryIndex(
		strings.NewReader(testBinaryIndex),
	)
	isok(t, err)
	assert(t, len(*candidates) == 3)

	dep, err := dependency.Parse("android-tools-fsutils (>= 1.0)")
	isok(t, err)
	possi := dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == true)

	dep, err = dependency.Parse("android-tools-fsutils (>= 1:1.0)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == false)

	dep, err = dependency.Parse("android-tools-fsutils (<= 1:1.0)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == true)

	dep, err = dependency.Parse("android-tools-fsutils (<= 0:0)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == false)

	dep, err = dependency.Parse("android-tools-fsutils (= 4.2.2+git20130529-5.1)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == true)

	dep, err = dependency.Parse("android-tools-fsutils (= 2.2.2+git20130529-5.1)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == false)

	dep, err = dependency.Parse("android-tools-fsutils (<< 4.2.2+git20130529-5.1)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == false)

	dep, err = dependency.Parse("android-tools-fsutils (<< 4.2.2+git20130529-6.1)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == true)

	dep, err = dependency.Parse("android-tools-fsutils (>> 4.2.2+git20130529-5.1)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == false)

	dep, err = dependency.Parse("android-tools-fsutils (>> 4.2.2+git20130529-4.1)")
	isok(t, err)
	possi = dep.GetAllPossibilities()[0]
	assert(t, candidates.Satisfies(*arch, possi) == true)
}

func TestResolverDependsVersion(t *testing.T) {
	candidates, err := resolver.ReadFromBinaryIndex(
		strings.NewReader(testBinaryIndex),
	)
	isok(t, err)
	assert(t, len(*candidates) == 3)

	arch, err := dependency.ParseArch("amd64")
	isok(t, err)

	dep, err := dependency.Parse("android-tools-fsutils (>= 1.0)")
	isok(t, err)
	assert(t, candidates.SatisfiesBuildDepends(*arch, *dep) == true)

	dep, err = dependency.Parse("android-tools-fsutils (>= 1.0), quix")
	isok(t, err)
	assert(t, candidates.SatisfiesBuildDepends(*arch, *dep) == false)
}

// vim: foldmethod=marker
