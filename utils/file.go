package utils

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unilab-backend/logging"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func CopyToDstDir(dstDir, srcDir string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		dstFile, err := os.Create(path.Join(dstDir, file.Name()))
		if err != nil {
			return err
		}
		srcFile, err := os.Open(path.Join(srcDir, file.Name()))
		if err != nil {
			return err
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDirSize(path string) (uint32, error) {
	var size uint32
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += uint32(info.Size())
		}
		return err
	})
	return size, err
}

type StatResult struct {
	FileNum   uint32
	DirSize   uint32
	FileLines uint32
}

func DirStat(path string) StatResult {
	result := StatResult{
		FileNum:   0,
		DirSize:   0,
		FileLines: 0,
	}
	if path == "" {
		return result
	}
	size, err := GetDirSize(path)
	if err != nil {
		logging.Info(err)
		return result
	}
	result.DirSize = size
	response := Subprocess("", 5, "ls -lR | grep \"^-\" | wc -l ; find ./ -name \"*\" | xargs cat | wc -l", path)
	logging.Info("stat <", path, "> response: ", response)
	if response.ExitCode != 0 || response.ServerErr != "" {
		logging.Info("error occurred when using dir stat...")
		return result
	}
	lines := strings.Split(strings.Trim(response.StdOut, " "), "\n")
	if len(lines) != 2 {
		logging.Info("dir stat result is not TWO lines but ", len(lines), " lines!")
		return result
	}
	filenum, err := StringToUint32(lines[0])
	if err != nil {
		logging.Info(err)
		return result
	}
	result.FileNum = filenum
	filelines, err := StringToUint32(lines[1])
	if err != nil {
		logging.Info(err)
		return result
	}
	result.FileLines = filelines
	return result
}
