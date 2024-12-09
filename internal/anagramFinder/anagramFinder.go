package anagramFinder

import (
	"agrep/internal/hashTable"
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const SIndexFileSuffix = ".idx"

type FinderResult struct {
	LineIndex uint64
	Value     string
}

func findAnagramFromTextFile(pattern string, createIndexFile bool, filePath string) (*FinderResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	patternCharsSorted := []byte(pattern)
	sort.Slice(patternCharsSorted, func(i, j int) bool { return patternCharsSorted[i] < patternCharsSorted[j] })

	var result *FinderResult
	var lineIndex uint64 = 1

	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		lineCharsSorted := make([]byte, len(line))
		copy(lineCharsSorted, line)

		sort.Slice(lineCharsSorted, func(i, j int) bool { return lineCharsSorted[i] < lineCharsSorted[j] })

		hashTable.Set(string(lineCharsSorted), uint32(lineIndex))

		if bytes.Equal(lineCharsSorted, patternCharsSorted) && result == nil {
			result = &FinderResult{
				LineIndex: lineIndex,
				Value:     string(line),
			}

			if createIndexFile == false {
				break
			}
		}

		lineIndex += 1
	}

	if createIndexFile {
		hashTable.DumpMemoryIntoFile(getIndexFileName(filePath))
	}

	hashTable.PrintMemoryUtilization()
	return result, nil
}

func findAnagramFromIndexFile(pattern string, filePath string) (*FinderResult, error) {
	hashTable.ResetAndInitFromFile(filePath)

	patternCharsSorted := []byte(pattern)
	sort.Slice(patternCharsSorted, func(i, j int) bool { return patternCharsSorted[i] < patternCharsSorted[j] })

	lineIndex, _ := hashTable.GetFirstVal(string(patternCharsSorted))

	return &FinderResult{
		LineIndex: uint64(lineIndex),
		Value:     "",
	}, nil
}

func getIndexFileName(sourceFilePath string) string {
	return strings.TrimSuffix(sourceFilePath, filepath.Ext(sourceFilePath)) + SIndexFileSuffix
}

func FindAnagram(pattern string, createIndexFile bool, filePath string) (*FinderResult, error) {
	if filepath.Ext(filePath) == SIndexFileSuffix {
		return findAnagramFromIndexFile(pattern, getIndexFileName(filePath))
	}

	return findAnagramFromTextFile(pattern, createIndexFile, filePath)
}
