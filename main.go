package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"flag"

	"fmt"
	"os"
	"strings"
)

// ==========================================
// Структура BitArray и методы (из Лабы 4)
// ==========================================

type BitArray struct {
	bits []bool
}

func NewBitArray(size int) *BitArray {
	return &BitArray{bits: make([]bool, size)}
}

func NewBitArrayFromHex(hexStr string) *BitArray {
	ba := &BitArray{}
	ba.FromHex(hexStr)
	return ba
}

func (ba *BitArray) Size() int {
	return len(ba.bits)
}

func (ba *BitArray) Get(index int) bool {
	if index < 0 || index >= len(ba.bits) {
		return false
	}
	return ba.bits[index]
}

func (ba *BitArray) Set(index int, value bool) {
	if index >= 0 && index < len(ba.bits) {
		ba.bits[index] = value
	}
}

func (ba *BitArray) Flip(index int) {
	if index >= 0 && index < len(ba.bits) {
		ba.bits[index] = !ba.bits[index]
	}
}

func (ba *BitArray) XOR(other *BitArray) *BitArray {
	maxSize := len(ba.bits)
	if len(other.bits) > maxSize {
		maxSize = len(other.bits)
	}
	result := NewBitArray(maxSize)
	for i := 0; i < maxSize; i++ {
		result.Set(i, ba.Get(i) != other.Get(i))
	}
	return result
}

func (ba *BitArray) Substring(start, length int) *BitArray {
	result := NewBitArray(length)
	for i := 0; i < length && (start+i) < len(ba.bits); i++ {
		result.Set(i, ba.bits[start+i])
	}
	return result
}

func (ba *BitArray) Concatenate(other *BitArray) {
	ba.bits = append(ba.bits, other.bits...)
}

func (ba *BitArray) CyclicLeftShift(positions int) {
	if len(ba.bits) == 0 {
		return
	}
	positions = (positions%len(ba.bits) + len(ba.bits)) % len(ba.bits)
	newBits := make([]bool, len(ba.bits))
	for i := 0; i < len(ba.bits); i++ {
		newBits[i] = ba.bits[(i+positions)%len(ba.bits)]
	}
	ba.bits = newBits
}

func (ba *BitArray) FromHex(hexStr string) {
	ba.bits = make([]bool, 0, len(hexStr)*4)
	for _, c := range hexStr {
		var value int
		if c >= '0' && c <= '9' {
			value = int(c - '0')
		} else if c >= 'A' && c <= 'F' {
			value = int(c - 'A' + 10)
		} else if c >= 'a' && c <= 'f' {
			value = int(c - 'a' + 10)
		}

		for i := 3; i >= 0; i-- {
			ba.bits = append(ba.bits, ((value>>i)&1) == 1)
		}
	}
}

func (ba *BitArray) ToHex() string {
	var sb strings.Builder
	for i := 0; i < len(ba.bits); i += 4 {
		value := 0
		for j := 0; j < 4 && (i+j) < len(ba.bits); j++ {
			bitVal := 0
			if ba.bits[i+j] {
				bitVal = 1
			}
			value = (value << 1) | bitVal
		}
		sb.WriteString(fmt.Sprintf("%X", value))
	}
	return sb.String()
}

func (ba *BitArray) ToBinary() string {
	var sb strings.Builder
	for _, bit := range ba.bits {
		if bit {
			sb.WriteRune('1')
		} else {
			sb.WriteRune('0')
		}
	}
	return sb.String()
}

func (ba *BitArray) Print() {
	fmt.Println(ba.ToBinary())
}

// ==========================================
// Константы DES (IP, E, P, S-box и т.д.)
// ==========================================

var (
	IP = [64]int{
		58, 50, 42, 34, 26, 18, 10, 2, 60, 52, 44, 36, 28, 20, 12, 4,
		62, 54, 46, 38, 30, 22, 14, 6, 64, 56, 48, 40, 32, 24, 16, 8,
		57, 49, 41, 33, 25, 17, 9, 1, 59, 51, 43, 35, 27, 19, 11, 3,
		61, 53, 45, 37, 29, 21, 13, 5, 63, 55, 47, 39, 31, 23, 15, 7,
	}

	IP_1 = [64]int{
		40, 8, 48, 16, 56, 24, 64, 32, 39, 7, 47, 15, 55, 23, 63, 31,
		38, 6, 46, 14, 54, 22, 62, 30, 37, 5, 45, 13, 53, 21, 61, 29,
		36, 4, 44, 12, 52, 20, 60, 28, 35, 3, 43, 11, 51, 19, 59, 27,
		34, 2, 42, 10, 50, 18, 58, 26, 33, 1, 41, 9, 49, 17, 57, 25,
	}

	E = [48]int{
		32, 1, 2, 3, 4, 5, 4, 5, 6, 7, 8, 9, 8, 9, 10, 11, 12, 13,
		12, 13, 14, 15, 16, 17, 16, 17, 18, 19, 20, 21, 20, 21, 22, 23, 24, 25,
		24, 25, 26, 27, 28, 29, 28, 29, 30, 31, 32, 1,
	}

	P = [32]int{
		16, 7, 20, 21, 29, 12, 28, 17, 1, 15, 23, 26, 5, 18, 31, 10,
		2, 8, 24, 14, 32, 27, 3, 9, 19, 13, 30, 6, 22, 11, 4, 25,
	}

	PC1 = [56]int{
		57, 49, 41, 33, 25, 17, 9, 1, 58, 50, 42, 34, 26, 18,
		10, 2, 59, 51, 43, 35, 27, 19, 11, 3, 60, 52, 44, 36,
		63, 55, 47, 39, 31, 23, 15, 7, 62, 54, 46, 38, 30, 22,
		14, 6, 61, 53, 45, 37, 29, 21, 13, 5, 28, 20, 12, 4,
	}

	PC2 = [48]int{
		14, 17, 11, 24, 1, 5, 3, 28, 15, 6, 21, 10,
		23, 19, 12, 4, 26, 8, 16, 7, 27, 20, 13, 2,
		41, 52, 31, 37, 47, 55, 30, 40, 51, 45, 33, 48,
		44, 49, 39, 56, 34, 53, 46, 42, 50, 36, 29, 32,
	}

	LS = [16]int{1, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1}

	S = [8][4][16]int{
		{
			{14, 4, 13, 1, 2, 15, 11, 8, 3, 10, 6, 12, 5, 9, 0, 7},
			{0, 15, 7, 4, 14, 2, 13, 1, 10, 6, 12, 11, 9, 5, 3, 8},
			{4, 1, 14, 8, 13, 6, 2, 11, 15, 12, 9, 7, 3, 10, 5, 0},
			{15, 12, 8, 2, 4, 9, 1, 7, 5, 11, 3, 14, 10, 0, 6, 13},
		},
		{
			{15, 1, 8, 14, 6, 11, 3, 4, 9, 7, 2, 13, 12, 0, 5, 10},
			{3, 13, 4, 7, 15, 2, 8, 14, 12, 0, 1, 10, 6, 9, 11, 5},
			{0, 14, 7, 11, 10, 4, 13, 1, 5, 8, 12, 6, 9, 3, 2, 15},
			{13, 8, 10, 1, 3, 15, 4, 2, 11, 6, 7, 12, 0, 5, 14, 9},
		},
		{
			{10, 0, 9, 14, 6, 3, 15, 5, 1, 13, 12, 7, 11, 4, 2, 8},
			{13, 7, 0, 9, 3, 4, 6, 10, 2, 8, 5, 14, 12, 11, 15, 1},
			{13, 6, 4, 9, 8, 15, 3, 0, 11, 1, 2, 12, 5, 10, 14, 7},
			{1, 10, 13, 0, 6, 9, 8, 7, 4, 15, 14, 3, 11, 5, 2, 12},
		},
		{
			{7, 13, 14, 3, 0, 6, 9, 10, 1, 2, 8, 5, 11, 12, 4, 15},
			{13, 8, 11, 5, 6, 15, 0, 3, 4, 7, 2, 12, 1, 10, 14, 9},
			{10, 6, 9, 0, 12, 11, 7, 13, 15, 1, 3, 14, 5, 2, 8, 4},
			{3, 15, 0, 6, 10, 1, 13, 8, 9, 4, 5, 11, 12, 7, 2, 14},
		},
		{
			{2, 12, 4, 1, 7, 10, 11, 6, 8, 5, 3, 15, 13, 0, 14, 9},
			{14, 11, 2, 12, 4, 7, 13, 1, 5, 0, 15, 10, 3, 9, 8, 6},
			{4, 2, 1, 11, 10, 13, 7, 8, 15, 9, 12, 5, 6, 3, 0, 14},
			{11, 8, 12, 7, 1, 14, 2, 13, 6, 15, 0, 9, 10, 4, 5, 3},
		},
		{
			{12, 1, 10, 15, 9, 2, 6, 8, 0, 13, 3, 4, 14, 7, 5, 11},
			{10, 15, 4, 2, 7, 12, 9, 5, 6, 1, 13, 14, 0, 11, 3, 8},
			{9, 14, 15, 5, 2, 8, 12, 3, 7, 0, 4, 10, 1, 13, 11, 6},
			{4, 3, 2, 12, 9, 5, 15, 10, 11, 14, 1, 7, 6, 0, 8, 13},
		},
		{
			{4, 11, 2, 14, 15, 0, 8, 13, 3, 12, 9, 7, 5, 10, 6, 1},
			{13, 0, 11, 7, 4, 9, 1, 10, 14, 3, 5, 12, 2, 15, 8, 6},
			{1, 4, 11, 13, 12, 3, 7, 14, 10, 15, 6, 8, 0, 5, 9, 2},
			{6, 11, 13, 8, 1, 4, 10, 7, 9, 5, 0, 15, 14, 2, 3, 12},
		},
		{
			{13, 2, 8, 4, 6, 15, 11, 1, 10, 9, 3, 14, 5, 0, 12, 7},
			{1, 15, 13, 8, 10, 3, 7, 4, 12, 5, 6, 11, 0, 14, 9, 2},
			{7, 11, 4, 1, 9, 12, 14, 2, 0, 6, 10, 13, 15, 3, 5, 8},
			{2, 1, 14, 7, 4, 10, 8, 13, 15, 12, 9, 0, 3, 5, 6, 11},
		},
	}
)

type DES struct {
	subkeys []*BitArray
}

func NewDES() *DES {
	return &DES{
		subkeys: make([]*BitArray, 0),
	}
}

// applyPermutation использует нумерацию таблиц DES с единицы.
func (d *DES) applyPermutation(input *BitArray, table []int, outputSize int) *BitArray {
	output := NewBitArray(outputSize)
	for i := 0; i < outputSize; i++ {
		output.Set(i, input.Get(table[i]-1))
	}
	return output
}

// f расширяет R до 48 бит, выполняет XOR с ключом, S-подстановки и P.
func (d *DES) f(R, K *BitArray) *BitArray {
	expanded := d.applyPermutation(R, E[:], 48)
	xored := expanded.XOR(K)

	sboxOutput := NewBitArray(32)
	for i := 0; i < 8; i++ {
		block := xored.Substring(i*6, 6)
		row := 0
		if block.Get(0) {
			row |= 2
		}
		if block.Get(5) {
			row |= 1
		}

		col := 0
		for j := 1; j <= 4; j++ {
			col = (col << 1)
			if block.Get(j) {
				col |= 1
			}
		}

		value := S[i][row][col]
		for j := 3; j >= 0; j-- {
			sboxOutput.Set(i*4+(3-j), ((value>>j)&1) == 1)
		}
	}

	return d.applyPermutation(sboxOutput, P[:], 32)
}

// generateSubkeys удаляет биты чётности PC1 и строит 16 ключей через LS и PC2.
func (d *DES) generateSubkeys(key *BitArray) {
	d.subkeys = make([]*BitArray, 0, 16)
	permutedKey := d.applyPermutation(key, PC1[:], 56)

	L := permutedKey.Substring(0, 28)
	R := permutedKey.Substring(28, 28)

	for i := 0; i < 16; i++ {
		L.CyclicLeftShift(LS[i])
		R.CyclicLeftShift(LS[i])

		combined := NewBitArray(56)
		for j := 0; j < L.Size(); j++ {
			combined.Set(j, L.Get(j))
		}
		for j := 0; j < R.Size(); j++ {
			combined.Set(L.Size()+j, R.Get(j))
		}

		subkey := d.applyPermutation(combined, PC2[:], 48)
		d.subkeys = append(d.subkeys, subkey)
	}
}

// encryptBlock выполняет IP, 16 раундов Фейстеля, обмен половин и IP^-1.
// Дешифрование отличается только обратным порядком раундовых ключей.
func (d *DES) encryptBlock(block *BitArray, decrypt bool) *BitArray {
	permuted := d.applyPermutation(block, IP[:], 64)

	L := permuted.Substring(0, 32)
	R := permuted.Substring(32, 32)

	for i := 0; i < 16; i++ {
		keyIndex := i
		if decrypt {
			keyIndex = 15 - i
		}
		tempL := L
		L = R
		R = tempL.XOR(d.f(R, d.subkeys[keyIndex]))
	}

	combined := NewBitArray(64)
	for j := 0; j < R.Size(); j++ {
		combined.Set(j, R.Get(j))
	}
	for j := 0; j < L.Size(); j++ {
		combined.Set(R.Size()+j, L.Get(j))
	}

	return d.applyPermutation(combined, IP_1[:], 64)
}

// validateHex проверяет HEX и длину до обращения к битам или срезам.
func validateHex(value string, blockChars int, single bool) error {
	if len(value) == 0 || len(value)%blockChars != 0 || (single && len(value) != blockChars) {
		return fmt.Errorf("ожидается %d HEX-символов на блок", blockChars)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("некорректный HEX: %w", err)
	}
	return nil
}

// SetKey принимает 64-битовый ключ, включая восемь битов чётности.
func (d *DES) SetKey(key string) error {
	if err := validateHex(key, 16, true); err != nil {
		return err
	}
	d.generateSubkeys(NewBitArrayFromHex(key))
	return nil
}

// cryptCBC обрабатывает полные блоки; дополнение добавляется отдельно.
func (d *DES) cryptCBC(data, iv string, decrypt bool) (string, error) {
	if len(d.subkeys) != 16 {
		return "", errors.New("ключ DES не установлен")
	}
	if err := validateHex(data, 16, false); err != nil {
		return "", err
	}
	if err := validateHex(iv, 16, true); err != nil {
		return "", err
	}
	previous := NewBitArrayFromHex(iv)
	var result strings.Builder
	for i := 0; i < len(data); i += 16 {
		block := NewBitArrayFromHex(data[i : i+16])
		if decrypt {
			// M_i = D_K(C_i) XOR C_(i-1); в буфер идёт исходный C_i.
			result.WriteString(d.encryptBlock(block, true).XOR(previous).ToHex())
			previous = block
		} else {
			// C_i = E_K(M_i XOR C_(i-1)); первый предыдущий блок равен IV.
			previous = d.encryptBlock(block.XOR(previous), false)
			result.WriteString(previous.ToHex())
		}
	}
	return result.String(), nil
}

func (d *DES) EncryptCBC(data, iv string) (string, error) {
	return d.cryptCBC(data, iv, false)
}
func (d *DES) DecryptCBC(data, iv string) (string, error) {
	return d.cryptCBC(data, iv, true)
}

// generateIV повторяет датчик лабораторной №3 для варианта 25 (строка 5).
// Сначала выдаётся T(0), затем T(i+1)=(21*T(i)+33) mod 256.
func generateIV() string {
	iv := make([]byte, 8)
	state := 20
	for i := range iv {
		iv[i] = byte(state)
		state = (21*state + 33) % 256
	}
	return strings.ToUpper(hex.EncodeToString(iv))
}

func stringToHex(text string) string {
	return strings.ToUpper(hex.EncodeToString([]byte(text)))
}

// padPKCS7 дополняет байты UTF-8; даже кратная 8 длина получает полный блок.
func padPKCS7(text string) string {
	n := 8 - len(text)%8
	return text + strings.Repeat(string([]byte{byte(n)}), n)
}

// unpadPKCS7 проверяет все байты дополнения перед восстановлением текста.
func unpadPKCS7(text string) (string, error) {
	if len(text) == 0 || len(text)%8 != 0 {
		return "", errors.New("неверная длина PKCS#7")
	}
	n := int(text[len(text)-1])
	if n < 1 || n > 8 {
		return "", errors.New("неверная длина дополнения")
	}
	for i := len(text) - n; i < len(text); i++ {
		if int(text[i]) != n {
			return "", errors.New("повреждены байты дополнения")
		}
	}
	return text[:len(text)-n], nil
}

// flipBit нумерует биты с нуля, от старшего бита первого байта.
func flipBit(data string, position int) (string, error) {
	raw, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	if position < 0 || position >= len(raw)*8 {
		return "", errors.New("бит вне шифрограммы")
	}
	raw[position/8] ^= byte(1 << (7 - position%8))
	return strings.ToUpper(hex.EncodeToString(raw)), nil
}

func countDifferentBits(a, b string) int {
	left, right := NewBitArrayFromHex(a), NewBitArrayFromHex(b)
	count := 0
	for i := 0; i < left.Size(); i++ {
		if left.Get(i) != right.Get(i) {
			count++
		}
	}
	return count
}

// experiment показывает шифрование, восстановление и распространение ошибки.
func experiment(d *DES, text, iv string, bit int) error {
	plain := stringToHex(padPKCS7(text))
	encrypted, err := d.EncryptCBC(plain, iv)
	if err != nil {
		return err
	}
	recovered, err := d.DecryptCBC(encrypted, iv)
	if err != nil {
		return err
	}
	raw, err := hex.DecodeString(recovered)
	if err != nil {
		return err
	}
	restored, err := unpadPKCS7(string(raw))
	if err != nil {
		return err
	}
	fmt.Printf("Открытый текст: %q\nОткрытый текст с PKCS#7 (HEX): %s\n", text, plain)
	fmt.Println("Шифрограмма CBC (HEX):", encrypted)
	fmt.Printf("Восстановленный текст: %q\nСовпадение: %t\n", restored, restored == text)
	if recovered != plain || restored != text {
		return errors.New("текст не восстановлен")
	}
	fmt.Println("Поблочное сравнение (64 бита):")
	for i := 0; i < len(plain); i += 16 {
		fmt.Printf("M[%d]=%s C[%d]=%s\n", i/16+1, plain[i:i+16], i/16+1, encrypted[i:i+16])
	}
	changed, err := flipBit(encrypted, bit)
	if err != nil {
		return err
	}
	damaged, err := d.DecryptCBC(changed, iv)
	if err != nil {
		return err
	}
	fmt.Printf("Изменён бит с индексом %d (нумерация с 0).\n", bit)
	fmt.Println("Искажённая шифрограмма (HEX):", changed)
	// HEX сохраняет все повреждённые байты, в том числе недопустимые в UTF-8.
	fmt.Println("Искажённый открытый текст с дополнением (HEX):", damaged)
	for i := 0; i < len(plain); i += 16 {
		fmt.Printf("Блок %d: изменилось %d из 64 бит\n", i/16+1,
			countDifferentBits(plain[i:i+16], damaged[i:i+16]))
	}
	if len(plain) >= 32 && plain[:16] == plain[16:32] {
		fmt.Printf("M[1]=M[2]: true; C[1]=C[2]: %t\n", encrypted[:16] == encrypted[16:32])
	}
	return nil
}

func run() error {
	text := flag.String("text", "", "открытый текст; без флага вводится с клавиатуры")
	key := flag.String("key", "133457799BBCDFF1", "ключ из 16 HEX-символов")
	bit := flag.Int("bit", 10, "индекс изменяемого бита, начиная с нуля")
	flag.Parse()
	supplied := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "text" {
			supplied = true
		}
	})
	if !supplied {
		fmt.Println("Введите открытый текст (Enter — пример ABCDEFGHABCDEFGH12345678):")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 4096), 4*1024*1024)
		if scanner.Scan() {
			*text = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		if *text == "" {
			*text = "ABCDEFGHABCDEFGH12345678"
		}
	}
	d := NewDES()
	if err := d.SetKey(*key); err != nil {
		return err
	}
	iv := generateIV()
	fmt.Println("Лабораторная работа №5. DES CBC. Вариант 25")
	fmt.Println("ПСЧ: A=21 C=33 T(0)=20 B=256; IV=T(0)..T(7)")
	fmt.Println("Ключ (HEX):", strings.ToUpper(*key))
	fmt.Println("IV (HEX):", iv)
	// Для короткого произвольного ввода показываем ошибку в его пределах.
	// Для отдельных опытов ниже используются минимум три полных блока.
	if err := experiment(d, *text, iv, *bit); err != nil {
		return err
	}
	if *text != "ABCDEFGHABCDEFGH12345678" {
		fmt.Println("\nОтдельный опыт: идентичные блоки и ошибка в первом блоке")
		return experiment(d, "ABCDEFGHABCDEFGH12345678", iv, 10)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}
