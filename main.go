package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ==========================================
// Структура GammaCipher и методы (Лаба 6)
// ==========================================

type GammaCipher struct {
	a        int // Множитель (Multiplier)
	c        int // Приращение (Increment)
	b        int // Модуль (Modulus), равный количеству значений в 1 байте
	t0       int // Начальное значение (seed)
	tCurrent int // Текущее состояние генератора
}

// NewGammaCipher создает новый экземпляр шифратора
func NewGammaCipher(t0 int) *GammaCipher {
	return &GammaCipher{
		a:        21,
		c:        33,
		b:        256,
		t0:       t0,
		tCurrent: t0,
	}
}

// Reset сбрасывает состояние генератора к исходному значению t0.
// Это необходимо для синхронизации генератора перед началом дешифрования.
func (gc *GammaCipher) Reset() {
	gc.tCurrent = gc.t0
}

// GenerateGamma генерирует псевдослучайную последовательность (гамму) заданной длины
func (gc *GammaCipher) GenerateGamma(length int) []int {
	gamma := make([]int, length)
	for i := 0; i < length; i++ {
		gamma[i] = gc.tCurrent
		// Вычисляем следующее состояние по формуле ЛКГ: T_{i+1} = (A * T_i + C) mod B
		gc.tCurrent = (gc.a*gc.tCurrent + gc.c) % gc.b
	}
	return gamma
}

// Encrypt шифрует открытый текст методом наложения гаммы и возвращает шифртекст и саму гамму
func (gc *GammaCipher) Encrypt(text string) (string, []int) {
	// Генерируем гамму длиной, равной количеству байт открытого текста
	gamma := gc.GenerateGamma(len(text))
	var cipherBuilder strings.Builder
	cipherBuilder.Grow(len(text))

	for i := 0; i < len(text); i++ {
		x := int(text[i]) // Берем код байта открытого текста
		// Накладываем гамму: C_i = (P_i + Gamma_i) mod 256
		y := (x + gamma[i]) % 256
		cipherBuilder.WriteByte(byte(y))
	}

	return cipherBuilder.String(), gamma
}

// Decrypt дешифрует шифртекст с использованием сгенерированной гаммы
func (gc *GammaCipher) Decrypt(cipher string, gamma []int) string {
	var textBuilder strings.Builder
	textBuilder.Grow(len(cipher))

	for i := 0; i < len(cipher); i++ {
		y := int(cipher[i]) // Берем код зашифрованного байта
		// Снимаем гамму: P_i = (C_i - Gamma_i + 256) mod 256
		x := (y - gamma[i] + 256) % 256
		textBuilder.WriteByte(byte(x))
	}

	return textBuilder.String()
}

// ==========================================
// Демонстрация работы Лабы 6
// ==========================================

func main() {
	fmt.Println("=========================================================")
	fmt.Println("ЛАБОРАТОРНАЯ РАБОТА 6: Шифрование методом гаммирования")
	fmt.Println("=========================================================\n")

	t0 := 20 // Инициализационный вектор (начальное состояние генератора)
	cipher := NewGammaCipher(t0)

	fmt.Print("Введите открытый текст:\n> ")
	reader := bufio.NewReader(os.Stdin)
	plaintext, _ := reader.ReadString('\n')
	
	// Очищаем от символов переноса строки (поддерживает и Windows, и Linux)
	plaintext = strings.TrimRight(plaintext, "\r\n")

	// Шифруем текст
	ciphertext, gamma := cipher.Encrypt(plaintext)

	// Перед дешифрованием сбрасываем генератор на начальное состояние t0
	cipher.Reset()
	decrypted := cipher.Decrypt(ciphertext, gamma)

	// --- Блок вывода результатов пользователю ---
	fmt.Println("\n--- Результаты ---")
	fmt.Printf("Открытый текст: %s\n\n", plaintext)

	fmt.Println("Гамма (значения состояний):")
	for i, g := range gamma {
		fmt.Printf("%3d ", g)
		if (i+1)%20 == 0 { // Перенос строки каждые 20 элементов
			fmt.Println()
		}
	}
	fmt.Println()

	fmt.Printf("\nШифртекст (HEX)      : %x\n", ciphertext) // Выводим в HEX, т.к. символы могут быть нечитаемыми
	fmt.Printf("Шифртекст (Символы)  : %s\n", ciphertext)
	fmt.Printf("Дешифрованный текст  : %s\n", decrypted)

	// Валидация алгоритма
	if plaintext == decrypted {
		fmt.Println("\nСтатус: Дешифрование прошло успешно!")
	} else {
		fmt.Println("\nСтатус: Ошибка дешифрования!")
	}
}