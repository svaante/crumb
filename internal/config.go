package crumb

import (
    "fmt"
    "io/ioutil"
    "log"
    "path/filepath"

    "github.com/pelletier/go-toml"
)

type PreSufFix struct {
    Prefix string
    Suffix string
}

type FunctionDesc struct {
    Name string
    Args []string
}

type Config struct {
    StopAt string
    CrumbFileName string
    Alias []FunctionDesc
    Filters []FunctionDesc
    Sorts []FunctionDesc
    Markers map[string]PreSufFix
    UnMarked PreSufFix
    Header PreSufFix
    Selector PreSufFix
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

func newDefaultConfig() *Config {
    return &Config{
        StopAt: "/",
        CrumbFileName: ".crumb",
        Markers: map[string]PreSufFix{"m": PreSufFix{}},
    }
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
