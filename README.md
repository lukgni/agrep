
# agrep 
### Anagram Goofy Exploration and Pattern-Matcher

`agrep` is a command-line tool designed for searching first anagram of a specified pattern in a given file.

Inspired by the indexing techniques used in database systems, this tool efficiently handles large files (e.g., 4GB) by implementing custom indexing, memory management, and page-based storage to optimize performance.

## Features

- **Efficient Search**: The program can find the first anagram of a specified phrase in a text file.
- **Custom Hash Table**: `agrep` uses a custom hash table optimized for large datasets. The hash table uses a page-based memory storage format, similar to how PostgreSQL organizes index data in memory pages. This format enables efficient data management, as it groups data into fixed-size pages for faster access.
- **File Indexing**: When enabled, the program indexes large text files and stores the custom hash table in a separate index file. This index file can be passed as an argument in place of the original text file for faster searches.
- **Custom Memory Allocator**: The custom memory allocator enables the hash table to be dumped to disk in the page-based format, then loaded back into memory, improving performance for large datasets.


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
Note: Creating index file requires to read the whole file!

Above command will give us this output:
```
-- Anagram found: [2] abi

...
<Memory stats info>
...
```

Once the `.idx` index file is created, it can be used for improved performance.

```bash
./agrep ./testdata/dictionary.idx iba
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
