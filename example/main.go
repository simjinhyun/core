package main

import (
	"fmt"

	"github.com/simjinhyun/x"
)

func initialize() {
	fmt.Println("Initialize: DB 연결, 캐시 준비 등")
}
func finalize() {
	fmt.Println("Finalize: 리소스 정리, 로그 flush 등")
}
func handler(c *x.Context) {
	fmt.Fprintln(c.Res, "YEP")
}
func OnShutdownErr(err error) {
	fmt.Println(err)
}
func main() {
	x.NewApp(
		initialize,
		handler,
		finalize,
		OnShutdownErr,
	).Run()
}

// var aeads []cipher.AEAD
// var rrIdx uint32 // 라운드 로빈 인덱스 (atomic)

// func init() {
// 	// 5개의 고정 키
// 	keys := [][]byte{
// 		[]byte("0123456789ABCDEF0123456789ABCDEF"),
// 		[]byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
// 		[]byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
// 		[]byte("CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
// 		[]byte("DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD"),
// 	}

// 	// 각 키마다 aead 생성
// 	for _, key := range keys {
// 		block, err := aes.NewCipher(key)
// 		if err != nil {
// 			panic(err)
// 		}
// 		a, err := cipher.NewGCM(block)
// 		if err != nil {
// 			panic(err)
// 		}
// 		aeads = append(aeads, a)
// 	}
// }

// func encrypt(plaintext []byte) []byte {
// 	// 라운드 로빈 인덱스 선택
// 	idx := atomic.AddUint32(&rrIdx, 1) % uint32(len(aeads))
// 	aead := aeads[idx]

// 	nonce := make([]byte, aead.NonceSize())
// 	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
// 		panic(err)
// 	}
// 	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

// 	// 암호문 앞에 idx를 1바이트로 붙여서 어떤 키인지 표시
// 	return append([]byte{byte(idx)}, append(nonce, ciphertext...)...)
// }

// func decrypt(output []byte) []byte {
// 	idx := int(output[0]) // 첫 바이트가 key index
// 	aead := aeads[idx]

// 	ns := aead.NonceSize()
// 	nonce, ct := output[1:1+ns], output[1+ns:]
// 	plaintext, err := aead.Open(nil, nonce, ct, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return plaintext
// }

// func main() {
// 	msg := []byte("Hello, Round Robin AES!")
// 	ct := encrypt(msg)
// 	fmt.Printf("암호문: %x\n", ct)

// 	pt := decrypt(ct)
// 	fmt.Printf("복호문: %s\n", pt)
// }
