package main

import (
	"C"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const maxUint8 = ^uint8(0)
const maxUint64 = ^uint64(0)
const chunk_size = 64

func Add512(a, b, res *[64]byte) {
	var tmp uint16
	tmp = 0
	for i := range res {
		ind := 63 - i
		tmp = uint16(a[ind]) + uint16(b[ind]) + (tmp >> 8)
		(*res)[ind] = byte(tmp)
	}
}

func TransformX(a, b, res *[64]byte) {
	for i := range res {
		(*res)[i] = a[i] ^ b[i]
	}
}

func TransformS(res *[64]byte) {
	for i, byteVal := range res {
		(*res)[i] = PI[byteVal]
	}
}

func TransformP(res *[64]byte) {
	temp := (*res)
	for i, byteVal := range TAU {
		(*res)[i] = temp[byteVal]
	}
}

func JoinBytes(data *[64]byte) (res [8]uint64) {
	tmp := uint64(0)
	for i, byteVal := range data {
		tmp += uint64(byteVal)
		if (i+1)%8 == 0 {
			res[i/8] = tmp
			tmp = uint64(0)
			continue
		}
		tmp = tmp << 8
	}
	return
}

func SplitBytesInto(data *[8]uint64, res *[64]byte) {
	i := 0
	for _, el := range data {
		tmp := make([]byte, 8)
		binary.LittleEndian.PutUint64(tmp, el)
		for j := range tmp {
			(*res)[i] = tmp[7-j]
			i++
		}
	}
}

func TransformL(res *[64]byte) {
	var buffers [8]uint64
	input64 := JoinBytes(res)

	for i := 0; i < 8; i++ {
		for j := 0; j < 64; j++ {
			if (input64[i]>>j)&1 == 1 {
				buffers[i] ^= A[63-j]
			}
		}
	}
	SplitBytesInto(&buffers, res)
}

func KeySchedule(keys *[64]byte, iter_index int) {
	TransformX(keys, &Cn[iter_index], keys)
	TransformS(keys)
	TransformP(keys)
	TransformL(keys)
}

func TransformE(keys, chunk, state *[64]byte) {
	TransformX(chunk, keys, state)
	for i := 0; i < 12; i++ {
		TransformS(state)
		TransformP(state)
		TransformL(state)
		KeySchedule(keys, i)
		TransformX(state, keys, state)
	}
}

func TransformG(n, hash, message *[64]byte) {
	var keys, temp [64]byte

	TransformX(n, hash, &keys)

	TransformS(&keys)
	TransformP(&keys)
	TransformL(&keys)

	TransformE(&keys, message, &temp)

	TransformX(&temp, hash, &temp)
	TransformX(&temp, message, hash)
}

type Streebog struct {
	use256     bool
	hash       [64]byte
	n          [64]byte
	sigma      [64]byte
	block_size [64]byte
	block      [64]byte
}

func (sb *Streebog) UpdateChunk(b []byte, size *[]byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	sb.block = [64]byte(b)
	sb.block_size[62] = (*size)[1]
	sb.block_size[63] = (*size)[0]

	TransformG(&sb.n, &sb.hash, &sb.block)

	Add512(&sb.n, &sb.block_size, &sb.n)
	Add512(&sb.sigma, &sb.block, &sb.sigma)
}

func (sb *Streebog) Update(src []byte) {
	buf := bytes.NewBuffer(src)
	size := []byte{0x02, 0x00}

	for buf.Len() >= chunk_size {
		chunk := buf.Next(chunk_size)
		sb.UpdateChunk(chunk, &size)
	}

	bl := buf.Len()
	binary.LittleEndian.PutUint16(size, uint16(bl)*8)
	if bl > 0 {
		pad := make([]byte, chunk_size-bl)
		data := append(buf.Next(buf.Len()), pad...)
		data[bl] = 1
		sb.UpdateChunk(data, &size)
	}
}

func (sb *Streebog) Digest() []byte {
	var z [64]byte
	TransformG(&z, &sb.hash, &sb.n)

	TransformG(&z, &sb.hash, &sb.sigma)

	for i, j := 0, len(sb.hash)-1; i < j; i, j = i+1, j-1 {
		sb.hash[i], sb.hash[j] = sb.hash[j], sb.hash[i]
	}

	if sb.use256 {
		return sb.hash[:32]
	}
	return sb.hash[:]
}

func InitStreebog(use256 bool) *Streebog {
	var hash, n, sigma, block_size, block [chunk_size]byte
	if use256 {
		for i := range hash {
			hash[i] = 1
		}
	}
	sb := Streebog{use256, hash, n, sigma, block_size, block}
	return &sb
}

func HashBytes(input []byte, use256 bool) string {
	sb := InitStreebog(use256)
	sb.Update(input)
	res := sb.Digest()
	return fmt.Sprintf("%x", res)
}

func HashFile(path string, use256 bool) string {
	input, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return HashBytes(input, use256)
}

//export HashFileWrapper
func HashFileWrapper(pathPtr *C.char) *C.char {
	path := C.GoString(pathPtr)
	fmt.Println(path)
	return C.CString(path)
}

func main() {
	sb := InitStreebog(false)

	input := []byte("hello world hello world")

	// input, err := os.ReadFile("/mnt/d/OS/balenaEtcher-Portable-1.7.9.exe")
	// if err != nil {
	// 	panic(err)
	// }

	start := time.Now()

	sb.Update(input)
	res := sb.Digest()
	fmt.Printf("\nres: %x\n", res)

	elapsed := time.Since(start)
	fmt.Printf("\ntime: %v\n", elapsed)
}
