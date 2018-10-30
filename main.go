package main

import (
	"strings"

	"github.com/fatih/color"
	"github.com/integrii/flaggy"
)

var appname = "findthese"
var version = "v0.1"

// flags
var sourcePath string
var endpoint string

func init() {
	// Set your program's name and description.  These appear in help output.
	// flaggy.SetName(color.CyanString("%s %s", appname, version))
	flaggy.SetName(appname)
	flaggy.SetDescription("Eat my shorts to be best at sports (" + version + ") ")
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://bitbucket.org/briiC/findthese/\n"
	flaggy.DefaultParser.AdditionalHelpPrepend += strings.Repeat(".", 80)

	// add a global bool flag for fun
	flaggy.String(&sourcePath, "s", "src", "Source path of directory")
	flaggy.String(&endpoint, "u", "url", "URL endpoint to hit")

	// set the version and parse all inputs into variables
	flaggy.SetVersion(version)
	flaggy.Parse()

	// On missing params show help
	if sourcePath == "" || endpoint == "" {
		flaggy.ShowHelpAndExit("")
	}

}

func main() {

	color.Cyan("-- %s", sourcePath)
	color.Cyan("-- %s", endpoint)

}
