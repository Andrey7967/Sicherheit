package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// Сравнение с независимой реализацией Go обнаруживает ошибки таблиц и раундов.
func TestDESKnownVector(t *testing.T) {
	d := NewDES()
	if err := d.SetKey("133457799BBCDFF1"); err != nil {
		t.Fatal(err)
	}
	got := d.encryptBlock(NewBitArrayFromHex("0123456789ABCDEF"), false).ToHex()
	if got != "85E813540F0AB405" {
		t.Fatalf("DES: %s", got)
	}
	if back := d.encryptBlock(NewBitArrayFromHex(got), true).ToHex(); back != "0123456789ABCDEF" {
		t.Fatal(back)
	}
}

func TestCBCMatchesStandardLibrary(t *testing.T) {
	random := rand.New(rand.NewSource(25))
	for n := 1; n <= 40; n++ {
		key, iv, input := make([]byte, 8), make([]byte, 8), make([]byte, n*8)
		random.Read(key)
		random.Read(iv)
		random.Read(input)
		standard, err := des.NewCipher(key)
		if err != nil {
			t.Fatal(err)
		}
		expected := make([]byte, len(input))
		cipher.NewCBCEncrypter(standard, iv).CryptBlocks(expected, input)
		d := NewDES()
		if err := d.SetKey(hex.EncodeToString(key)); err != nil {
			t.Fatal(err)
		}
		got, err := d.EncryptCBC(hex.EncodeToString(input), hex.EncodeToString(iv))
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := hex.DecodeString(got)
		if !bytes.Equal(raw, expected) {
			t.Fatalf("CBC differs at %d blocks", n)
		}
		back, err := d.DecryptCBC(got, hex.EncodeToString(iv))
		if err != nil || back != strings.ToUpper(hex.EncodeToString(input)) {
			t.Fatalf("round trip %d: %v", n, err)
		}
	}
}

func TestVariantIVAndErrorPropagation(t *testing.T) {
	if generateIV() != "14C54A3350B1A6BF" {
		t.Fatal(generateIV())
	}
	d := NewDES()
	if err := d.SetKey("133457799BBCDFF1"); err != nil {
		t.Fatal(err)
	}
	plain := stringToHex(padPKCS7("ABCDEFGHABCDEFGH12345678"))
	encrypted, err := d.EncryptCBC(plain, generateIV())
	if err != nil {
		t.Fatal(err)
	}
	if encrypted[:16] == encrypted[16:32] {
		t.Fatal("identical CBC blocks")
	}
	// Проверяем каждый бит первых двух блоков, а не только пример отчёта.
	for bit := 0; bit < 128; bit++ {
		changed, err := flipBit(encrypted, bit)
		if err != nil {
			t.Fatal(err)
		}
		back, err := d.DecryptCBC(changed, generateIV())
		if err != nil {
			t.Fatal(err)
		}
		block := bit / 64
		if countDifferentBits(plain[(block+1)*16:(block+2)*16], back[(block+1)*16:(block+2)*16]) != 1 {
			t.Fatal(bit)
		}
		for i := 0; i < 4; i++ {
			if i != block && i != block+1 && plain[i*16:(i+1)*16] != back[i*16:(i+1)*16] {
				t.Fatal(bit, i)
			}
		}
	}
}

func TestPaddingAndInvalidInput(t *testing.T) {
	for _, text := range []string{"", "a", "ABCDEFGH", "Привет, мир!", "🙂\x00"} {
		padded := padPKCS7(text)
		got, err := unpadPKCS7(padded)
		if err != nil || got != text {
			t.Fatalf("%q: %v", text, err)
		}
	}
	for _, raw := range []string{"", "123", "1234567\x00", "123456\x01\x02"} {
		if _, err := unpadPKCS7(raw); err == nil {
			t.Fatal(fmt.Sprintf("accepted %q", raw))
		}
	}
	d := NewDES()
	if _, err := d.EncryptCBC("0000000000000000", generateIV()); err == nil {
		t.Fatal("unset key")
	}
	for _, key := range []string{"", "1234", "ZZZZZZZZZZZZZZZZ"} {
		if err := d.SetKey(key); err == nil {
			t.Fatal("invalid key", key)
		}
	}
	if err := d.SetKey("133457799BBCDFF1"); err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{"", "123", "ZZZZZZZZZZZZZZZZ"} {
		if _, err := d.EncryptCBC(data, generateIV()); err == nil {
			t.Fatal("invalid data", data)
		}
	}
	if _, err := d.DecryptCBC("0000000000000000", "AA"); err == nil {
		t.Fatal("invalid IV")
	}
	for _, bit := range []int{-1, 64} {
		if _, err := flipBit("0000000000000000", bit); err == nil {
			t.Fatal("invalid bit", bit)
		}
	}
}
