package crumb

import (
    "path/filepath"
    "strings"
    "io/ioutil"
    "os"
    "log"
    "fmt"
    "bufio"
    "strconv"
)

func parseSelection(input string, lineNumbers []int) []int {
    var selections []int
    addToSelection := func (i int) bool {
        if 0 < i && i <= len(lineNumbers) {
            selections = append(selections, lineNumbers[i - 1])
            return true
        }
        return false
    }
    parseAndAddToSelection := func (str string) bool {
        if i, err := strconv.Atoi(str); err == nil {
            return addToSelection(i)
        }
        return false
    }

    if parseAndAddToSelection(input) {
    } else if splits := strings.Split(input, " "); len(splits) > 1 {
        for _, str := range splits {
            parseAndAddToSelection(str)
        }
    } else if splits := strings.Split(input, "-"); len(splits) == 2 {
        start, startErr := strconv.Atoi(splits[0])
        end, endErr := strconv.Atoi(splits[1])
        if startErr == nil && endErr == nil {
            for i := start; i <= end; i++ {
                addToSelection(i)
            }
        }
    }
    return selections
}


func selectionInteractive(dir string, cmdName string, action func(Crumb) *Crumb) {
    filter := buildFilters(conf.Filters)
    sortFns := buildSorts(conf.Sorts)

    crumbFilePath := filepath.Join(dir, conf.CrumbFileName)
    if fileExists(crumbFilePath) {
        fileContent := readFile(crumbFilePath)

        header := preSufFixString(conf.Header, filepath.Join(crumbFilePath, ".."))
        fmt.Println(header)

        crumbLines := strings.Split(fileContent, "\n")
        crumbs, lineNumbers := getCrumbsFromLines(crumbLines, filter, sortFns, conf)

        printCrumbs(crumbs, true, conf)

        fmt.Printf("%s>> ", cmdName)
        reader := bufio.NewReader(os.Stdin)
        input, _ := reader.ReadString('\n')

        selections := parseSelection(input[:len(input) - 1], lineNumbers)

        newContent := newFileContent(crumbLines, selections, action, conf)
        writeFile(crumbFilePath, newContent)
    }
}

func selection(dir string, input string, action func(Crumb) *Crumb) {
    filter := buildFilters(conf.Filters)
    sortFns := buildSorts(conf.Sorts)

    crumbFilePath := filepath.Join(dir, conf.CrumbFileName)
    if fileExists(crumbFilePath) {
        fileContent := readFile(crumbFilePath)

        crumbLines := strings.Split(fileContent, "\n")
        _, lineNumbers := getCrumbsFromLines(crumbLines, filter, sortFns, conf)

        selections := parseSelection(input, lineNumbers)

        newContent := newFileContent(crumbLines, selections, action, conf)
        writeFile(crumbFilePath, newContent)
    }
}

func interactive(dir string, conf *Config) {
    helpText := "\n*** Commands ***\n  [l]s  [a]d  [m]a  [u]m  [r]m  [b]a  [w]a  [e]d  [f]i\n> "
    ls(dir, conf)
    reader := bufio.NewReader(os.Stdin)
    for true {
        fmt.Printf(helpText)
        input, _ := reader.ReadString('\n')

        cmd := input[:len(input) - 1]
        if cmd == "" {
            return
        } else if cmd == "f" || cmd == "fi" {
            fmt.Print("fi>>")
            input, _ := reader.ReadString('\n')

            stack := NewSimpleStack(strings.Split(input[:len(input) - 1], " "))
            for stack.Size() > 0 {
                arg := stack.Pop()
                if arg == "noFilter" {
                    conf.Filters = []FunctionDesc{}
                } else {
                    if filter, ok := filterMap[arg]; ok {
                        desc := filter.buildDesc(stack)
                        conf.Filters = append(conf.Filters, desc)
                    }
                }
            }
        } else if cmd == "l" || cmd == "ls" {
            ls(dir, conf);
        } else if cmd == "a" || cmd == "ad" {
            fmt.Printf("ad>> ")
            input, _ := reader.ReadString('\n')
            ad(dir, input[:len(input) - 1], conf)
        } else if cmd == "e" || cmd == "ed" {
            setText := func (crumb Crumb) *Crumb {
                fmt.Printf("text>> ")
                input, _ := reader.ReadString('\n')
                text := input[:len(input) - 1]
                if text == "" {
                    text = editWithEditor(crumb.text)
                }
                crumb.text = text
                return &crumb
            }
            selectionInteractive(dir, "ed", setText)
        } else if cmd == "m" || cmd == "ma" {
            mark := func (crumb Crumb) *Crumb {
                var marks []string
                for mark, _ := range conf.Markers {
                    marks = append(marks, mark)
                }
                fmt.Printf("\n*** Marks ***\n  %s\nma>> ", strings.Join(marks, "  "))
                input, _ := reader.ReadString('\n')

                if input != "\n" {
                    marker := markerFromShortHand(input[:len(input) - 1], conf)
                    if marker == "" {
                        fmt.Sprintf("Could not evaluate marker %s\n", input[:len(input) - 1])
                        return &crumb
                    } else {
                        crumb.marker = marker
                        return &crumb
                    }
                }
                return &crumb
            }
            selectionInteractive(dir, "ma", mark)
        } else if cmd == "u" || cmd == "um" {
            unMark := func (crumb Crumb) *Crumb {
                crumb.marker = ""
                return &crumb
            }
            selectionInteractive(dir, "um", unMark)
        } else if cmd == "r" || cmd == "rm" {
            rmCrumb := func (crumb Crumb) *Crumb {
                return nil
            }
            selectionInteractive(dir, "rm", rmCrumb)
        } else if cmd == "b" || cmd == "ba" {
            ba(dir, conf);
        } else if cmd == "w" || cmd == "wa" {
            wa(dir, conf);
        }
    }
}

func wa(dir string, conf *Config) {
    var crumbFilePaths []string
    maxDepth := 3

    var walk func(string, int)
    walk = func (dir string, depth int) {
        if (depth >= maxDepth) {
            return
        }

        files, err := ioutil.ReadDir(dir)
        if err != nil {
            return
        }
        for _, file := range files {
            if file.Name() == conf.CrumbFileName {
                crumbFilePaths = append(crumbFilePaths, filepath.Join(dir, conf.CrumbFileName))
            } else if file.IsDir() {
                walk(filepath.Join(dir, file.Name()), depth + 1)
            }
        }
    }
    walk(filepath.Join(dir), 0)

    filter := buildFilters(conf.Filters)
    sortFns := buildSorts(conf.Sorts)
    for _, crumbFilePath := range crumbFilePaths {
        printCrumbFile(crumbFilePath, filter, sortFns, conf)
    }
}

func ad(dir string, text string, conf *Config) {
    if text == "" {
        text = editWithEditor("")
    }

    crumbFilePath := filepath.Join(dir, conf.CrumbFileName)

    file, err := os.OpenFile(crumbFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(fmt.Sprintf("Unable to access %s for writing", crumbFilePath))
    }

    if _, err := file.Write([]byte(createCrumbEntry(text) + "\n")); err != nil {
        log.Fatal(err)
    }

    if err := file.Close(); err != nil {
        log.Fatal(err)
    }
}

func ed(dir string, args string, text string, conf *Config) {
    setText := func (crumb Crumb) *Crumb {
        if text == "" {
            text = editWithEditor(crumb.text)
        }
        crumb.text = text
        return &crumb
    }

    selection(dir, args, setText)
}

func ma(dir string, marker string, args string, conf *Config) {
    mark := func (crumb Crumb) *Crumb {
        crumb.marker = marker
        return &crumb
    }

    selection(dir, args, mark)
}

func um(dir string, arg string, conf *Config) {
    unMark := func (crumb Crumb) *Crumb {
        crumb.marker = ""
        return &crumb
    }

    selection(dir, arg, unMark)
}

func ls(dir string, conf *Config) {
    crumbFilePath := filepath.Join(dir, conf.CrumbFileName)

    filter := buildFilters(conf.Filters)
    sortFns := buildSorts(conf.Sorts)
    printCrumbFile(crumbFilePath, filter, sortFns, conf)
}

func ba(dir string, conf *Config) {
    crumbFilePaths := findCrumbFiles(dir, conf)

    filter := buildFilters(conf.Filters)
    sortFns := buildSorts(conf.Sorts)
    for _, crumbFilePath := range crumbFilePaths {
        printCrumbFile(crumbFilePath, filter, sortFns, conf)
    }
}

func rm(dir string, arg string, conf *Config) {
    rmCrumb := func (crumb Crumb) *Crumb {
        return nil
    }

    selection(dir, arg, rmCrumb)
}
