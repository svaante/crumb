package crumb

import (
    "strconv"
    "sort"
    "path/filepath"
    "fmt"
)

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

func formatCrumb(crumb Crumb, conf *Config) string {
    var str string
    if crumb.marker == "" {
        str = preSufFixString(conf.UnMarked, crumb.text)
    } else {
        str = preSufFixString(conf.Markers[crumb.marker], crumb.text)
    }

    return str
}

func printCrumbFile(crumbFilePath string, filter func (Crumb) bool, sortFns []func(func (int) Crumb) less, conf *Config) {
    if fileExists(crumbFilePath) {
        fileContent := readFile(crumbFilePath)
        crumbs := crumbsFromFileContent(fileContent, conf)

        for _, sortFn := range sortFns {
            sort.SliceStable(crumbs, sortFn(func (i int) Crumb {
                return crumbs[i]
            }))
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
