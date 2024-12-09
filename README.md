
# agrep 
### Anagram Goofy Exploration and Pattern-Matcher

`agrep` is a command-line tool designed for searching first anagram of a specified pattern in a given file. It supports optional creation of a hash index to optimize search performance when using as an input file.

## Usage

```bash
./agrep [--create-index] file pattern
```

### Options

- `--create-index`: Creates a hash index file based on the provided input file. This index file can be reused for faster searches when passed as an argument.

### Arguments

- `file`: The input file to search. This can be a plain text file or an existing hash index file with the `.idx` extension.
- `pattern`: The text pattern to search for anagrams. This is the word or phrase you want to find first anagram of in the provided `file`. The pattern should be a single word.

## Example

To search for the first anagram of the phrase `iba` in `dictionary.txt`:

```bash
./agrep ./testdata/dictionary.txt iba
```

To search for the first anagram of the phrase `iba` and create a hash index file for `dictionary.txt`:

```bash
./agrep --create-index ./testdata/dictionary.txt iba
```
Note: Creating index file require to read whole file!

Above command will give us this output:
```
-- Anagram found: [2] abi

...
<Memory stats info>
...
```

Once the index file is created, it can be used for improved performance.

```bash
./agrep --create-index ./testdata/dictionary.idx iba
```

This will return line index of original file where first anagram of `iba` can be found: 
```
-- Anagram found: [2] 

...
<Memory stats info>
...
```

### Memory stats info

As custom hash-table implementation uses it's own memory allocator (to optimize loading of index file into memory). Each run of `agrep` will print memory statistics information:

```
-- Memory Allocator Stats
   Memory Block (Page) Size: 8192
   Number of allocated pages: 5

   Page space utilization:
     * Page 0 - free space: 4 bytes
     * Page 1 - free space: 8 bytes
     * Page 2 - free space: 0 bytes
     * Page 3 - free space: 0 bytes
     * Page 4 - free space: 4232 bytes

```

## Build

Clone the repository and build the project using `go build` command:

```bash
go build -o agrep ./cmd/agrep/main.go
```

This will generate the `agrep` binary in the project directory.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

For more details, refer to the project repository.
