package logprocessor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// containsError проверяет строку на наличие паттернов
func containsError(line string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	return false
}

// processFile обрабатывает один файл логов
func processFile(filePath string, patterns []string, wg *sync.WaitGroup, output chan<- string, bufferSize int) {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Ошибка открытия файла %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	lines := make(chan string, bufferSize)

	var lineWg sync.WaitGroup
	for i := 0; i < bufferSize; i++ {
		lineWg.Add(1)
		go func() {
			defer lineWg.Done()
			for line := range lines {
				if containsError(line, patterns) {
					output <- fmt.Sprintf("[%s]: %s\n", filepath.Base(filePath), line)
				}
			}
		}()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
	close(lines)

	lineWg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Printf("Ошибка чтения файла %s: %v\n", filePath, err)
	}
}

// AnalyzeLogs проводит анализ логов по указанной директории
func AnalyzeLogs(logDir string, outputFile string, patterns []string, bufferSize int) error {
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("ошибка создания файла %s: %v", outputFile, err)
	}
	defer output.Close()

	errorLines := make(chan string)

	go func() {
		for line := range errorLines {
			output.WriteString(line)
		}
	}()

	start := time.Now()

	var wg sync.WaitGroup
	err = filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".log") {
			wg.Add(1)
			go processFile(path, patterns, &wg, errorLines, bufferSize)
		}
		return nil
	})

	if err != nil {
		return err
	}

	wg.Wait()
	close(errorLines)

	duration := time.Since(start)
	fmt.Printf("Время выполнения: %s\n", duration)

	return nil
}
