package hashTable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

type PageHeader struct {
	PageId uint16 // Page ID numbert
	Lower  uint16 // Offset to end of free space
	Upper  uint16 // Offset to start of free space
	_      uint16 // Reserved for future use
}

type PageDataId struct {
	Offset uint16
	Length uint16
}

const (
	SMemoryBlockSize       = 8 * 1024
	SPageHeaderSize        = unsafe.Sizeof(PageHeader{})
	SPageDataIdSize        = unsafe.Sizeof(PageDataId{})
	SPageDataPartitionSize = SMemoryBlockSize - SPageHeaderSize
)

type Page struct {
	Header PageHeader
	Data   [SPageDataPartitionSize]byte
}

// MemoryAllocator - No-memory optimization - memory can only grow
type MemoryAllocator struct {
	activePageForMalloc *Page
	pageTable           map[uint16]*Page
}

type MemPtr uint32

var memAlloc *MemoryAllocator = &MemoryAllocator{
	pageTable: map[uint16]*Page{},
}

func newPage(id int) *Page {
	return &Page{
		Header: PageHeader{
			PageId: uint16(id),
			Upper:  uint16(SPageHeaderSize),
			Lower:  uint16(SPageDataPartitionSize),
		},
	}
}

func (p *Page) freeSpace() uint16 {
	spaceLeft := p.Header.Lower - (p.Header.Upper + uint16(SPageDataIdSize))
	return max(0, spaceLeft)
}

func (p *Page) alloc(size uint16) MemPtr {
	chunkId := PageDataId{
		Offset: p.Header.Lower - size,
		Length: size,
	}

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, chunkId)
	if err != nil {
		return 0
	}

	chunkIdOffset := p.Header.Upper
	copy((p.Data)[chunkIdOffset:chunkIdOffset+uint16(SPageDataIdSize)], buf.Bytes()[:SPageDataIdSize])

	p.Header.Upper = chunkIdOffset + uint16(SPageDataIdSize)
	// reserve data in lower memory by moving offset to the end of free memory
	p.Header.Lower = chunkId.Offset

	return MemPtr(uint32(p.Header.PageId)<<16 | uint32(chunkIdOffset&0xFFFF))
}

func (m *MemoryAllocator) Alloc(size uintptr) (MemPtr, error) {
	const maxLenAllowed = SPageDataPartitionSize - SPageDataIdSize
	if size > (maxLenAllowed) {
		return 0, fmt.Errorf("requested length %d is larger than maximum allowed %d", size, maxLenAllowed)
	}

	if m.activePageForMalloc == nil {
		m.activePageForMalloc = newPage(int(0))
		m.pageTable[0] = m.activePageForMalloc
	}

	if size+SPageDataIdSize > uintptr(m.activePageForMalloc.freeSpace()) {
		nextPageId := m.activePageForMalloc.Header.PageId + 1

		m.activePageForMalloc = newPage(int(nextPageId))
		m.pageTable[nextPageId] = m.activePageForMalloc
	}

	return m.activePageForMalloc.alloc(uint16(size)), nil
}

func (m *MemoryAllocator) PrintMemoryUtilization() {
	fmt.Printf("\n--Memory Allocator Stats---\n")
	fmt.Printf("  Memory Block (Page) Size: %d\n", SMemoryBlockSize)
	fmt.Printf("  Number of allocated pages: %d\n", len(m.pageTable))
	fmt.Printf("\n  Page space utilization:\n")

	for i := 0; i < len(m.pageTable); i++ {
		fmt.Printf("    * Page %d - free space: %d bytes\n", i, m.pageTable[uint16(i)].freeSpace())
	}
	fmt.Printf("---------------------------\n")

}

func (p MemPtr) Get(T interface{}) error {
	pageId := uint16((p >> 16) & 0xFFFF)
	chunkIdOffset := uint16(p & 0xFFFF)

	pagePtr, ok := memAlloc.pageTable[pageId]
	if !ok {
		return fmt.Errorf("invalid memory address: %x", p)
	}

	chunkId := PageDataId{}
	err := binary.Read(
		bytes.NewReader(pagePtr.Data[chunkIdOffset:chunkIdOffset+uint16(SPageDataIdSize)]),
		binary.LittleEndian,
		&chunkId)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}

	buff := make([]byte, chunkId.Length)

	err = binary.Read(
		bytes.NewReader(pagePtr.Data[chunkId.Offset:chunkId.Offset+chunkId.Length]),
		binary.LittleEndian,
		&buff)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}

	binary.Read(bytes.NewReader(buff), binary.LittleEndian, T)

	return nil
}

func (p MemPtr) Set(T interface{}) error {
	pageId := uint16((p >> 16) & 0xFFFF)
	chunkIdOffset := uint16(p & 0xFFFF)

	pagePtr, ok := memAlloc.pageTable[pageId]
	if !ok {
		return fmt.Errorf("invalid memory address: %x", p)
	}

	// reading from upper memory
	chunkId := PageDataId{}
	err := binary.Read(
		bytes.NewReader(pagePtr.Data[chunkIdOffset:chunkIdOffset+uint16(SPageDataIdSize)]),
		binary.LittleEndian,
		&chunkId)
	if err != nil {
		return fmt.Errorf("eror in accesing stored data information: %v", err)
	}

	// TODO: return error if passed buffer is larger in size than allocated memory that we get from [ChunkID]
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.LittleEndian, T)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// writing to lower memory
	copy((pagePtr.Data)[chunkId.Offset:chunkId.Offset+chunkId.Length], buf.Bytes()[:chunkId.Length])
	return nil
}

func memAllocDumpMemoryIntoFile(filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 0; i < len(memAlloc.pageTable); i++ {
		err = binary.Write(f, binary.LittleEndian, memAlloc.pageTable[uint16(i)])
		if err != nil {
			return err
		}
	}

	return nil
}

func memAllocResetNInitFromFile(filepath string) error {
	memAlloc = &MemoryAllocator{
		pageTable: map[uint16]*Page{},
	}

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	// calculate number of memory pages dumped into the file
	fInfo, _ := f.Stat()
	memPagesNum := int(fInfo.Size() / SMemoryBlockSize)

	if memPagesNum <= 0 {
		return fmt.Errorf("error in identifying dumped memory pages in file")
	}

	for i := 0; i < memPagesNum; i++ {
		pageBuff := Page{}
		err = binary.Read(f, binary.LittleEndian, &pageBuff)
		if err != nil {
			return err
		}

		memAlloc.pageTable[uint16(i)] = &pageBuff
		memAlloc.activePageForMalloc = &pageBuff
	}

	return nil
}
