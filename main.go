package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"math/bits"
	"os"
	"strings"
)

const demoText = "ABCDEFGHABCDEFGH12345678abcdefghijklmnop"

// GenerateIV выдаёт T(0)..T(7) датчика варианта 25.
func GenerateIV() []byte {
	iv := make([]byte, 8)
	state := 20
	for i := range iv {
		iv[i] = byte(state)
		state = (21*state + 33) % 256
	}
	return iv
}

func getBit(data []byte, position int) uint64 {
	return uint64((data[position/8] >> uint(7-position%8)) & 1)
}

// CryptCFB использует прямое DES в обоих направлениях и обратную связь
// по шифртексту. Последний сегмент может быть короче k; дополнение не нужно.
func (d *DES) CryptCFB(data, iv []byte, k int, decrypt bool) ([]byte, error) {
	if d == nil {
		return nil, errors.New("ключ DES не установлен")
	}
	if len(iv) != 8 {
		return nil, errors.New("начальный вектор должен содержать 8 байтов")
	}
	if k < 2 || k > 62 {
		return nil, errors.New("размер сегмента должен быть от 2 до 62 бит")
	}
	var register uint64
	for _, b := range iv {
		register = register<<8 | uint64(b)
	}
	result := make([]byte, len(data))
	for offset := 0; offset < len(data)*8; offset += k {
		size := k
		if remaining := len(data)*8 - offset; remaining < size {
			size = remaining
		}
		stream := d.EncryptBlock(register)
		var ciphertext uint64
		for j := 0; j < size; j++ {
			input := getBit(data, offset+j)
			output := input ^ ((stream >> uint(63-j)) & 1)
			result[(offset+j)/8] |= byte(output) << uint(7-(offset+j)%8)
			feedback := output
			if decrypt {
				feedback = input
			}
			ciphertext = ciphertext<<1 | feedback
		}
		// Сдвиг не циклический: вытесненные старшие биты отбрасываются.
		register = register<<uint(size) | ciphertext
	}
	return result, nil
}

func flipBit(data []byte, position int) ([]byte, error) {
	if position < 0 || position >= len(data)*8 {
		return nil, errors.New("индекс изменяемого бита вне шифрограммы")
	}
	result := append([]byte(nil), data...)
	result[position/8] ^= 1 << uint(7-position%8)
	return result, nil
}

func differentBits(a, b []byte) int {
	count := 0
	for i := range a {
		count += bits.OnesCount8(a[i] ^ b[i])
	}
	return count
}

func experiment(d *DES, text string, iv []byte, k, bit int) error {
	plain := []byte(text)
	encrypted, err := d.CryptCFB(plain, iv, k, false)
	if err != nil {
		return err
	}
	recovered, err := d.CryptCFB(encrypted, iv, k, true)
	if err != nil {
		return err
	}
	fmt.Printf("Открытый текст: %q\nОткрытый текст (HEX): %X\n", text, plain)
	fmt.Printf("Шифрограмма CFB-%d (HEX): %X\n", k, encrypted)
	fmt.Printf("Восстановленный текст: %q\nСовпадение: %t\n", recovered, bytes.Equal(plain, recovered))
	if !bytes.Equal(plain, recovered) {
		return errors.New("исходный текст не восстановлен")
	}
	fmt.Println("Сравнение по группам из 8 байтов (последняя может быть короче):")
	for i := 0; i < len(plain); i += 8 {
		end := i + 8
		if end > len(plain) {
			end = len(plain)
		}
		fmt.Printf("Группа %d: M=%X C=%X\n", i/8+1, plain[i:end], encrypted[i:end])
	}
	if len(plain) == 0 {
		fmt.Println("Пустая шифрограмма: опыт с изменением бита пропущен.")
		return nil
	}
	changed, err := flipBit(encrypted, bit)
	if err != nil {
		return err
	}
	damaged, err := d.CryptCFB(changed, iv, k, true)
	if err != nil {
		return err
	}
	fmt.Printf("Изменён бит %d (с нуля, от старшего бита первого байта).\n", bit)
	fmt.Printf("Искажённая шифрограмма (HEX): %X\nИскажённый открытый текст (HEX): %X\n", changed, damaged)
	for i := 0; i < len(plain); i += 8 {
		end := i + 8
		if end > len(plain) {
			end = len(plain)
		}
		fmt.Printf("Группа %d: изменилось %d из %d бит\n", i/8+1, differentBits(plain[i:end], damaged[i:end]), (end-i)*8)
	}
	if len(plain) >= 16 && bytes.Equal(plain[:8], plain[8:16]) {
		fmt.Printf("Первые открытые группы равны: да; соответствующие шифрогруппы равны: %t\n", bytes.Equal(encrypted[:8], encrypted[8:16]))
	}
	return nil
}

func run() error {
	text := flag.String("text", "", "открытый текст; без параметра вводится одна строка")
	key := flag.String("key", "133457799BBCDFF1", "ключ из 16 HEX-символов")
	k := flag.Int("k", 8, "размер сегмента CFB от 2 до 62 бит")
	bit := flag.Int("bit", 0, "индекс изменяемого бита с нуля")
	flag.Parse()
	provided := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "text" {
			provided = true
		}
	})
	if !provided {
		fmt.Println("Введите открытый текст (пустой ввод — демонстрационный пример):")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 4096), 4*1024*1024)
		if scanner.Scan() {
			*text = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		if *text == "" {
			*text = demoText
		}
	}
	d, err := NewDES(*key)
	if err != nil {
		return err
	}
	iv := GenerateIV()
	fmt.Printf("Лабораторная №6. DES CFB-%d. Вариант 25\n", *k)
	fmt.Printf("Ключ: %s\nПСЧ: A=21 C=33 T(0)=20 B=256\nIV=T(0)..T(7): %X\n", strings.ToUpper(*key), iv)
	if err := experiment(d, *text, iv, *k, *bit); err != nil {
		return err
	}
	if *text != demoText {
		fmt.Println("\nОтдельный опыт с одинаковыми блоками и изменением бита 10:")
		return experiment(d, demoText, iv, *k, 10)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}
