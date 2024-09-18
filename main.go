package main

import (
	"LoGo/logprocessor"
	"LoGo/utils"
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Инициализация приложения
	a := app.New()
	w := a.NewWindow("Log Analyzer")

	w.Resize(fyne.NewSize(720, 360))

	// Поля для ввода и выбора
	logDirEntry := widget.NewEntry()
	logDirEntry.SetText("test_logs")

	outputFileEntry := widget.NewEntry()
	outputFileEntry.SetText("results/errors.log")

	patterns := []string{"ERROR", "WARN", "error"}
	patternsCheck := make([]*widget.Check, len(patterns))
	for i, pattern := range patterns {
		patternsCheck[i] = widget.NewCheck(pattern, nil)
	}

	numWorkersLabel := widget.NewLabel("Количество потоков: ")
	numWorkersEntry := widget.NewEntry()
	numWorkersEntry.SetText("1")

	// Поле для вывода времени выполнения
	executionTimeLabel := widget.NewLabel("Время выполнения: -")

	// Поле для вывода содержимого итогового файла
	resultFileContent := widget.NewMultiLineEntry()
	resultFileContent.SetPlaceHolder("Результаты анализа будут показаны здесь...")
	resultFileContent.Disable()

	// Кнопки выбора директорий
	selectLogDirButton := widget.NewButton("Выбрать директорию с логами", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err == nil && dir != nil {
				logDirEntry.SetText(dir.Path())
			}
		}, w)
	})

	selectOutputFileButton := widget.NewButton("Выбрать файл для записи", func() {
		dialog.ShowFileSave(func(file fyne.URIWriteCloser, err error) {
			if err == nil && file != nil {
				outputFileEntry.SetText(file.URI().Path())
				file.Close()
			}
		}, w)
	})

	// Кнопка запуска
	startButton := widget.NewButton("Запустить анализ", func() {
		logDir := logDirEntry.Text
		outputFile := outputFileEntry.Text
		bufferSize := 4
		runtime.GOMAXPROCS(utils.ParseNumWorkers(numWorkersEntry.Text))
		// numWorkers := utils.ParseNumWorkers(numWorkersEntry.Text)

		// Сбор выбранных паттернов
		var selectedPatterns []string
		for i, check := range patternsCheck {
			if check.Checked {
				selectedPatterns = append(selectedPatterns, patterns[i])
			}
		}

		start := time.Now()

		// Запуск анализа
		if err := logprocessor.AnalyzeLogs(logDir, outputFile, selectedPatterns, bufferSize); err != nil {
			log.Printf("Ошибка анализа: %v\n", err)
		} else {
			executionTimeLabel.SetText(fmt.Sprintf("Время выполнения: %s", time.Since(start)))
			fmt.Println("Анализ завершён")

			// Чтение и вывод только первых 20 строк итогового файла
			file, err := os.Open(outputFile)
			if err != nil {
				resultFileContent.SetText(fmt.Sprintf("Ошибка открытия файла: %v", err))
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			var output string
			lineCount := 0
			maxLines := 50

			for scanner.Scan() {
				output += scanner.Text() + "\n"
				lineCount++
				if lineCount >= maxLines {
					break
				}
			}

			if err := scanner.Err(); err != nil {
				resultFileContent.SetText(fmt.Sprintf("Ошибка чтения файла: %v", err))
			} else {
				resultFileContent.SetText(output)
			}
		}
	})

	// Преобразование patternsCheck в []fyne.CanvasObject
	checkBoxes := make([]fyne.CanvasObject, len(patternsCheck))
	for i, check := range patternsCheck {
		checkBoxes[i] = check
	}

	// Упаковка интерфейса
	leftColumn := container.NewVBox(
		logDirEntry, selectLogDirButton,
		outputFileEntry, selectOutputFileButton,
		widget.NewLabel("Паттерны для поиска"),
		container.NewVBox(checkBoxes...),
		numWorkersLabel,
		numWorkersEntry,
		startButton,
		executionTimeLabel,
	)
	rightColumn := container.NewVScroll(resultFileContent)

	mainContent := container.NewGridWithColumns(2,
		leftColumn,
		rightColumn,
	)

	w.SetContent(
		container.NewBorder(nil, nil, nil, nil, mainContent),
	)

	w.ShowAndRun()
}
