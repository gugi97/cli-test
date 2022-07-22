package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type inputFile struct {
	filePath   string
	outputPath *string
	convType   *string
}

func main() {
	fileData, err := getFileData()
	if err != nil {
		exitGracefully(err)
	}

	if err = checkIfValidFile(fileData.filePath); err != nil {
		exitGracefully(err)
	}

	writeChannel := make(chan string)

	if *fileData.convType == "json" {
		done := make(chan bool)
		go processLogFile(fileData.filePath, writeChannel)
		go writeJsonFile(fileData, writeChannel, done)
		<-done
	} else if *fileData.convType == "text" || *fileData.convType == "plaintext" {
		done := make(chan bool)
		go processLogFile(fileData.filePath, writeChannel)
		go writeTextFile(fileData, writeChannel, done)
		<-done
	} else {
		err = errors.New("conversion type not valid")
		exitGracefully(err)
	}
}

// =====================================================================================
func getFileData() (inputFile, error) {
	// to validate the correct number of arguments
	if len(os.Args) < 2 {
		return inputFile{}, errors.New("A filepath argument is required")
	}

	var args []string
	var flags []string
	var isFlag = false

	// to make argument can assign before flag (cause default is can't)
	for i := 0; i < len(os.Args); i++ {
		if string(os.Args[i][0]) == "-" {
			isFlag = true
		}
		if i == 0 || isFlag {

			if isFlag {
				flags = append(flags, os.Args[i])
				i++
				if !(i >= len(os.Args)) {
					flags = append(flags, os.Args[i])
				}
			} else {
				flags = append(flags, os.Args[i])
			}
		} else {
			args = append(args, os.Args[i])
		}
		isFlag = false
	}
	os.Args = flags

	convType := flag.String("t", "plaintext", "Conversion Type")
	outputPath := flag.String("o", "basefile", "Output Path File")

	flag.Parse() // parse all the arguments from the terminal

	filePath := args[0] // is the file location (log file)

	return inputFile{filePath, outputPath, convType}, nil
}

func processLogFile(filePath string, writeChannel chan<- string) {
	file, err := os.Open(filePath)
	checkError(err)
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	// read file per line
	for fileScanner.Scan() {
		writeChannel <- fileScanner.Text()
		fmt.Println(fileScanner.Text())
	}

	// when file reach EOF close channel
	if len(fileScanner.Text()) == 0 {
		close(writeChannel)
	}
}

func createStringWriter(fileData inputFile) func(string, bool) {
	var jsonDir, jsonName string

	if *fileData.outputPath != "basefile" {
		jsonDir = filepath.Dir(*fileData.outputPath)
		err := ensureDir(jsonDir)
		checkError(err)
		jsonName = fmt.Sprintf("%s", filepath.Base(*fileData.outputPath))
	} else {
		var extension string
		if *fileData.convType == "text" || *fileData.convType == "plaintext" {
			extension = "txt"
		} else {
			extension = "json"
		}
		jsonDir = filepath.Dir(fileData.filePath)
		jsonName = fmt.Sprintf("%s.%s", strings.TrimSuffix(filepath.Base(fileData.filePath), ".log"), extension)
	}

	finalLocation := fmt.Sprintf("%s/%s", jsonDir, jsonName)
	f, err := os.Create(finalLocation)
	checkError(err)

	return func(data string, close bool) {
		_, err = f.WriteString(data)
		checkError(err)

		if close {
			f.Close()
		}
	}
}

// ==================================== Json File ======================================
func writeJsonFile(fileData inputFile, writerChannel <-chan string, done chan<- bool) {
	writeString := createStringWriter(fileData)
	jsonFunc, breakLine := getJSONFunc()

	fmt.Println("Writing JSON file...")

	writeString("["+breakLine, false)
	first := true

	for {
		record, more := <-writerChannel
		// if more true it means channel stil open
		if more {
			if !first {
				writeString(","+breakLine, false)
			} else {
				first = false
			}

			jsonData := jsonFunc(record)
			writeString(jsonData, false)
		} else {
			writeString(breakLine+"]", true)
			fmt.Println("Completed!")
			done <- true
			break
		}

	}
}

func getJSONFunc() (jsonFunc func(string) string, breakLine string) {
	breakLine = "\n"
	jsonFunc = func(record string) string {
		jsonData, _ := json.MarshalIndent(record, "   ", "   ")
		return "   " + string(jsonData)
	}

	return jsonFunc, breakLine
}

// ================================= Plain Text File ====================================
func writeTextFile(fileData inputFile, writerChannel <-chan string, done chan<- bool) {
	writeString := createStringWriter(fileData)

	fmt.Println("Writing Plain Text file...")

	for {
		record, more := <-writerChannel
		// if more true it means channel stil open
		if more {
			writeString(record, false)
			writeString("\n", false)
		} else {
			writeString("\n", true)
			fmt.Println("Completed!")
			done <- true
			break
		}

	}
}

// ==================================== Validation ======================================
func checkIfValidFile(file string) error {
	// Check if file is Log
	if fileExtension := filepath.Ext(file); fileExtension != ".log" {
		return fmt.Errorf("file %s is not Log", file)
	}

	// Check if file does exist
	if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", file)
	}

	return nil
}

func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0770) // directory linux permissions
	checkError(err)
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dirName)
		checkError(err)
		if !info.IsDir() {
			return errors.New("path exists but is not a directory")
		}
		return nil
	}
	return err
}

// ==================================== Handle Error ====================================
func checkError(e error) {
	if e != nil {
		exitGracefully(e)
	}
}

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
