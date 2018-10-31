package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/integrii/flaggy"
)

func parseArgs() {
	// Set your program's name and description.  These appear in help output.
	// flaggy.SetName(color.CyanString("%s %s", appname, version))
	flaggy.SetName(appname)
	flaggy.SetDescription("Eat my shorts to be best at sports (" + version + ") ")
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://bitbucket.org/briiC/findthese/\n"
	flaggy.DefaultParser.AdditionalHelpPrepend += strings.Repeat(".", 80)

	// add a global bool flag for fun
	flaggy.String(&argSourcePath, "s", "src", "Source path of directory -- REQUIRED")
	flaggy.String(&argEndpoint, "u", "url", "URL endpoint to hit -- REQUIRED")
	flaggy.String(&argMethod, "m", "method", "HTTP Method to use (default: "+argMethod+")")
	flaggy.String(&argOutput, "o", "output", "Output report to file (default: "+argOutput+")")
	flaggy.String(&argOutput, "z", "delay", "Delay every request for N milliseconds (default: 250)")
	flaggy.StringSlice(&argSkipExts, "", "skip-ext", "Skip files with these extensions (default: png, gif, jpg, jpeg)")

	// set the version and parse all inputs into variables
	flaggy.SetVersion(version)
	flaggy.Parse()

	// On missing params show help
	if argSourcePath == "" || argEndpoint == "" {
		flaggy.ShowHelpAndExit("")
	}

	// Validate
	if err := validateArgs(); err != nil {
		color.Red("\n%v\n\n", err)
		return
	}

}

// Validate arguments
func validateArgs() error {

	// Does source path exists
	if _, err := os.Stat(argSourcePath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		return fmt.Errorf("Source path [-s, --src]: \n\t%v", err)
	}

	// NB! Do not check here if URL is available!
	// Because of different configurations given base URL could not be "200 OK"
	// Also there could be configurations where only valid files gives different response and others fails

	// // Trailing slash - URL must end with slash
	// argEndpoint = strings.TrimSuffix(argEndpoint, "/") + "/"

	// Method uppercase - necessary only for visual appearance
	argMethod = strings.ToUpper(argMethod)

	// Delay
	argDelay = int(math.Abs(float64(argDelay)))

	// Skiped extensions
	var exts []string
	for _, ext := range argSkipExts {
		ext = strings.Trim(ext, " .")
		ext = strings.ToLower(ext)
		ext = "." + ext // must be prefixed with dot
		if ext != "" {
			exts = append(exts, ext)
		}
	}
	argSkipExts = exts

	// No errors
	return nil
}

func printUsedArgs() {
	fmt.Println(strings.Repeat("-", 80))
	color.Cyan("%12s: %s", "Source path", argSourcePath)
	color.Cyan("%12s: %s", "URL", argEndpoint)
	color.Cyan("%12s: %s", "Method", argMethod)
	color.Cyan("%12s: %s", "Output", argOutput)
	color.Cyan("%12s: %d", "Delay (ms)", argDelay)
	color.Cyan("%12s: %v", "Ignored extensions", argSkipExts)
	fmt.Println(strings.Repeat("-", 80))
}
