package utils

import (
	"strings"
	"unilab-backend/logging"
)

type DiffResult struct {
	InsertLines uint32
	DeleteLines uint32
	FileChanged uint32
}

func DirDiff(prevDir, nowDir string) DiffResult {
	result := DiffResult{
		InsertLines: 0,
		DeleteLines: 0,
		FileChanged: 0,
	}
	if nowDir == "" || prevDir == "" {
		return result
	}
	response := Subprocess("", 5, "git diff", "",
		"--no-index",
		"--numstat",
		"--no-color",
		"--ignore-space-at-eol",
		"--ignore-cr-at-eol",
		"--ignore-blank-lines",
		prevDir,
		nowDir,
	)
	logging.Info("diff <", prevDir, ">, <", nowDir, "> response: ", response)
	// parse outputs to numbers
	if response.ExitCode != 0 || response.ServerErr != "" {
		logging.Info("error occurred when using git diff...")
		return result
	}
	lines := strings.Split(strings.Trim(response.StdOut, " "), "\n")
	for _, line := range lines {
		temp := strings.Split(line, "\t")
		if len(temp) < 2 {
			continue
		}
		insert, err := StringToUint32(temp[0])
		if err != nil {
			logging.Info(err)
			continue
		}
		delete, err := StringToUint32(temp[1])
		if err != nil {
			logging.Info(err)
			continue
		}
		result.InsertLines += insert
		result.DeleteLines += delete
		result.FileChanged++
	}
	return result
}
