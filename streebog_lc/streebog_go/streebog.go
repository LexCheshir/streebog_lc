package main

import (
	"C"
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

const maxUint8 = ^uint8(0)
const maxUint64 = ^uint64(0)
const chunk_size = 64

func add_512(a, b, res *[64]byte) {
	var tmp uint16
	tmp = 0
	for i := range res {
		ind := 63 - i
		tmp = uint16(a[ind]) + uint16(b[ind]) + (tmp >> 8)
		(*res)[ind] = byte(tmp)
	}
}

func transformX(a, b, res *[64]byte) {
	for i := range res {
		(*res)[i] = a[i] ^ b[i]
	}
}

func transformS(res *[64]byte) {
	for i, byteVal := range res {
		(*res)[i] = PI[byteVal]
	}
}

func transformP(res *[64]byte) {
	temp := (*res)
	for i, byteVal := range TAU {
		(*res)[i] = temp[byteVal]
	}
}

func join_bytes(data *[64]byte) (res [8]uint64) {
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

func split_bytes_into(data *[8]uint64, res *[64]byte) {
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

func transformL(res *[64]byte) {
	var buffers [8]uint64
	input64 := join_bytes(res)

	for i := 0; i < 8; i++ {
		for j := 0; j < 64; j++ {
			if (input64[i]>>j)&1 == 1 {
				buffers[i] ^= A[63-j]
			}
		}
	}
	split_bytes_into(&buffers, res)
}

func keySchedule(keys *[64]byte, iter_index int) {
	transformX(keys, &Cn[iter_index], keys)
	transformS(keys)
	transformP(keys)
	transformL(keys)
}

func transformE(keys, chunk, state *[64]byte) {
	transformX(chunk, keys, state)
	for i := 0; i < 12; i++ {
		transformS(state)
		transformP(state)
		transformL(state)
		keySchedule(keys, i)
		transformX(state, keys, state)
	}
}

func transformG(n, hash, message *[64]byte) {
	var keys, temp [64]byte

	transformX(n, hash, &keys)

	transformS(&keys)
	transformP(&keys)
	transformL(&keys)

	transformE(&keys, message, &temp)

	transformX(&temp, hash, &temp)
	transformX(&temp, message, hash)
}

type Streebog struct {
	use256     bool
	hash       [64]byte
	n          [64]byte
	sigma      [64]byte
	block_size [64]byte
	block      [64]byte
}

func (sb *Streebog) updateChunk(b []byte, size *[]byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	sb.block = [64]byte(b)
	sb.block_size[62] = (*size)[1]
	sb.block_size[63] = (*size)[0]

	transformG(&sb.n, &sb.hash, &sb.block)

	add_512(&sb.n, &sb.block_size, &sb.n)
	add_512(&sb.sigma, &sb.block, &sb.sigma)
}

func (sb *Streebog) update(src []byte) {
	buf := bytes.NewBuffer(src)
	size := []byte{0x02, 0x00}

	for buf.Len() >= chunk_size {
		chunk := buf.Next(chunk_size)
		sb.updateChunk(chunk, &size)
	}

	bl := buf.Len()
	binary.LittleEndian.PutUint16(size, uint16(bl)*8)
	if bl > 0 {
		pad := make([]byte, chunk_size-bl)
		data := append(buf.Next(buf.Len()), pad...)
		data[bl] = 1
		sb.updateChunk(data, &size)
	}
}

func (sb *Streebog) digest() []byte {
	var z [64]byte
	transformG(&z, &sb.hash, &sb.n)

	transformG(&z, &sb.hash, &sb.sigma)

	for i, j := 0, len(sb.hash)-1; i < j; i, j = i+1, j-1 {
		sb.hash[i], sb.hash[j] = sb.hash[j], sb.hash[i]
	}

	if sb.use256 {
		return sb.hash[:32]
	}
	return sb.hash[:]
}

func init_streebog(use256 bool) *Streebog {
	if use256 {
		fmt.Println("Using 256")
	} else {
		fmt.Println("Using 512")
	}

	var hash, n, sigma, block_size, block [chunk_size]byte
	if use256 {
		for i := range hash {
			hash[i] = 1
		}
	}

	sb := Streebog{use256, hash, n, sigma, block_size, block}

	return &sb
}

func main() {
	sb := init_streebog(false)

	input := []byte("hello world hello world")

	// input, err := os.ReadFile("/mnt/d/OS/balenaEtcher-Portable-1.7.9.exe")
	// if err != nil {
	// 	panic(err)
	// }

	start := time.Now()

	sb.update(input)
	res := sb.digest()
	fmt.Printf("\nres: %x\n", res)

	elapsed := time.Since(start)
	fmt.Printf("\ntime: %v\n", elapsed)
}
