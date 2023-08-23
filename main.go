package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/carlzhc-go/sh"
	"github.com/gookit/ini/v2"
	"github.com/icza/gog"
)

// Global variables
var dryRun bool
var configFile string

// First run
func init() {
	log.SetFlags(0)
	flag.BoolVar(&dryRun, "n", false, "Run in dry-run mode")
	flag.Parse()

	configFile = flag.Arg(0)
	if configFile == "" {
		log.Fatalln("Usage: rename-drama [-n] CONFIG-FILE")
	}

	if err := ini.LoadFiles(configFile); err != nil {
		log.Fatalln(err)
	}
}

func renameDrama(pattern, extension, dramaName string) {
	var serial int
	var err error
	re := regexp.MustCompile(pattern)
	fileNames := gog.Must(filepath.Glob("*" + extension))
	for _, fileName := range fileNames {
		if !sh.Test("-f", fileName) {
			continue
		}

		// Try to determine the serial number from the file name.
		log.Println("Checking file " + fileName)
		matches := re.FindStringSubmatch(fileName)

		if matches == nil || len(matches) < 2 {
			continue
		}

		serial, err = strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		var ext string
		if extension == "" {
			ext = filepath.Ext(fileName)
		} else {
			ext = extension
		}

		newName := fmt.Sprintf("%s-%03d%s", dramaName, serial, ext)
		if dryRun {
			log.Printf("[DRY-RUN] Rename '%s' -> '%s'\n", fileName, newName)
		} else if err := os.Rename(fileName, newName); err != nil {
			log.Fatalf("Cannot rename file '%s' -> '%s'\n", fileName, newName)
		} else {
			log.Printf("Rename '%s' -> '%s'\n", fileName, newName)
		}
	}
}

func main() {
	// Entering destination directory
	destDir := filepath.Dir(configFile)
	if err := os.Chdir(destDir); err != nil {
		log.Fatal(err)
	}

	log.Printf("Entering %s\n", destDir)

	pattern := ini.String("pattern")
	extension := ini.String("extension")
	dramaName := ini.String("name")

	if pattern == "" {
		log.Fatal("Parameter 'pattern' not found in config file " + configFile)
	}

	if dramaName == "" {
		cwd := gog.Must(os.Getwd())
		dramaName = filepath.Base(cwd)
	}

	if dramaName == "" || dramaName == "/" || dramaName == "." || dramaName == "\\" {
		log.Fatalln("No name provided or cannot deduce name from directory name")
	}

	log.Println("File name pattern: ", pattern)
	renameDrama(pattern, extension, dramaName)
}
