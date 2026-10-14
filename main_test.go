package main

import (
	"bytes"
	"crypto/des"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"testing"
)

func TestDESAgainstStandardLibrary(t *testing.T) {
	d, err := NewDES("133457799BBCDFF1")
	if err != nil {
		t.Fatal(err)
	}
	if got := d.EncryptBlock(0x0123456789ABCDEF); got != 0x85E813540F0AB405 {
		t.Fatalf("контрольный вектор DES: %016X", got)
	}
	if d.subkeys[0] != 0x1B02EFFC7072 || d.subkeys[15] != 0xCB3D8B0E17F5 {
		t.Fatal("контрольные раундовые ключи")
	}
	r := rand.New(rand.NewSource(25))
	for i := 0; i < 256; i++ {
		key, input := make([]byte, 8), make([]byte, 8)
		r.Read(key)
		r.Read(input)
		standard, _ := des.NewCipher(key)
		expected := make([]byte, 8)
		standard.Encrypt(expected, input)
		custom, _ := NewDES(hex.EncodeToString(key))
		got := custom.EncryptBlock(binary.BigEndian.Uint64(input))
		if got != binary.BigEndian.Uint64(expected) {
			t.Fatalf("DES, случай %d", i)
		}
	}
}

// Независимый эталон использует стандартный DES и явный массив из 64 бит.
// Он не вызывает функции и преобразования проверяемой реализации CFB.
func referenceCFB(key, iv, data []byte, k int, decrypt bool) []byte {
	block, _ := des.NewCipher(key)
	register := make([]byte, 64)
	for i := range register {
		register[i] = (iv[i/8] >> uint(7-i%8)) & 1
	}
	result := make([]byte, len(data))
	for offset := 0; offset < len(data)*8; offset += k {
		size := k
		if size > len(data)*8-offset {
			size = len(data)*8 - offset
		}
		input, stream := make([]byte, 8), make([]byte, 8)
		for i, b := range register {
			input[i/8] |= b << uint(7-i%8)
		}
		block.Encrypt(stream, input)
		feedback := make([]byte, size)
		for j := 0; j < size; j++ {
			in := (data[(offset+j)/8] >> uint(7-(offset+j)%8)) & 1
			out := in ^ ((stream[j/8] >> uint(7-j%8)) & 1)
			result[(offset+j)/8] |= out << uint(7-(offset+j)%8)
			feedback[j] = out
			if decrypt {
				feedback[j] = in
			}
		}
		register = append(append([]byte(nil), register[size:]...), feedback...)
	}
	return result
}

func TestCFBAllSegmentSizes(t *testing.T) {
	r := rand.New(rand.NewSource(6))
	for k := 2; k <= 62; k++ {
		for _, length := range []int{0, 1, 2, 7, 8, 9, 17, 40} {
			key, iv, input := make([]byte, 8), make([]byte, 8), make([]byte, length)
			r.Read(key)
			r.Read(iv)
			r.Read(input)
			d, _ := NewDES(hex.EncodeToString(key))
			got, err := d.CryptCFB(input, iv, k, false)
			if err != nil || !bytes.Equal(got, referenceCFB(key, iv, input, k, false)) {
				t.Fatalf("CFB-%d, длина %d: %v", k, length, err)
			}
			back, err := d.CryptCFB(got, iv, k, true)
			if err != nil || !bytes.Equal(back, input) || !bytes.Equal(back, referenceCFB(key, iv, got, k, true)) {
				t.Fatalf("дешифрование CFB-%d, длина %d", k, length)
			}
			if len(got) != len(input) {
				t.Fatal("изменилась длина")
			}
		}
	}
}

func TestCFB8ErrorPropagation(t *testing.T) {
	d, _ := NewDES("133457799BBCDFF1")
	plain := []byte(demoText)
	iv := GenerateIV()
	encrypted, _ := d.CryptCFB(plain, iv, 8, false)
	for bit := 0; bit < len(plain)*8; bit++ {
		changed, _ := flipBit(encrypted, bit)
		back, err := d.CryptCFB(changed, iv, 8, true)
		if err != nil {
			t.Fatal(err)
		}
		position := bit / 8
		if back[position] != plain[position]^(1<<uint(7-bit%8)) {
			t.Fatalf("прямая ошибка: %d", bit)
		}
		for i := range plain {
			if (i < position || i > position+8) && plain[i] != back[i] {
				t.Fatalf("граница распространения: %d, %d", bit, i)
			}
		}
	}
}

func TestVariantAndReportExample(t *testing.T) {
	if got := hex.EncodeToString(GenerateIV()); got != "14c54a3350b1a6bf" {
		t.Fatal(got)
	}
	d, _ := NewDES("133457799BBCDFF1")
	encrypted, _ := d.CryptCFB([]byte(demoText), GenerateIV(), 8, false)
	if hex.EncodeToString(encrypted) != "119e5df1a26641e8d3f421ad66a3a734ed5cbe2c866e87e2f7353bf10f4663d4d07169c7633ad514" {
		t.Fatal("пример отчёта")
	}
	if bytes.Equal(encrypted[:8], encrypted[8:16]) {
		t.Fatal("совпали шифрогруппы")
	}
	changed, _ := flipBit(encrypted, 10)
	back, _ := d.CryptCFB(changed, GenerateIV(), 8, true)
	if hex.EncodeToString(back) != "41622e8ae18208e0756743444546474831323334353637386162636465666768696a6b6c6d6e6f70" {
		t.Fatal("искажённый текст")
	}
	t.Logf("Первый выход DES: %016X", d.EncryptBlock(0x14C54A3350B1A6BF))
}

func TestTextRoundTrip(t *testing.T) {
	d, _ := NewDES("133457799BBCDFF1")
	for _, text := range []string{"", "a", "ABCDEFGH", "Привет, мир!", "🙂\x00", "строка 1\nстрока 2"} {
		for _, k := range []int{7, 8, 16, 62} {
			input := []byte(text)
			encrypted, err := d.CryptCFB(input, GenerateIV(), k, false)
			if err != nil {
				t.Fatal(err)
			}
			back, err := d.CryptCFB(encrypted, GenerateIV(), k, true)
			if err != nil || !bytes.Equal(input, back) {
				t.Fatalf("%q, k=%d", text, k)
			}
		}
	}
}

func TestInvalidInput(t *testing.T) {
	for _, key := range []string{"", "1234", "ZZZZZZZZZZZZZZZZ"} {
		if _, err := NewDES(key); err == nil {
			t.Fatal("принят неверный ключ")
		}
	}
	d, _ := NewDES("133457799BBCDFF1")
	for _, k := range []int{-1, 0, 1, 63, 64} {
		if _, err := d.CryptCFB(nil, GenerateIV(), k, false); err == nil {
			t.Fatal("неверный размер сегмента")
		}
	}
	if _, err := d.CryptCFB([]byte{0}, []byte{0}, 8, false); err == nil {
		t.Fatal("неверный IV")
	}
	var missing *DES
	if _, err := missing.CryptCFB(nil, GenerateIV(), 8, false); err == nil {
		t.Fatal("ключ отсутствует")
	}
	for _, bit := range []int{-1, 8} {
		if _, err := flipBit([]byte{0}, bit); err == nil {
			t.Fatal("неверный бит")
		}
	}
	if _, err := flipBit(nil, 0); err == nil {
		t.Fatal("пустая шифрограмма")
	}
}
