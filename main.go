package main

import (
	"bufio"

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
	if index >= len(ba.bits) {
		return false
	}
	return ba.bits[index]
}

func (ba *BitArray) Set(index int, value bool) {
	if index < len(ba.bits) {
		ba.bits[index] = value
	}
}

func (ba *BitArray) Flip(index int) {
	if index < len(ba.bits) {
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
	positions %= len(ba.bits)
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
			{7, 11, 4, 1, 9, 12, 14, 2, 0, 6, 10, 13, 15, 13, 5, 8},
			{2, 1, 14, 7, 4, 10, 8, 13, 5, 12, 9, 0, 3, 5, 5, 11},
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

func (d *DES) applyPermutation(input *BitArray, table []int, outputSize int) *BitArray {
	output := NewBitArray(outputSize)
	for i := 0; i < outputSize; i++ {
		output.Set(i, input.Get(table[i]-1))
	}
	return output
}

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

func (d *DES) SetKey(hexKey string) {
	key := NewBitArrayFromHex(hexKey)
	d.generateSubkeys(key)
}

func (d *DES) EncryptBlock(hexPlaintext string) string {
	plaintext := NewBitArrayFromHex(hexPlaintext)
	ciphertext := d.encryptBlock(plaintext, false)
	return ciphertext.ToHex()
}

func (d *DES) DecryptBlock(hexCiphertext string) string {
	ciphertext := NewBitArrayFromHex(hexCiphertext)
	plaintext := d.encryptBlock(ciphertext, true)
	return plaintext.ToHex()
}

func (d *DES) FlipBit(hexCiphertext string, bitPosition int) string {
	ciphertext := NewBitArrayFromHex(hexCiphertext)
	ciphertext.Flip(bitPosition)
	return ciphertext.ToHex()
}

// ==========================================
// НОВОЕ для Лабы 5: Реализация режима CBC
// ==========================================

// EncryptCBC шифрует всю строку (в HEX, кратную 16 символам) в режиме CBC
func (d *DES) EncryptCBC(hexPlaintext string, hexIV string) string {
	var result strings.Builder
	prevCipherBlock := NewBitArrayFromHex(hexIV)

	for i := 0; i < len(hexPlaintext); i += 16 {
		plainBlock := NewBitArrayFromHex(hexPlaintext[i : i+16])
		
		// 1. P_i XOR C_{i-1} (сцепление блоков)
		xored := plainBlock.XOR(prevCipherBlock)
		
		// 2. Шифрование блока
		cipherBlock := d.encryptBlock(xored, false)
		
		result.WriteString(cipherBlock.ToHex())
		prevCipherBlock = cipherBlock // запоминаем для следующего шага
	}
	return result.String()
}

// DecryptCBC дешифрует всю строку (в HEX) в режиме CBC
func (d *DES) DecryptCBC(hexCiphertext string, hexIV string) string {
	var result strings.Builder
	prevCipherBlock := NewBitArrayFromHex(hexIV)

	for i := 0; i < len(hexCiphertext); i += 16 {
		cipherBlock := NewBitArrayFromHex(hexCiphertext[i : i+16])
		
		// 1. Дешифрование блока
		decrypted := d.encryptBlock(cipherBlock, true)
		
		// 2. D(C_i) XOR C_{i-1} (восстановление открытого текста)
		plainBlock := decrypted.XOR(prevCipherBlock)
		
		result.WriteString(plainBlock.ToHex())
		prevCipherBlock = cipherBlock // запоминаем оригинальный зашифрованный блок
	}
	return result.String()
}

// ==========================================
// Вспомогательные функции
// ==========================================

func stringToHex(input string) string {
	const hexUpper = "0123456789ABCDEF"
	var sb strings.Builder
	sb.Grow(len(input) * 2)
	for i := 0; i < len(input); i++ {
		c := input[i]
		sb.WriteByte(hexUpper[c>>4])
		sb.WriteByte(hexUpper[c&0x0F])
	}
	return sb.String()
}

func hexToString(input string) string {
    bytes := make([]byte, len(input)/2)
    for i := 0; i < len(input); i += 2 {
        var b byte
        fmt.Sscanf(input[i:i+2], "%X", &b)
        bytes[i/2] = b
    }
    return string(bytes)
}

func padPKCS7(input string, blockSize int) string {
	paddingLen := blockSize - (len(input) % blockSize)
	padding := strings.Repeat(string(rune(paddingLen)), paddingLen)
	return input + padding
}

// ==========================================
// Демонстрация работы Лабы 5
// ==========================================

type DESDemo struct {
	des *DES
}

func NewDESDemo() *DESDemo {
	return &DESDemo{des: NewDES()}
}

func (demo *DESDemo) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=========================================================")
	fmt.Println("ЛАБОРАТОРНАЯ РАБОТА 5: Режимы работы блочных шифров (CBC)")
	fmt.Println("=========================================================\n")

	key := "133457799BBCDFF1"
	iv :=  "AABBCCDDEEFF0011" // Вектор инициализации (IV)

	fmt.Println("Ключ (HEX):", key)
	fmt.Println("Вектор инициализации IV (HEX):", iv)
	demo.des.SetKey(key)

	// Специально подбираем текст с повторяющимися паттернами для наглядности
	defaultText := "СЕКРЕТ!!СЕКРЕТ!!" 
	fmt.Printf("\nВведите открытый текст (или нажмите Enter для '%s'):\n> ", defaultText)

	rawPlaintext := ""
	if scanner.Scan() {
		rawPlaintext = scanner.Text()
	}
	if rawPlaintext == "" {
		rawPlaintext = defaultText
	}

	paddedPlaintext := padPKCS7(rawPlaintext, 8)
	fullPlaintextHex := stringToHex(paddedPlaintext)

	fmt.Println("\nОткрытый текст (HEX):", fullPlaintextHex)

	// ---------------------------------------------------------
	// ЧАСТЬ 1: Демонстрация уязвимости ECB
	// ---------------------------------------------------------
	fmt.Println("\n--- 1. Режим ECB (Electronic Codebook) ---")
	var ecbCiphertext strings.Builder
	for i := 0; i < len(fullPlaintextHex); i += 16 {
		block := fullPlaintextHex[i : i+16]
		encrypted := demo.des.EncryptBlock(block)
		ecbCiphertext.WriteString(encrypted)
		fmt.Printf("Блок P[%d] = %s -> C[%d] = %s\n", i/16+1, block, i/16+1, encrypted)
	}
	fmt.Println("Итоговый шифротекст (ECB):", ecbCiphertext.String())
	fmt.Println("! ОБРАТИТЕ ВНИМАНИЕ: Одинаковые блоки текста (если они есть) дают одинаковые блоки шифротекста.")

	// ---------------------------------------------------------
	// ЧАСТЬ 2: Демонстрация защиты CBC
	// ---------------------------------------------------------
	fmt.Println("\n--- 2. Режим CBC (Cipher Block Chaining) ---")
	cbcCiphertextHex := demo.des.EncryptCBC(fullPlaintextHex, iv)
	
	for i := 0; i < len(cbcCiphertextHex); i += 16 {
		fmt.Printf("Блок C[%d] = %s\n", i/16+1, cbcCiphertextHex[i:i+16])
	}
	fmt.Println("Итоговый шифротекст (CBC):", cbcCiphertextHex)
	fmt.Println("! ОБРАТИТЕ ВНИМАНИЕ: Паттерны скрыты. Одинаковые блоки открытого текста дают разный шифротекст.")

	cbcDecryptedHex := demo.des.DecryptCBC(cbcCiphertextHex, iv)
	fmt.Println("Расшифрованный текст (HEX):", cbcDecryptedHex)
	if fullPlaintextHex == cbcDecryptedHex {
		fmt.Println("Статус дешифрования CBC: УСПЕХ (Текст совпадает)")
	}

	// ---------------------------------------------------------
	// ЧАСТЬ 3: Распространение ошибки (Лавинный эффект в CBC)
	// ---------------------------------------------------------
	fmt.Println("\n--- 3. Распространение ошибки (искажение канала связи) в CBC ---")
	
	if len(cbcCiphertextHex) >= 32 { // нужно минимум 2 блока
		// Искажаем 10-й бит в ПЕРВОМ блоке шифротекста
		fmt.Println("Изменим 1 бит (10-й бит) в ПЕРВОМ блоке зашифрованного текста...")
		
		// Извлекаем первый блок и меняем бит
		firstBlockModified := demo.des.FlipBit(cbcCiphertextHex[:16], 10)
		modifiedCbcCiphertext := firstBlockModified + cbcCiphertextHex[16:]
		
		fmt.Println("Оригинальный шифротекст: ", cbcCiphertextHex)
		fmt.Println("Искаженный шифротекст:   ", modifiedCbcCiphertext)

		// Пытаемся расшифровать
		corruptedDecryptedHex := demo.des.DecryptCBC(modifiedCbcCiphertext, iv)
		fmt.Println("\nРезультат расшифровки (HEX):")
		fmt.Println("Ожидалось: ", fullPlaintextHex)
		fmt.Println("Получилось:", corruptedDecryptedHex)
		
		// Анализируем блоки
		origBlock1 := NewBitArrayFromHex(fullPlaintextHex[:16])
		corrBlock1 := NewBitArrayFromHex(corruptedDecryptedHex[:16])
		diff1 := 0
		for i := 0; i < origBlock1.Size(); i++ {
			if origBlock1.Get(i) != corrBlock1.Get(i) { diff1++ }
		}

		origBlock2 := NewBitArrayFromHex(fullPlaintextHex[16:32])
		corrBlock2 := NewBitArrayFromHex(corruptedDecryptedHex[16:32])
		diff2 := 0
		for i := 0; i < origBlock2.Size(); i++ {
			if origBlock2.Get(i) != corrBlock2.Get(i) { diff2++ }
		}

		fmt.Printf("\nАнализ ошибок:\n")
		fmt.Printf("- Блок 1 поврежден сильно (лавинный эффект): изменилось %d бит из 64\n", diff1)
		fmt.Printf("- Блок 2 изменен точечно (XOR с предыдущим): изменилось %d бит из 64\n", diff2)
		fmt.Println("- Остальные блоки (если есть) расшифруются без ошибок.")
	} else {
		fmt.Println("Для демонстрации лавинного эффекта в CBC текст должен быть длиннее 8 символов.")
	}
}

func main() {
	demo := NewDESDemo()
	demo.Run()
}