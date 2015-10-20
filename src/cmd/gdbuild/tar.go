package main

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

func AddStringToTar(tw *tar.Writer, name, file string) error {
	hdr := &tar.Header{
		Name: name,
		Size: int64(len(file)),

		Mode: 0666,

		Uid: syscall.Getuid(),
		Gid: syscall.Getgid(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(file)); err != nil {
		return err
	}
	return nil
}

func AddFileToTar(tw *tar.Writer, name, file string) error {
	return addFileToTarRecurse(tw, name, file, true)
}

func addFileToTarRecurse(tw *tar.Writer, name, file string, recurse bool) error {
	fi, err := os.Lstat(file)
	if err != nil {
		return err
	}

	if fi.IsDir() && recurse {
		err = filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			subPath, err := filepath.Rel(file, path)
			if err != nil {
				return err
			}
			return addFileToTarRecurse(tw, name+"/"+subPath, path, false)
		})
		if err != nil {
			return err
		}
		return nil
	}

	linkTarget := ""
	if fi.Mode()&os.ModeSymlink != 0 {
		linkTarget, err = os.Readlink(file)
		if err != nil {
			return err
		}
	}

	hdr, err := tar.FileInfoHeader(fi, linkTarget)
	if err != nil {
		return err
	}
	hdr.Name = name

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	if fi.Mode().IsRegular() {
		fh, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fh.Close()

		_, err = io.Copy(tw, fh)
		if err != nil {
			return err
		}
	}

	return nil
}
