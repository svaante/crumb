package crumb

import (
    "fmt"
    "strings"
    "log"
)

var conf *Config

type CliArg struct {
    do func(*SimpleStack)
    help string
}

func parseDir(args *SimpleStack) string  {
    if args.Size() > 1 {
        dir, err := getValidDir(args.Peek())
        if err == nil {
            args.Pop()
            return dir
        }
    }
    return getWD()
}

func parseMarker(args *SimpleStack) string {
    if args.Size() == 0 {
            log.Fatal(fmt.Sprintf("Cannot mark without a marker"))
        }

        marker := markerFromShortHand(args.Pop(), conf)
        if marker == "" {
            log.Fatal(fmt.Sprintf("Could not evaluate marker %s", marker))
        }
        return marker
}

func parseString(args *SimpleStack) string  {
    if args.Size() == 0 {
        log.Fatal("Needs matching arg value")
    }
    return args.Pop()
}

func parseRest(args *SimpleStack) string  {
    return strings.Join(args.Empty(), " ")
}

func printHelp() {
    fmt.Println(fmt.Sprintf(`
Usage: crumb [OPTIONS] COMMAND

COMMAND:
    ls
        Lists crumbs in "DIR/%s"
    wa
        Follows the bread crumb from "DIR" N deep
    ba
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

crumb sports a config file at "$HOME/.crumbrc.json"`,
        conf.CrumbFileName,
        conf.StopAt,
        conf.CrumbFileName,
        conf.CrumbFileName,
        conf.CrumbFileName,
        conf.CrumbFileName))
}

func Parse(args []string) {
    conf = newDefaultConfig()
    applyUserConfig(conf)

    flags := (map[string]CliArg{
        "--noFilter": CliArg{
            do: func (_ *SimpleStack) {
                conf.Filters = []FunctionDesc{}
            },
            help: "",
        },
        "--noSort": CliArg{
            do: func (_ *SimpleStack) {
                conf.Sorts = []FunctionDesc{}
            },
            help: "",
        },
    })

    for fName, f := range filterMap {
        flags["--" + fName] = CliArg{
            do: func (f filterFnI) func (*SimpleStack) {
                    return func (args *SimpleStack) {
                        desc := f.buildDesc(args)
                        conf.Filters = append(conf.Filters, desc)
                    }
                }(f),
            help: "",
        }
    }

    for sName, s := range sortMap {
        flags["--" + sName] = CliArg{
            do: func (s lessFnI) func (*SimpleStack) {
                    return func (args *SimpleStack) {
                        desc := s.buildDesc(args)
                        conf.Sorts = append(conf.Sorts, desc)
                    }
                }(s),
            help: "",
        }
    }

    for _, alias := range conf.Alias {
        flags[alias.Name] = CliArg{
            do: func (args *SimpleStack) {
                args.Prepend(alias.Args)
            },
            help: "",
        }
    }

    var cmds map[string]CliArg
    cmds = map[string]CliArg{
        "ls": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                ls(dir, conf)
            },
            help: "ls [PATH]",
        },
        "wa": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                wa(dir, conf)
            },
            help: "wa [PATH]",
        },
        "ba": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                ba(dir, conf)
            },
            help: "ba [PATH]",
        },
        "ad": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                text := parseRest(args)
                ad(dir, text, conf)
            },
            help: "add [PATH] [...CRUMB_BITS]",
        },
        "rm": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                selection := parseRest(args)
                rm(dir, selection, conf)
            },
            help: "rm [PATH] [...CRUMB_SELECTION]",
        },
        "ma": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                marker := parseMarker(args)
                selection := parseRest(args)
                ma(dir, marker, selection, conf)
            },
            help: "ma [PATH] <marker> [...CRUMB_SELECTION]",
        },
        "um": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                selection := parseRest(args)
                um(dir, selection, conf)
            },
            help: "um [PATH] [...CRUMB_SELECTION]",
        },
        "ed": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                selection := parseString(args)
                edit := parseRest(args)
                ed(dir, selection, edit, conf)
            },
            help: "ed [PATH] <...CRUMB_SELECTION> [...CRUMB_BITS]",
        },
        "i": {
            do: func (args *SimpleStack) {
                dir := parseDir(args)
                interactive(dir, conf)
            },
            help: "it [PATH]",
        },
        "help": {
            do: func (args *SimpleStack) {
                if args.Size() > 0 {
                    cmd, found := cmds[args.Peek()]
                    if found {
                        fmt.Println(cmd.help)
                    } else {
                        fmt.Println(fmt.Sprintf(`Unrecognized command help %s. See 'crumb help'`, args.Peek()))
                    }
                } else {
                    printHelp()
                }
            },
            help: "help <CMD>",
        },
    }

    if len(args) == 0 {
        printHelp()
    }

    argStack := NewSimpleStack(args)
    for argStack.Size() > 0 {
        arg := argStack.Pop()
        if flag, found := flags[arg]; found {
            flag.do(argStack)
        } else {
            cmd, found := cmds[arg]
            if found {
                cmd.do(argStack)
            } else {
                fmt.Println(fmt.Sprintf(`Unrecognized command/flag %s. See 'crumb help'`, arg))
            }
            break
        }
    }

}
