package hashTable

import (
	"encoding/binary"
	"unsafe"
)

const SBucketNum = 1024

type ValueListNodePtr MemPtr

type ValueListNode struct {
	Value uint32
	Next  ValueListNodePtr
}

type BucketListNodePtr MemPtr

type BucketListNode struct {
	Hash             uint32
	ValueListNodePtr ValueListNodePtr
	Next             BucketListNodePtr
}

type HashTable struct {
	Buckets [SBucketNum]BucketListNodePtr
	Seed    uint32
}

var hPtr MemPtr = 0x0

// initialize main hash table object
var h *HashTable = func() *HashTable {
	hPtr, _ = memAlloc.Alloc(unsafe.Sizeof(HashTable{}))

	obj := HashTable{Seed: 0x1234}
	hPtr.Set(obj)

	return &obj
}()

func murmur3(key string, seed uint32) uint32 {
	data := []byte(key)
	length := len(data)
	nblocks := length / 4
	var h1 uint32 = seed
	const c1 uint32 = 0xcc9e2d51
	const c2 uint32 = 0x1b873593

	for i := 0; i < nblocks; i++ {
		k1 := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])

		k1 *= c1
		k1 = (k1 << 15) | (k1 >> (32 - 15)) // rotl32
		k1 *= c2

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> (32 - 13))
		h1 = h1*5 + 0xe6546b64
	}

	var k1 uint32
	tail := data[nblocks*4:]
	switch len(tail) {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << 15) | (k1 >> (32 - 15))
		k1 *= c2
		h1 ^= k1
	}
	h1 ^= uint32(length)
	h1 ^= (h1 >> 16)
	h1 *= 0x85ebca6b
	h1 ^= (h1 >> 13)
	h1 *= 0xc2b2ae35
	h1 ^= (h1 >> 16)

	return h1
}

func NewBucketListNode(in *BucketListNode) BucketListNodePtr {
	ptr, _ := memAlloc.Alloc(unsafe.Sizeof(BucketListNode{}))
	ptr.Set(*in)

	return BucketListNodePtr(ptr)
}

func NewValueListNode(value uint32) ValueListNodePtr {
	ptr, _ := memAlloc.Alloc(unsafe.Sizeof(ValueListNode{}))
	ptr.Set(ValueListNode{
		Value: value,
	})

	return ValueListNodePtr(ptr)
}

func (p ValueListNodePtr) getNodeData() *ValueListNode {
	ret := ValueListNode{}

	err := MemPtr(p).Get(&ret)
	if err != nil {
		panic(err)
	}

	return &ret
}

func (p BucketListNodePtr) getNodeData() *BucketListNode {
	ret := BucketListNode{}

	err := MemPtr(p).Get(&ret)
	if err != nil {
		panic(err)
	}

	return &ret
}

func (p ValueListNodePtr) setNodeData(v *ValueListNode) {
	MemPtr(p).Set(*v)
}

func (p BucketListNodePtr) setNodeData(b *BucketListNode) {
	MemPtr(p).Set(*b)
}

func (p BucketListNodePtr) appendNode(newNode *BucketListNode) {
	n := p.getNodeData()
	nodeMemPtr := p
	for n.Next != 0x0 {
		n = n.Next.getNodeData()
		nodeMemPtr = n.Next
	}

	n.Next = NewBucketListNode(&BucketListNode{
		Hash:             newNode.Hash,
		ValueListNodePtr: newNode.ValueListNodePtr,
		Next:             0,
	})

	MemPtr(nodeMemPtr).Set(*n)
}

func (p ValueListNodePtr) appendNode(newNode *ValueListNode) {
	n := p.getNodeData()
	nodeMemPtr := p
	for n.Next != 0x0 {
		n = n.Next.getNodeData()
		nodeMemPtr = n.Next
	}

	n.Next = NewValueListNode(newNode.Value)
	MemPtr(nodeMemPtr).Set(*n)
}

func Set(key string, val uint32) {
	hash := murmur3(key, h.Seed)
	bucketNo := hash % SBucketNum

	nodeMemPtr := h.Buckets[bucketNo]
	if nodeMemPtr == 0x0 {
		h.Buckets[bucketNo] = NewBucketListNode(&BucketListNode{
			Hash:             hash,
			ValueListNodePtr: NewValueListNode(val),
			Next:             0x0,
		})
		return
	}

	// TODO: Add sorting on key insertions
	for nodeMemPtr != 0x0 {
		nodeMem := nodeMemPtr.getNodeData()
		// TODO: No hash collision resolution
		if nodeMem.Hash == hash {
			nodeMem.ValueListNodePtr.appendNode(&ValueListNode{Value: val})
			return
		}

		if nodeMem.Next == 0x0 {
			nodeMem.Next = NewBucketListNode(&BucketListNode{
				Hash:             hash,
				ValueListNodePtr: NewValueListNode(val),
			})
			nodeMemPtr.setNodeData(nodeMem)
		}
		nodeMemPtr = nodeMem.Next
	}
}

func GetFirstVal(key string) (uint32, bool) {
	hash := murmur3(key, h.Seed)
	bucketNo := hash % SBucketNum

	nodeMemPtr := h.Buckets[bucketNo]
	for nodeMemPtr != 0x0 {
		nodeMem := nodeMemPtr.getNodeData()

		// TODO: No hash collision resolution
		if nodeMem.Hash == hash {
			return nodeMem.ValueListNodePtr.getNodeData().Value, true
		}

		nodeMemPtr = nodeMem.Next
	}

	return 0, false
}

func PrintMemoryUtilization() {
	memAlloc.PrintMemoryUtilization()
}

func DumpMemoryIntoFile(filepath string) error {
	hPtr.Set(h)
	return memAllocDumpMemoryIntoFile(filepath)
}

func ResetAndInitFromFile(filepath string) error {
	err := memAllocResetNInitFromFile(filepath)
	if err != nil {
		return err
	}

	hPtr.Get(h)
	return nil
}
