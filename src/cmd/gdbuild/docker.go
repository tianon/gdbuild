package main

import (
	"archive/tar"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// see https://github.com/docker/distribution/blob/0a3acb2625f553a71392ed9dc6ba1dde2cf3e3a8/reference/regexp.go#L25
const tagRepWith = "-"

var tagRep = regexp.MustCompile(`[^\w.-]`)

func scrubForDockerTag(s string) string {
	return tagRep.ReplaceAllString(s, tagRepWith)
}

func docker(args ...string) *exec.Cmd {
	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr
	return cmd
}

func dockerCpTmp(img string, path string) (string, error) {
	cidBytes, err := docker("create", img, "true").Output()
	if err != nil {
		return "", err
	}
	cid := strings.TrimSpace(string(cidBytes))
	defer docker("rm", "-vf", string(cid)).Run()

	dir, err := ioutil.TempDir("", "gdbuild-")
	if err != nil {
		return "", err
	}

	cmd := docker("cp", string(cid)+":"+path, dir)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}

func dockerBuild(tag string, dockerfile string, files ...string) error {
	dockerfileMd5, err := md5string(dockerfile)
	if err != nil {
		return err
	}
	dockerfileFile := ".gdbuild-dockerfile." + dockerfileMd5

	cmd := docker("build", "--rm", "--force-rm", "-f", dockerfileFile, "-t", tag, "-")
	cmd.Stdout = os.Stdout

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	tw := tar.NewWriter(stdin)
	defer tw.Close()

	if err := AddStringToTar(tw, dockerfileFile, dockerfile); err != nil {
		cmd.Process.Kill()
		return err
	}
	if err := tw.Flush(); err != nil {
		cmd.Process.Kill()
		return err
	}

	for _, file := range files {
		if err := AddFileToTar(tw, filepath.Base(file), file); err != nil {
			cmd.Process.Kill()
			return err
		}
		if err := tw.Flush(); err != nil {
			cmd.Process.Kill()
			return err
		}
	}

	if err := tw.Close(); err != nil {
		cmd.Process.Kill()
		return err
	}
	if err := stdin.Close(); err != nil {
		cmd.Process.Kill()
		return err
	}

	return cmd.Wait()
}
