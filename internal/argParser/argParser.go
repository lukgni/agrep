package argParser

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type positionalArg struct {
	v           *string
	name        string
	description string
}

type booleanFlag struct {
	v           *bool
	name        string
	description string
}

type ArgumentParser struct {
	ProgName         string
	Use              string
	ShortDescription string
	Description      string

	positionalArgs []*positionalArg
	booleanFlags   map[string]*booleanFlag

	unrecognizedArgs  []string
	unrecognizedFlags map[string]struct{}
}

func (p *ArgumentParser) parseFlag(name string) {
	if p.booleanFlags != nil {
		flag, found := p.booleanFlags[name]
		if found {
			*flag.v = true
			return
		}
	}

	if p.unrecognizedFlags != nil {
		p.unrecognizedFlags[name] = struct{}{}
	}
}

func (p *ArgumentParser) parseArg(argIndex int, arg string) {
	if argIndex < len(p.positionalArgs) {
		*p.positionalArgs[argIndex].v = arg
		return
	}

	if p.unrecognizedArgs != nil {
		p.unrecognizedArgs = append(p.unrecognizedArgs, arg)
	}
}

func (p *ArgumentParser) forEachArg(fn func(*positionalArg)) {
	for _, arg := range p.positionalArgs {
		fn(arg)
	}
}

func (p *ArgumentParser) forEachFlag(fn func(*booleanFlag)) {
	var flags []*booleanFlag

	for _, f := range p.booleanFlags {
		flags = append(flags, f)
	}

	sort.Slice(flags, func(i, j int) bool { return flags[i].name < flags[j].name })

	for _, f := range flags {
		fn(f)
	}
}

func (p *ArgumentParser) PositionalArgument(name string, description string) *string {
	var posArg string

	p.positionalArgs = append(p.positionalArgs, &positionalArg{
		v:           &posArg,
		name:        name,
		description: description,
	})

	return &posArg
}

func (p *ArgumentParser) BooleanFlag(name string, description string) *bool {
	var bFlag bool

	if p.booleanFlags == nil {
		p.booleanFlags = map[string]*booleanFlag{}
	}

	p.booleanFlags[name] = &booleanFlag{
		v:           &bFlag,
		name:        name,
		description: description,
	}

	return &bFlag
}

func (p *ArgumentParser) Parse() {
	if p.ProgName == "" {
		p.ProgName = os.Args[0]
	}

	p.unrecognizedArgs = make([]string, 0)
	p.unrecognizedFlags = make(map[string]struct{}, 0)

	parsedArgIndex := 0
	for _, inArg := range os.Args[1:] {
		arg, isFlag := strings.CutPrefix(inArg, "--")
		if isFlag {
			p.parseFlag(arg)
			continue
		}

		p.parseArg(parsedArgIndex, arg)
		parsedArgIndex += 1
	}

	_, helpRequested := p.unrecognizedFlags["help"]

	if parsedArgIndex < len(p.positionalArgs) || helpRequested {
		p.Usage()
		os.Exit(0)
	}
}

func (p *ArgumentParser) Usage() {
	var helpText string

	// 'Usage' section
	helpText += fmt.Sprintf("\nUsage: \t%s ", p.ProgName)
	p.forEachFlag(func(f *booleanFlag) { helpText += fmt.Sprintf("[--%s] ", f.name) })
	p.forEachArg(func(a *positionalArg) { helpText += fmt.Sprintf("%s ", a.name) })
	helpText += "\n"

	// 'Description' [short] section
	if len(p.ShortDescription) > 0 {
		helpText += fmt.Sprintf("\n%s\n", p.ShortDescription)
	}

	// 'Options' section
	if len(p.booleanFlags) > 0 {
		helpText += "\nOptions: \n"
		p.forEachFlag(func(f *booleanFlag) { helpText += fmt.Sprintf("\t--%-10s\t - %s\n", f.name, f.description) })
	}

	// 'Arguments' section
	if len(p.positionalArgs) > 0 {
		helpText += "\nArguments: \n"
		p.forEachArg(func(a *positionalArg) { helpText += fmt.Sprintf("\t  %-10s\t - %s\n", a.name, a.description) })
		helpText += "\n"
	}

	// 'Description' [long] section
	if len(p.Description) > 0 {
		helpText += fmt.Sprintf("\n%s\n", p.Description)
	}

	fmt.Printf(helpText)
}

func (p *ArgumentParser) UnrecognizedArgs() []string {
	return p.unrecognizedArgs
}

func (p *ArgumentParser) UnrecognizedFlags() map[string]struct{} {
	return p.unrecognizedFlags
}
