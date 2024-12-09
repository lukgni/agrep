package main

import (
	"agrep/internal/anagramFinder"
	"agrep/internal/argParser"
	"agrep/internal/hashTable"
	"fmt"
	"os"
)

func main() {
	argParser := argParser.ArgumentParser{}

	filePath := argParser.PositionalArgument("file", "file that consists of alphabetically sorted strings")
	pattern := argParser.PositionalArgument("pattern", "text pattern that will be used to search for anagrams.")
	createIndexFile := argParser.BooleanFlag("create-index", "enable b-tree indexing over searching")

	argParser.Parse()

	foundAnagram, err := anagramFinder.FindAnagram(*pattern, *createIndexFile, *filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	if foundAnagram != nil {
		fmt.Printf("Anagram found: [%d] %s\n", foundAnagram.LineIndex, foundAnagram.Value)
	} else {
		fmt.Println("Anagram not found...")
	}

	hashTable.PrintMemoryUtilization()
}
