package crumb

import (
    "strings"
    "path/filepath"
    "sort"
)

func crumbsFromFileContent(crumbContent string, conf *Config) []Crumb {
    crumbLines := strings.Split(crumbContent, "\n")

    var crumbs []Crumb
    for _, crumbLine := range crumbLines {
        if (crumbLine != "") {
            crumb, err := makeCrumb(crumbLine, conf)

            if err == nil {
                crumbs = append(crumbs, crumb)
            }
        }
    }

    return crumbs
}

func findCrumbFiles(dir string, conf *Config) []string {
    var crumbFilePaths []string
    for basePath := dir; basePath != conf.StopAt; basePath = filepath.Join(basePath, "..") {
        crumbFilePath := filepath.Join(basePath, conf.CrumbFileName)
        crumbFilePaths = append(crumbFilePaths, crumbFilePath)
    }

    return crumbFilePaths
}

func getCrumbsFromLines(crumbLines []string, filter func(Crumb) bool, sortFns []func(func (int) Crumb) less, conf *Config) ([]Crumb, []int) {
    var zip []struct{Crumb; int}

    for lineNumber, crumbLine := range crumbLines {
        if (crumbLine != "") {
            if crumb, err := makeCrumb(crumbLine, conf); err == nil {
                zip = append(zip, struct{Crumb; int}{crumb, lineNumber})
            }
        }
    }

    for _, sortFn := range sortFns {
        sort.SliceStable(zip, sortFn(func (i int) Crumb {
            return zip[i].Crumb
        }))
    }

    var crumbs []Crumb
    var lineNumbers []int

    for _, e := range zip {
        if filter(e.Crumb) {
            crumbs = append(crumbs, e.Crumb)
            lineNumbers = append(lineNumbers, e.int)
        }
    }

    return crumbs, lineNumbers
}

func newFileContent(crumbLines []string, selections []int, action func(Crumb) *Crumb, conf *Config) string {
    for _, lineNumber := range selections {
        crumb, err := makeCrumb(crumbLines[lineNumber], conf)

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
