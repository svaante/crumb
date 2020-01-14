package main

import (
    "os"
    "fmt"
    "path/filepath"
    "github.com/pelletier/go-toml"
    "strings"
    "io/ioutil"
    "regexp"
    "log"
    "errors"
    "strconv"
    "bufio"
    "time"
    "sort"
)

type PreSufFix struct {
    Prefix string
    Suffix string
}

type Function struct {
    Name string
    Args []string
}

type Config struct {
    StopAt string
    CrumbFileName string
    Markers map[string]PreSufFix
    UnMarked PreSufFix
    Header PreSufFix
    Selector PreSufFix
    Filters []Function
    Sorts []Function
}

type Crumb struct {
    marker string
    text string
    modifiedDate *time.Time
    createdDate *time.Time
}

func Unquote(str string) string {
    if strUnqouted, err := strconv.Unquote(str); err == nil {
        return strUnqouted
    } else {
        return str
    }
}

func preSufFixString(preSufFix PreSufFix, str string) string {
    return Unquote(preSufFix.Prefix) + str + Unquote(preSufFix.Suffix)
}

func newDefaultConfig() *Config {
    return &Config{
        StopAt: "/",
        CrumbFileName: ".crumbs",
        Markers: map[string]PreSufFix{"m": PreSufFix{}},
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
    configFileName := ".crumbrc.toml"
    crumbrcFilePath := filepath.Join(getHomePath(), configFileName)

    content, err := ioutil.ReadFile(crumbrcFilePath)
    if err != nil {
        return
    }

    if err = toml.Unmarshal(content, conf); err != nil {
        fmt.Println(err)
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

func formatDate(date time.Time) string {
    return date.Format("2006-01-02 15:04:05")
}

func parseDate(dateString string) time.Time {
    date, err := time.Parse("2006-01-02 15:04:05", dateString)
    if err != nil {
        log.Fatal("Invalid time format")
    }
    return date
}

func stringFromCrumb(crumb Crumb, conf *Config) string {
    var marker string
    if (crumb.marker != "") {
        marker = crumb.marker + " "
    }
    var createdDateString string
    if (crumb.createdDate == nil) {
        createdDateString = formatDate(time.Now()) + " "
    } else {
        createdDateString = formatDate(*crumb.createdDate) + " "
    }
    modifedDateString := formatDate(time.Now()) + " "
    return fmt.Sprintf("%s%s%s%s", modifedDateString, createdDateString, marker, crumb.text)
}

func crumbFromString(crumbLine string, conf *Config) (Crumb, error) {
    var markers []string
    for marker, _ := range conf.Markers {
        markers = append(markers, marker)
    }
    markersRe := fmt.Sprintf("(?:(%s) )", strings.Join(markers, "|"))

    re, err := regexp.Compile(fmt.Sprintf(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} )?(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} )%s?(.*)`, markersRe))
    if err != nil {
        log.Fatal(fmt.Sprintf("Bad `Markers=%s` unable to compile regexp",
                              strings.Join(markers, ", ")))
    }

    matches := re.FindStringSubmatch(crumbLine)
    crumb := Crumb{}

    if len(matches) == 0 {
        return crumb, errors.New("Unable to parse line to crumb")
    }

    if matches[1] != "" && matches[2] == "" {
        dateCreated := parseDate(matches[1][:len(matches[1]) - 1])
        crumb.createdDate = &dateCreated
    } else if matches[1] != "" && matches[2] != "" {
        dateModified := parseDate(matches[1][:len(matches[1]) - 1])
        dateCreated := parseDate(matches[2][:len(matches[2]) - 1])
        crumb.createdDate = &dateCreated
        crumb.modifiedDate = &dateModified
    }
    crumb.marker = matches[3]
    crumb.text = matches[4]

    return crumb, nil
}

func crumbsFromFileContent(crumbContent string, conf *Config) []Crumb {
    crumbLines := strings.Split(crumbContent, "\n")

    var crumbs []Crumb
    for _, crumbLine := range crumbLines {
        if (crumbLine != "") {
            crumb, err := crumbFromString(crumbLine, conf)

            if err == nil {
                crumbs = append(crumbs, crumb)
            }
        }
    }

    return crumbs
}

func formatCrumb(crumb Crumb, conf *Config) string {
    var str string
    if crumb.marker == "" {
        str = preSufFixString(conf.UnMarked, crumb.text)
    } else {
        str = preSufFixString(conf.Markers[crumb.marker], crumb.text)
    }

    return str
}

func getWD() string {
    dir, err := os.Getwd()
    if err != nil {
        log.Fatal("Failed getting working directory")
    }
    return dir
}

func findCrumbFiles(dir string, conf *Config) []string {
    var crumbFilePaths []string
    for basePath := dir; basePath != conf.StopAt; basePath = filepath.Join(basePath, "..") {
        crumbFilePath := filepath.Join(basePath, conf.CrumbFileName)
        crumbFilePaths = append(crumbFilePaths, crumbFilePath)
    }

    return crumbFilePaths
}

func printCrumbFile(crumbFilePath string, filter func (Crumb) bool, sortFns []func([]Crumb) func(int, int) bool, conf *Config) {
    if fileExists(crumbFilePath) {
        fileContent := readFile(crumbFilePath)
        crumbs := crumbsFromFileContent(fileContent, conf)
        for _, sortFn := range sortFns {
            sort.SliceStable(crumbs, sortFn(crumbs))
        }

        header := preSufFixString(conf.Header, filepath.Join(crumbFilePath, ".."))
        fmt.Println(header)
        for _, crumb := range crumbs {
            if filter(crumb) {
                crumbString := formatCrumb(crumb, conf)
                fmt.Printf("%s\n", crumbString)
            }
        }
    }
}

func ls(dir string, conf *Config) {
    crumbFilePath := filepath.Join(dir, conf.CrumbFileName)

    filter := buildFilters(conf.Filters)
    sortFns := buildSorts(conf.Sorts)
    printCrumbFile(crumbFilePath, filter, sortFns, conf)
}

func fl(dir string, conf *Config) {
    crumbFilePaths := findCrumbFiles(dir, conf)

    filter := buildFilters(conf.Filters)
    sortFns := buildSorts(conf.Sorts)
    for _, crumbFilePath := range crumbFilePaths {
        printCrumbFile(crumbFilePath, filter, sortFns, conf)
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

func createCrumbEntry(text string) string {
    createDate := formatDate(time.Now())
    return fmt.Sprintf("%s %s", createDate, text)
}

func ad(dir string, text string, conf *Config) {
    if text == "" {
        return
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

func getCrumbsFromLines(crumbLines []string, filter func(Crumb) bool, sortFns []func([]Crumb) func(int, int) bool, conf *Config) ([]Crumb, []int) {
    var zip []struct{Crumb; int}
    var tmpCrumbs []Crumb

    for lineNumber, crumbLine := range crumbLines {
        if (crumbLine != "") {
            crumb, err := crumbFromString(crumbLine, conf)
            if err == nil && filter(crumb) {
                tmpCrumbs = append(tmpCrumbs, crumb)
                zip = append(zip, struct{Crumb; int}{crumb, lineNumber})
            }
        }
    }
    for _, sortFn := range sortFns {
        sort.SliceStable(zip, sortFn(tmpCrumbs))
    }
    var crumbs []Crumb
    var lineNumbers []int

    for _, e := range zip {
        crumbs = append(crumbs, e.Crumb)
        lineNumbers = append(lineNumbers, e.int)
    }
    return crumbs, lineNumbers
}

func printCrumbs(crumbs []Crumb, withSelectors bool, conf *Config) {
    for i, crumb := range crumbs {
        crumbString := formatCrumb(crumb, conf)
        if withSelectors {
            selector := preSufFixString(conf.Selector,  strconv.Itoa(i + 1))
            fmt.Printf("%s%s\n", selector, crumbString)
        } else {
            fmt.Printf("%s\n", crumbString)
        }
    }
}

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

func newFileContent(crumbLines []string, selections []int, action func(Crumb) *Crumb, conf *Config) string {
    for _, lineNumber := range selections {
        crumb, err := crumbFromString(crumbLines[lineNumber], conf)

        if err != nil {
            continue
        }

        newCrumb := action(crumb)

        if newCrumb == nil {
            crumbLines[lineNumber] = ""
        } else if crumb != *newCrumb {
            crumbLines[lineNumber] = stringFromCrumb(*newCrumb, conf)
        }
    }

    var newFileContent string
    for _, crumbLine := range crumbLines {
        if crumbLine != "" {
            newFileContent += crumbLine + "\n"
        }
    }
    return newFileContent
}

func selectionInteractive(dir string, cmdName string, filter func(Crumb) bool, sortFns []func([]Crumb) func(int, int) bool, action func(Crumb) *Crumb, conf *Config) {
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

func selection(dir string, input string, filter func(Crumb) bool, sortFns []func([]Crumb) func(int, int) bool, action func(Crumb) *Crumb, conf *Config) {
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

func isNot(args []string) func(Crumb) bool {
    return func(crumb Crumb) bool {
        for _, marker := range args {
            if marker == crumb.marker {
                return false
            }
        }
        return true
    }
}

func all(crumb Crumb) bool {
    return true
}

func delete(crumb Crumb) *Crumb {
    return nil
}

func rm(dir string, arg string, conf *Config) {
    sortFns := buildSorts(conf.Sorts)
    selection(dir, arg, all, sortFns, delete, conf)
}

func isNotMarked(crumb Crumb) bool {
    return crumb.marker == ""
}

func buildMarkFn(marker string) func (Crumb) *Crumb {
    return func (crumb Crumb) *Crumb {
        crumb.marker = marker
        return &crumb
    }
}

func isMarked(crumb Crumb) bool {
    return crumb.marker != ""
}

func unMark(crumb Crumb) *Crumb {
    crumb.marker = ""
    return &crumb
}

func sortNone(_ []Crumb) func (int, int) bool {
    return func (i, j int) bool {
        return i < j
    }
}

func sortReverse(_ []Crumb) func (int, int) bool {
    return func (i, j int) bool {
        return i > j
    }
}

func sortMarked(crumbs []Crumb) func (int, int) bool {
    return func (i, j int) bool {
        return crumbs[i].marker == "" && crumbs[j].marker != ""
    }
}

func buildSorts(sortFunctions []Function) []func ([]Crumb) func (int, int) bool {
    var sortFns []func ([]Crumb) func (int, int) bool
    for _, sortFn := range sortFunctions {
        if sortFn.Name == "sortNone" {
            sortFns = append(sortFns, sortNone)
        } else if sortFn.Name == "sortReverse" {
            sortFns = append(sortFns, sortReverse)
        } else if sortFn.Name == "sortMarked" {
            sortFns = append(sortFns, sortMarked)
        }
    }
    return sortFns
}

func buildFilters(filterFunctions []Function) func (Crumb) bool {
    var filters []func (Crumb) bool
    for _, filterFn := range filterFunctions {
        if filterFn.Name == "isMarked" {
            filters = append(filters, isMarked)
        } else if filterFn.Name == "isNotMarked" {
            filters = append(filters, isNotMarked)
        } else if filterFn.Name == "all" {
            filters = append(filters, all)
        } else if filterFn.Name == "isNot" {
            filters = append(filters, isNot(filterFn.Args))
        }
    }
    return func (crumb Crumb) bool {
        ret := true
        for _, filter := range filters {
            ret = ret && filter(crumb)
        }
        return ret
    }
}

func ma(dir string, mark string, args string, conf *Config) {

    filter := func (crumb Crumb) bool {
        return crumb.marker != mark
    }
    sortFns := buildSorts(conf.Sorts)
    selection(dir, args, filter, sortFns, buildMarkFn(mark), conf)
}

func um(dir string, arg string, conf *Config) {
    sortFns := buildSorts(conf.Sorts)
    selection(dir, arg, isMarked, sortFns, unMark, conf)
}

func interactive(dir string, conf *Config) {
    helpText := "\n*** Commands ***\n  [l]s  [a]d  [m]a  [u]m  [r]m  [f]l  [w]a\n> "
    ls(dir, conf)
    reader := bufio.NewReader(os.Stdin)
    for true {
        fmt.Printf(helpText)
        input, _ := reader.ReadString('\n')

        cmd := input[:len(input) - 1]
        if (cmd == "") {
            return
        } else if (cmd == "l" || cmd == "ls") {
            ls(dir, conf);
        } else if (cmd == "a" || cmd == "ad") {
            fmt.Printf("ad>> ")
            input, _ := reader.ReadString('\n')
            ad(dir, input[:len(input) - 1], conf)
        } else if (cmd == "m" || cmd == "ma") {
            var marks []string
            for mark, _ := range conf.Markers {
                marks = append(marks, mark)
            }
            fmt.Printf("\n*** Marks ***\n  %s\nmarks>> ", strings.Join(marks, "  "))
            input, _ := reader.ReadString('\n')

            if input != "\n" {
                marker := markerFromShortHand(input[:len(input) - 1], conf)
                if marker == "" {
                    fmt.Sprintf("Could not evaluate marker %s\n", input[:len(input) - 1])
                } else {
                    filter := func (crumb Crumb) bool {
                        return crumb.marker != marker
                    }
                    sortFns := buildSorts(conf.Sorts)
                    selectionInteractive(dir, "ma", filter, sortFns, buildMarkFn(marker), conf)
                }
            }
        } else if (cmd == "u" || cmd == "um") {
            sortFns := buildSorts(conf.Sorts)
            selectionInteractive(dir, "um", isMarked, sortFns, unMark, conf)
        } else if (cmd == "r" || cmd == "rm") {
            sortFns := buildSorts(conf.Sorts)
            selectionInteractive(dir, "rm", all, sortFns, delete, conf)
        } else if (cmd == "f" || cmd == "fl") {
            fl(dir, conf);
        } else if (cmd == "w" || cmd == "wa") {
            wa(dir, conf);
        }
    }
}

func parseDirAndArgs(args []string) (string, []string) {
    var dir string
    var input []string
    for i, arg := range args {
        if i == 0 {
            var err error
            dir, err = getValidDir(arg)
            if err != nil {
                dir = getWD()
                input = append(input, arg)
            }
        } else {
            input = append(input, arg)
        }
    }
    return dir, input
}

func getShortHandMarker(conf *Config) map[string]string {
    var markers []string
    shortHand := make(map[string]string)

    for marker, _ := range conf.Markers {
        shortHand[marker] = marker
        markers = append(markers, marker)
    }
    for _, marker := range markers {
        var partial string
        for _, c := range marker[:len(marker) - 1] {
            partial += fmt.Sprintf("%c", c)
            _, found := shortHand[partial]
            if !found {
                shortHand[partial] = marker
            }
        }
    }
    return shortHand
}

func markerFromShortHand(markerShortHand string, conf *Config) string {
    shortHandMap := getShortHandMarker(conf)
    marker := shortHandMap[markerShortHand]
    return marker
}

func main() {
    conf := newDefaultConfig()
    applyUserConfig(conf)

    helpText := fmt.Sprintf(`
Usage: crumbs [OPTIONS] COMMAND [DIR] [CRUMB_BITS ...]

COMMAND:
    ls
        Lists crumbs in "DIR/%s"
    wa
        Follows the bread crumb from "DIR" N deep
    fl
        Follows the bread crumb from "DIR" and up to "%s"
    ad
        Add a crumb from CRUMB_BITS in "DIR/%s" does its best to difirentiate CRUMB_BITS from PATH
    ma
        Mark crumb in "DIR/%s" as done/invalid/archived/... depending on the your metafysical understanding of crumbs
    ua
        Unmark crumb in "DIR/%s", what unmark means still depends on the your metafysical understanding of crumbs
    rm
        Remove crumb (eat?) in "DIR/%s"
    help
        prints this
DIR:
    path to dir defaults to "."

crumbs sports a config file at "$HOME/.crumbrc.json"`,
        conf.CrumbFileName,
        conf.StopAt,
        conf.CrumbFileName,
        conf.CrumbFileName,
        conf.CrumbFileName,
        conf.CrumbFileName)


    args := os.Args[1:]

    if len(args) == 0 {
        fmt.Println(helpText)
        os.Exit(1)
    }

    cmd := args[0]
    if cmd == "ls" {
        var dir string
        if len(args) > 1 {
            var err error
            dir, err = getValidDir(args[1])
            if err != nil {
                log.Fatal(fmt.Sprintf("%s is not a valid dir", args[1]))
            }
        } else {
            dir = getWD()
        }

        ls(dir, conf)
    } else if cmd == "wa" {
        var dir string
        if len(args) > 1 {
            var err error
            dir, err = getValidDir(args[1])
            if err != nil {
                log.Fatal(fmt.Sprintf("%s is not a valid dir", args[1]))
            }
        } else {
            dir = getWD()
        }

        wa(dir, conf)
    } else if cmd == "fl" {
        var dir string
        if len(args) > 1 {
            var err error
            dir, err = getValidDir(args[1])
            if err != nil {
                log.Fatal(fmt.Sprintf("%s is not a valid dir", args[1]))
            }
        } else {
            dir = getWD()
        }

        fl(dir, conf)
    } else if cmd == "ad" {
        if len(args) < 2 {
            log.Fatal(fmt.Sprintf("What is an empty crumb?"))
        }
        dir, crumbText := parseDirAndArgs(args[1:])
        ad(dir, strings.Join(crumbText, " "), conf)
    } else if cmd == "ma" {
        if len(args) < 3 {
            log.Fatal(fmt.Sprintf("Cannot mark without a marker and/or number selection"))
        }

        dir, markerAndSelection := parseDirAndArgs(args[1:])

        markerShortHand := markerAndSelection[0]
        selection := strings.Join(markerAndSelection[1:], " ")
        marker := markerFromShortHand(markerShortHand, conf) 

        if marker == "" {
            log.Fatal(fmt.Sprintf("Could not evaluate marker %s", markerShortHand))
        }

        ma(dir, marker, selection, conf)
    } else if cmd == "um" {
        if len(args) < 2 {
            log.Fatal(fmt.Sprintf("Cannot unmark without a number selection"))
        }
        dir, selections := parseDirAndArgs(args[1:])
        um(dir, strings.Join(selections, " "), conf)
    } else if cmd == "rm" {
        if len(args) < 2 {
            log.Fatal(fmt.Sprintf("Cannot remove crumbs without a number selection"))
        }
        dir, selections := parseDirAndArgs(args[1:])
        rm(dir, strings.Join(selections, " "), conf)
    } else if cmd == "i" {
        var dir string
        if len(args) > 1 {
            var err error
            dir, err = getValidDir(args[1])
            if err != nil {
                log.Fatal(fmt.Sprintf("%s is not a valid dir", args[1]))
            }
        } else {
            dir = getWD()
        }

        interactive(dir, conf)
    } else if cmd == "help" {
        fmt.Println(helpText)
    } else {
        log.Fatal(fmt.Sprintf(`Unrecognized command %s. See 'crumb help'`, cmd))
    }
}
