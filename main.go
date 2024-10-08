package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/HeadspaceMeditation/terraform-config-inspect/tfconfig"
	flag "github.com/spf13/pflag"
)

var showJSON = flag.Bool("json", false, "produce JSON-formatted output")
var showVersion = flag.Bool("version", false, "print version of terraform-config-inspect")

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println("0.3.1")
		return
	}

	var dir string
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	} else {
		dir = "."
	}

	module, _ := tfconfig.LoadModule(dir)

	if *showJSON {
		showModuleJSON(module)
	} else {
		showModuleMarkdown(module)
	}

	if module.Diagnostics.HasErrors() {
		os.Exit(1)
	}
}

func showModuleJSON(module *tfconfig.Module) {
	j, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error producing JSON: %s\n", err)
		os.Exit(2)
	}
	os.Stdout.Write(j)
	os.Stdout.Write([]byte{'\n'})
}

func showModuleMarkdown(module *tfconfig.Module) {
	err := tfconfig.RenderMarkdown(os.Stdout, module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error rendering template: %s\n", err)
		os.Exit(2)
	}
}
