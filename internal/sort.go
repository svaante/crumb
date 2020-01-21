package crumb

import (
    "log"
    "fmt"
    "strings"
)

type less func (int, int) bool
type lessFnI interface {
    applyFn([]string) func (func (int) Crumb) less
    buildDesc(*SimpleStack) FunctionDesc
}

type lessFn struct {
    name string
    fn func (func (int) Crumb) less
}

type lessArgsFn struct {
    name string
    fn func ([]string) func (func (int) Crumb) less
}

var sortMap = map[string]lessFnI{
    "sortNone": lessFn{
        name: "sortNone",
        fn: sortNone,
    },
    "sortReverse": lessFn{
        name: "sortReverse",
        fn: sortReverse,
    },
    "sortMarked": lessFn{
        name: "sortMarked",
        fn: sortMarked,
    },
    "sortMarkedOrder": lessArgsFn{
        name: "sortMarkedOrder",
        fn: sortMarkedOrder,
    },
}

func (f lessFn) buildDesc(_ *SimpleStack) FunctionDesc {
    return FunctionDesc{
        Name: f.name,
    };
}

func (f lessArgsFn) buildDesc(s *SimpleStack) FunctionDesc {
    if s.Size() == 0 {
        log.Panic(fmt.Sprintf("Sort %s needs atleast one arg", f.name))
    }
    return FunctionDesc{
        Name: f.name,
        Args: strings.Split(s.Pop(), ","),
    };
}

func (f lessFn) applyFn(_ []string) func (func(int) Crumb) less {
    return f.fn
}

func (f lessArgsFn) applyFn(args []string) func (func(int) Crumb) less {
    return f.fn(args)
}

func sortNone(_ func (int) Crumb) less {
    return func (i, j int) bool {
        return i < j
    }
}

func sortReverse(_ func (int) Crumb) less {
    return func (i, j int) bool {
        return i > j
    }
}

func sortMarked(crumbs func (int) Crumb) less {
    return func (i, j int) bool {
        return crumbs(i).marker == "" && crumbs(j).marker != ""
    }
}

func sortMarkedOrder(args []string) func (func (int) Crumb) less {
    return func (crumbs func (int) Crumb) less {
        return func (i, j int) bool {
            for _, marker := range args {
                if crumbs(i).marker == marker {
                    return true
                }
                if crumbs(j).marker == marker {
                    return false
                }
            }
            return false
        }
    }
}

func buildSorts(sortFunctions []FunctionDesc) []func (func (int) Crumb) less {
    var sortFns []func (func (int) Crumb) less
    for _, sortFn := range sortFunctions {
        if sort, ok := sortMap[sortFn.Name]; ok {
            sortFns = append(sortFns, sort.applyFn(sortFn.Args))
        }
    }
    return sortFns
}
