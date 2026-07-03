package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SearchResult struct {
	Line int    //номер строки
	Text string // текст строки
}

func main() {
	var pattern string
	var minSize int64
	var ext string
	var content string
	var showHelp bool
	var output string

	flag.StringVar(&pattern, "p", "", "поиск имени")
	flag.Int64Var(&minSize, "s", 0, "минимальны размер")
	flag.StringVar(&ext, "ext", "", "расширение файла")
	flag.StringVar(&content, "content", "", "поиск внутри файла")
	flag.BoolVar(&showHelp, "h", false, "показать справку")
	flag.BoolVar(&showHelp, "help", false, "показать справку")
	flag.StringVar(&output, "o", "", "сохранить результат в файл")

	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	args := flag.Args()
	root := "."
	if len(args) > 0 {
		root = args[0]
	}
	files, err := findFiles(root, pattern, minSize, ext)
	if err != nil {
		os.Exit(1)
	}

	var outputLines []string

	if content != "" {
		var filtered []string
		for _, file := range files {
			matchers := searchInFile(file, content)
			if len(matchers) > 0 {
				filtered = append(filtered, file)
				for _, match := range matchers {
					line := fmt.Sprintf("%s: строка %d: %s", file, match.Line, match.Text)
					fmt.Println(line)
					outputLines = append(outputLines, line)
				}
			}
		}
		files = filtered
	}

	if len(files) == 0 {
		fmt.Println("Файлы не найдены")
		return
	}

	if output != "" {
		content := strings.Join(outputLines, "\n")
		err := os.WriteFile(output, []byte(content), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Результат сохранен в %s\n", output)
	}
}

func findFiles(root string, pattern string, minSize int64, ext string) ([]string, error) {
	var results []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		findExt := filepath.Ext(path)
		if info.IsDir() {
			return nil
		}
		if ext != "" && ext != findExt {
			return nil
		}
		if pattern != "" {
			if !strings.Contains(info.Name(), pattern) {
				return nil
			}
		}
		if minSize > info.Size() {
			return nil
		}
		results = append(results, path)

		return nil
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}

func searchInFile(path string, findStr string) []SearchResult {
	result := []SearchResult{}
	isFile, err := os.Stat(path)

	if err != nil {
		return result
	}

	if isFile.IsDir() {
		return result
	}

	file, err := os.Open(path)
	if err != nil {
		return result
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, findStr) {
			result = append(result, SearchResult{
				Line: lineNum,
				Text: line,
			})
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return result
	}
	return result
}

func printHelp() {
	fmt.Println(`find-file — утилита для поиска файлов с фильтрацией

Использование:
  find-file [флаги] <путь>

Флаги:
  -p <строка>      искать в имени файла
  -s <байты>       минимальный размер файла
  -ext <расширение> фильтр по расширению (например .go)
  -content <строка> искать внутри файла
  -o <файл>        сохранить результат в файл
  -h, -help        показать эту справку

Примеры:
  find-file . -p="main"
  find-file /home/user -ext=".go" -s=1024
  find-file . -content="timeout"
  find-file . -ext=".txt" -content="Hello" -o=result.txt`)
}
