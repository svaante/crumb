package main

import (
    "os"
    "fmt"
    "path/filepath"
    "encoding/json"
    "strings"
    "io/ioutil"
    "regexp"
    "flag"
    "log"
)

type PrePostFix struct {
    postfix string `json:postfix`
    prefix string `json:prefix`
}

type Config struct {
    stopAt string `json:stopAt`
    todoFileName string `json:todoFileName`
    visualDoneMarker string `json:visualDoneMarker`
    visualNotDoneMarker string `json:visualNotDoneMarker`
    visualMarkerPrefix string `json:visualMarkerPrefix`
    visualMarkerPostfix string `json:visualMarkerPostfix`
    doneMarker string `json:doneMarker`
    notDoneMarker string `json:notDoneMarker`
}

type Todo struct {
    done bool
    text string
}

func newDefaultConfig() *Config {
    return &Config{
        stopAt: "/",
        todoFileName: ".todo",
        visualDoneMarker: "[]",
        visualNotDoneMarker: "[x]",
        doneMarker: "x ",
        notDoneMarker: "",
    }
}

func getHomePath() string {
    home := os.Getenv("HOME")
    if home == "" {
        log.Fatal("User envar is empty")
    }
    return home
}

func applyUserConfig(conf *Config) {
    configFileName := ".todorc.json"
    todorcFilePath := filepath.Join(getHomePath(), configFileName)

    content, err := ioutil.ReadFile(todorcFilePath)
    if err != nil {
        log.Fatal(fmt.Sprintf("Unable to open %s", configFileName))
    }

    if json.NewDecoder(strings.NewReader(string(content))).Decode(conf) != nil {
        log.Fatal(fmt.Sprintf("Invalid %s", configFileName))
    }
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

func formatTodo(todoLine string, conf *Config) Todo {
    rDone, rDoneErr := regexp.Compile(fmt.Sprintf("^%s", conf.doneMarker))
    if rDoneErr != nil {
        log.Fatal(fmt.Sprintf("Bad `doneMarker=%s` unable to compile regexp",
                                   conf.doneMarker))
    }

    rText, rTextErr := regexp.Compile(fmt.Sprintf("(^%s|)(.*)", conf.doneMarker))
    if rTextErr != nil {
        log.Fatal(fmt.Sprintf("Bad `doneMarker=%s` unable to compile regexp",
                                   conf.doneMarker))
    }

    todo := Todo{}

    todo.done = rDone.MatchString(todoLine)
    todo.text = rText.FindStringSubmatch(todoLine)[2]

    return todo
}

func formatIntoTodos(todoContent string, conf *Config) []Todo {
    todoLines := strings.Split(todoContent, "\n")


    var todos []Todo
    for _, todoLine := range(todoLines) {
        if (todoLine != "") {
            todos = append(todos, formatTodo(todoLine, conf))
        }
    }

    return todos
}

func printTodo(todo Todo, conf *Config) {
    var marker string
    if todo.done {
        marker = conf.visualDoneMarker
    } else {
        marker = conf.visualNotDoneMarker
    }

    fmt.Printf("%s %s\n", marker, todo.text)
}

func getWD() string {
    dir, err := os.Getwd()
    if err != nil {
        log.Fatal("Failed getting working directory")
    }
    return dir
}

func findTodoFiles(dir string, conf *Config) []string {
    var todoFilePaths []string
    for basePath := dir; basePath != conf.stopAt; basePath = filepath.Join(basePath, "..") {
        todoFilePath := filepath.Join(basePath, conf.todoFileName)
        todoFilePaths = append(todoFilePaths, todoFilePath)
    }

    return todoFilePaths
}

func printTodoFile(todoFilePath string, conf *Config) {
    if fileExists(todoFilePath) {
        fmt.Println(filepath.Join(todoFilePath, ".."))
        fileContent := readFile(todoFilePath)
        todos := formatIntoTodos(fileContent, conf)
        for _, todo := range todos {
            printTodo(todo, conf)
        }
    }
}

func ls(dir string, bubble bool, conf *Config) {
    if bubble {
        todoFilePaths := findTodoFiles(dir, conf)

        for _, todoFilePath := range todoFilePaths {
            printTodoFile(todoFilePath, conf)
        }
    } else {
        todoFilePath := filepath.Join(dir, conf.todoFileName)

        printTodoFile(todoFilePath, conf)
    }
}

func add(dir string, todoLine string, conf *Config) {
    todoFilePath := filepath.Join(dir, conf.todoFileName)

    file, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(fmt.Sprintf("Unable to access %s for writing", todoFilePath))
    }

    if _, err := file.Write([]byte(todoLine + "\n")); err != nil {
        log.Fatal(err)
    }

    if err := file.Close(); err != nil {
        log.Fatal(err)
    }
}

func main() {
    conf := newDefaultConfig()
    applyUserConfig(conf)

    dir := getWD()
    ls(dir, true, conf)
}
