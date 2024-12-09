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

	filePath := argParser.PositionalArgument("file", "text file or hash index file (.idx)")
	pattern := argParser.PositionalArgument("pattern", "text pattern that will be used to search for anagrams")
	createIndexFile := argParser.BooleanFlag("create-index", "create hash index file based on passed input")

	argParser.Parse()

	foundAnagram, err := anagramFinder.FindAnagram(*pattern, *createIndexFile, *filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	if foundAnagram != nil {
		fmt.Printf("-- Anagram found: [%d] %s\n", foundAnagram.LineIndex, foundAnagram.Value)
	} else {
		fmt.Println("-- Anagram not found...")
	}

	hashTable.PrintMemoryUtilization()
}
