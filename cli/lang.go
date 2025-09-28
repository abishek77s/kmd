// package main

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// // Supported languages and file extensions
// var languageExtensions = map[string][]string{
// 	"c":   {".c"},
// 	"cpp": {".cpp", ".cxx", ".cc"},
// 	"go":  {".go"},
// 	"py":  {".py"},
// }

// // Template generator for Makefile per language
// func generateMakefile(lang string, srcFiles []string, target string) string {
// 	switch lang {
// 	case "c":
// 		objs := []string{}
// 		for _, f := range srcFiles {
// 			objs = append(objs, strings.Replace(f, ".c", ".o", 1))
// 		}
// 		return fmt.Sprintf(`# Generated Makefile for C project
// CC = gcc
// CFLAGS = -Wall -g
// TARGET = %s
// SRCS = %s
// OBJS = %s

// $(TARGET): $(OBJS)
// 	$(CC) $(CFLAGS) -o $@ $(OBJS)

// %%.o: %%.c
// 	$(CC) $(CFLAGS) -c $< -o $@

// clean:
// 	rm -f $(OBJS) $(TARGET)
// `, target, strings.Join(srcFiles, " "), strings.Join(objs, " "))
// 	case "cpp":
// 		objs := []string{}
// 		for _, f := range srcFiles {
// 			objs = append(objs, strings.Replace(f, ".cpp", ".o", 1))
// 		}
// 		return fmt.Sprintf(`# Generated Makefile for C++ project
// CXX = g++
// CXXFLAGS = -Wall -g
// TARGET = %s
// SRCS = %s
// OBJS = %s

// $(TARGET): $(OBJS)
// 	$(CXX) $(CXXFLAGS) -o $@ $(OBJS)

// %%.o: %%.cpp
// 	$(CXX) $(CXXFLAGS) -c $< -o $@

// clean:
// 	rm -f $(OBJS) $(TARGET)
// `, target, strings.Join(srcFiles, " "), strings.Join(objs, " "))
// 	case "go":
// 		return fmt.Sprintf(`# Generated Makefile for Go project
// TARGET = %s

// build:
// 	go build -o $(TARGET) .

// run: build
// 	./$(TARGET)

// clean:
// 	rm -f $(TARGET)
// `, target)
// 	case "py":
// 		return fmt.Sprintf(`# Generated Makefile for Python project
// run:
// 	python3 main.py

// clean:
// 	rm -rf __pycache__
// `)
// 	default:
// 		return "# Unsupported language"
// 	}
// }

// // Detect language based on files
// func detectLanguage(srcFiles []string) string {
// 	for lang, exts := range languageExtensions {
// 		for _, f := range srcFiles {
// 			for _, ext := range exts {
// 				if strings.HasSuffix(f, ext) {
// 					return lang
// 				}
// 			}
// 		}
// 	}
// 	return ""
// }

// // Scan project directory recursively for source files
// func scanProject(dir string) ([]string, string) {
// 	var srcFiles []string
// 	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil || info.IsDir() {
// 			return err
// 		}
// 		ext := filepath.Ext(path)
// 		for _, exts := range languageExtensions {
// 			for _, e := range exts {
// 				if ext == e {
// 					srcFiles = append(srcFiles, path)
// 				}
// 			}
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		fmt.Printf("Error scanning directory: %v\n", err)
// 		os.Exit(1)
// 	}
// 	lang := detectLanguage(srcFiles)
// 	return srcFiles, lang
// }

// func main() {
// 	dir := "."
// 	if len(os.Args) > 1 {
// 		dir = os.Args[1]
// 	}

// 	srcFiles, lang := scanProject(dir)
// 	if lang == "" || len(srcFiles) == 0 {
// 		fmt.Println("No supported source files found in the project.")
// 		return
// 	}

// 	target := "app"
// 	makefileContent := generateMakefile(lang, srcFiles, target)

// 	err := os.WriteFile("Makefile", []byte(makefileContent), 0644)
// 	if err != nil {
// 		fmt.Printf("Error writing Makefile: %v\n", err)
// 		return
// 	}

// 	fmt.Printf("✅ Makefile generated for %s project with %d source files.\n", lang, len(srcFiles))
// }
