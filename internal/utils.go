package crumb

import (
    "log"
    "os"
    "fmt"
    "io/ioutil"
    "errors"
    "path/filepath"
)

func getWD() string {
    dir, err := os.Getwd()
    if err != nil {
        log.Fatal("Failed getting working directory")
    }
    return dir
}

func getHomePath() string {
    home := os.Getenv("HOME")
    if home == "" {
        log.Fatal("User envar is empty")
    }
    return home
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func readFile(path string) string {
    content, err := ioutil.ReadFile(path)
    if err != nil {
        log.Fatal(fmt.Sprintf("Unable to open file %s",
                                   path))
    }

    return string(content)
}

func getValidDir(dirPath string) (string, error) {
    absDirPath, absDirPathErr := filepath.Abs(dirPath)
    if absDirPathErr != nil {
        return "", errors.New("Could not create a valid filepath")
    }

    info, infoErr := os.Stat(absDirPath)
    if os.IsNotExist(infoErr) {
        return "", errors.New("Dir does not exist")
    }

    if info.IsDir() {
        return absDirPath, nil
    }

    return "", errors.New("Path does not lead to a dir")
}

func writeFile(path string, content string) {
    if err := os.Truncate(path, 0); err != nil {
        log.Fatal(fmt.Sprintf("Unable to truncate %s", path))
    }

    file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(fmt.Sprintf("Unable to access %s for writing", path))
    }

    if _, err := file.Write([]byte(content)); err != nil {
        log.Fatal(err)
    }

    if err := file.Close(); err != nil {
        log.Fatal(err)
    }
}

