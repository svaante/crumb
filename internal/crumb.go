package crumb

import (
    "regexp"
    "time"
    "fmt"
    "strings"
    "log"
    "errors"
)

type Crumb struct {
    marker string
    text string
    modifiedDate *time.Time
    createdDate *time.Time
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

func makeCrumb(crumbLine string, conf *Config) (Crumb, error) {
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

    if matches[2] != "" && matches[1] == "" {
        dateCreated := parseDate(matches[2][:len(matches[2]) - 1])
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

func createCrumbEntry(text string) string {
    createDate := formatDate(time.Now())
    return fmt.Sprintf("%s %s", createDate, text)
}
