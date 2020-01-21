package crumb

import (
    "time"
    "strconv"
    "fmt"
    "log"
    "strings"
)

type filter func(Crumb) bool

type filterFnI interface {
    applyFn([]string) filter
    buildDesc(*SimpleStack) FunctionDesc
}

type filterFn struct {
    name string
    fn func () filter
}

type filterArgsFn struct {
    name string
    fn func ([]string) filter
}

var filterMap = map[string]filterFnI{
    "isMarked": filterFn{
        name: "isMarked",
        fn: isMarked,
    },
    "isNotMarked": filterFn{
        name: "isNotMarked",
        fn: isNotMarked,
    },
    "isCreatedSinceH": filterArgsFn{
        name: "isCreatedSinceH",
        fn: isCreatedSinceH,
    },
    "isModifiedSinceH": filterArgsFn{
        name: "isModifiedSinceH",
        fn: isModifiedSinceH,
    },
    "isNot": filterArgsFn{
        name: "isNot",
        fn: isNot,
    },
    "is": filterArgsFn{
        name: "is",
        fn: is,
    },
}

func (f filterFn) buildDesc(_ *SimpleStack) FunctionDesc {
    return FunctionDesc{
        Name: f.name,
    };
}

func (f filterArgsFn) buildDesc(s *SimpleStack) FunctionDesc {
    if s.Size() == 0 {
        log.Panic(fmt.Sprintf("Filter %s needs atleast one arg", f.name))
    }
    return FunctionDesc{
        Name: f.name,
        Args: strings.Split(s.Pop(), ","),
    };
}

func (f filterFn) applyFn(_ []string) filter {
    return f.fn()
}

func (f filterArgsFn) applyFn(args []string) filter {
    return f.fn(args)
}

func buildFilters(filterFunctions []FunctionDesc) func (Crumb) bool {
    var filters []func (Crumb) bool
    for _, filterFn := range filterFunctions {
        if filter, found := filterMap[filterFn.Name]; found {
            filters = append(filters, filter.applyFn(filterFn.Args))
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

func is(args []string) filter {
    return func(crumb Crumb) bool {
        for _, marker := range args {
            if markerFromShortHand(marker, conf) == crumb.marker {
                return true
            }
        }
        return false
    }
}

func isCreatedSinceH(args []string) filter {
    if len(args) != 1 {
        log.Fatal(fmt.Sprintf("Filter isCreatedSinceH only excepts 1 arg not %d", len(args[0])))
    }
    i, err := strconv.Atoi(args[0])
    if err != nil {
        log.Fatal(fmt.Sprintf("Could not parse arg %s to int isCreatedSinceH", args[0]))
    }

    return func (crumb Crumb) bool {
        dT := time.Since(*crumb.createdDate)
        dH := dT.Hours()
        return dH <= float64(i)
    }
}

func isModifiedSinceH(args []string) filter {
    if len(args) != 1 {
        log.Fatal(fmt.Sprintf("Filter isModifiedSinceH only excepts 1 arg not %d", len(args[0])))
    }
    i, err := strconv.Atoi(args[0]); if err != nil {
        log.Fatal(fmt.Sprintf("Could not parse arg %s to int isModifiedSinceH", args[0]))
    }

    return func (crumb Crumb) bool {
        dT := time.Since(*crumb.modifiedDate )
        dH := dT.Hours()
        return dH <= float64(i)
    }
}

func isNot(args []string) filter {
    return func(crumb Crumb) bool {
        for _, marker := range args {
            if markerFromShortHand(marker, conf) == crumb.marker {
                return false
            }
        }
        return true
    }
}

func isMarked() filter {
    return func(crumb Crumb) bool {
        return crumb.marker != ""
    }
}

func isNotMarked() filter {
    return func (crumb Crumb) bool {
        return crumb.marker == ""
    }
}
