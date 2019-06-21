package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// If no error is returned (exit status 0), the task is considered up-to-date:
func main() {
	var command = flag.String("c", "", "(required) golang fully qualified package for target command")
	var generatesBinary = flag.String("g", "", "(required) path to the target compiled command binary (does not need to exist)")
	var projectPrefix = flag.String("p", "", "(optional) for multi command go projects, this is the go project's base path. This helps exclude go standard library source files and vendor source. Defaults to the `-c` value")
	var help = flag.Bool("help", false, "print this help")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	validateFlags(*command, *generatesBinary)
	if *projectPrefix == "" {
		projectPrefix = command
	}

	msg, exitCode := run(*command, *generatesBinary, *projectPrefix)
	fmt.Println(msg)
	os.Exit(exitCode)
}

func validateFlags(command, generatesBinary string) {
	errors := make([]string, 0)
	if command == "" {
		errors = append(errors, fmt.Sprintf("-c flag is not set"))
	}
	if generatesBinary == "" {
		errors = append(errors, fmt.Sprintf("-g flag is not set"))
	}
	if len(errors) > 0 {
		fmt.Printf("%v\n", strings.Join(errors[:], ", "))
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func run(command, generatesBinary, projectPrefix string) (string, int) {
	if !generatedBinaryExists(generatesBinary) {
		return fmt.Sprintf("%s is out of date because it does not exist", generatesBinary), 1
	}

	sources := listAllSources(command, projectPrefix)
	lastModTime, lastModFile := maxModifiedDate(sources)
	if binaryUpToDate(lastModTime, generatesBinary) {
		return fmt.Sprintf("%s is up to date", generatesBinary), 0
	}
	return fmt.Sprintf("%s is out of date with %s", generatesBinary, lastModFile), 1
}

func generatedBinaryExists(generatesBinary string) bool {
	_, err := os.Stat(generatesBinary)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			unhandledError(fmt.Sprintf("cannot get file info for %s", generatesBinary), err)
		}
	}
	return true
}

func unhandledError(msg string, err error) {
	fmt.Printf("%s\n%v\n", msg, err)
	os.Exit(255)
}

func binaryUpToDate(lastModTime time.Time, generatesBinary string) bool {
	fileInfo, err := os.Stat(generatesBinary)
	if err != nil {
		if os.IsNotExist(err) {
			return false // binary doesn't exist so it needs to be recompiled
		} else {
			unhandledError(fmt.Sprintf("cannot get file info for %s", generatesBinary), err)
		}
	}
	return fileInfo.ModTime().After(lastModTime)
}

// maxModifiedDate returns the max modified time from the list of files as well as the corresponding filename
func maxModifiedDate(fileList []string) (time.Time, string) {
	modTimes := make([]time.Time, 0)
	for _, file := range fileList {
		fileInfo, err := os.Stat(file)
		if err != nil {
			unhandledError(fmt.Sprintf("cannot get file info for %s", file), err)
		}
		modTimes = append(modTimes, fileInfo.ModTime())
	}
	maxTime := time.Unix(0, 0)
	var maxModFile string
	for i, modTime := range modTimes {
		if modTime.After(maxTime) {
			maxTime = modTime
			maxModFile = fileList[i]
		}
	}
	return maxTime, maxModFile
}

// listAllSources returns the absolute path of direct and imported source files for the project. Import source files are
// queried in parallel to improve performance.
func listAllSources(goCommand, projectPrefix string) []string {
	sources := listPackageSources(goCommand, projectPrefix)
	packageImports := listImports(goCommand)
	mux := &sync.Mutex{}
	waitgroup := &sync.WaitGroup{}
	for _, importedPackage := range packageImports {
		waitgroup.Add(1)
		go func() {
			listedSources := listPackageSources(importedPackage, projectPrefix)
			mux.Lock()
			sources = append(sources, listedSources...)
			mux.Unlock()
			waitgroup.Done()
		}()
	}
	waitgroup.Wait()
	return sources
}

func listImports(goPackage string) []string {
	goListCmd := exec.Command("go", "list", "-f", `'{{join .Imports "\n"}}'`, goPackage)

	outputBytes, err := goListCmd.CombinedOutput()
	if err != nil {
		unhandledError(string(outputBytes), err)
	}

	return strings.Split(strings.TrimSpace(string(outputBytes)), "\n")
}

func listPackageSources(goPackage, projectPrefix string) (sources []string) {
	gopath := os.Getenv("GOPATH")

	// exclude vendor and standard library packages.
	// TODO include vendor source
	if !strings.Contains(goPackage, projectPrefix) {
		return
	}

	goListSourceCmd := exec.Command("go", "list", "-f", `'{{join .GoFiles "\n"}}'`, goPackage)
	sourcesBytes, err := goListSourceCmd.CombinedOutput()
	if err != nil {
		unhandledError(string(sourcesBytes), err)
	}
	cleanSources := strings.Trim(strings.TrimSpace(string(sourcesBytes)), "'")
	for _, sourceFileName := range strings.Split(cleanSources, "\n") {
		fullyQualified := fmt.Sprintf("%s/src/%s/%s", gopath, goPackage, sourceFileName)
		sources = append(sources, fullyQualified)
	}
	return
}
