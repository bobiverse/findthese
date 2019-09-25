package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/integrii/flaggy"
)

func parseArgs() {
	// Set your program's name and description.  These appear in help output.
	// flaggy.SetName(color.CyanString("%s %s", appname, version))
	flaggy.SetName(appname)
	flaggy.SetDescription("Eat my shorts to be best at sports (" + version + ") ")
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://github.com/briiC/findthese\n"
	flaggy.DefaultParser.AdditionalHelpPrepend += strings.Repeat(".", 80)

	// add a global bool flag for fun
	flaggy.String(&argSourcePath, "s", "src", "Source path of directory -- REQUIRED")
	flaggy.String(&argEndpoint, "u", "url", "URL endpoint to hit -- REQUIRED")
	flaggy.String(&argMethod, "m", "method", "HTTP Method to use (default: "+argMethod+")")
	flaggy.String(&argOutput, "o", "output", "Output report to file (default: "+argOutput+")")
	flaggy.Int(&argDepth, "", "depth", "How deep go in folders. '0' no limit  (default: "+fmt.Sprintf("%d", argDepth)+")")
	flaggy.Int(&argDelay, "z", "delay", "Delay every request for N milliseconds (default: "+fmt.Sprintf("%d", argDelay)+")")
	flaggy.Int(&argTimeout, "", "timeout", "Timeout (seconds) to wait for response  (default: "+fmt.Sprintf("%d", argTimeout)+")")
	flaggy.StringSlice(&argSkip, "", "skip", "Skip files with these extensions (default: "+fmt.Sprintf("%v", argSkip)+")")
	flaggy.StringSlice(&argSkipExts, "", "skip-ext", "Skip files with these extensions (default: "+fmt.Sprintf("%v", argSkipExts)+")")
	flaggy.StringSlice(&argSkipCodes, "", "skip-code", "Skip responses with this response HTTP code (default: "+fmt.Sprintf("%v", argSkipCodes)+")")
	flaggy.StringSlice(&argSkipSizes, "", "skip-size", "Skip responses with this body size (default: "+fmt.Sprintf("%v", argSkipSizes)+")")
	flaggy.Bool(&argDirOnly, "", "dir-only", "Scan directories only")
	flaggy.String(&argUserAgent, "", "user-agent", "User-Agent used")
	flaggy.String(&argCookieString, "C", "cookie", "Cookie string sent with requests")
	flaggy.String(&argHeaderString, "H", "headers", "Custom Headers sent with requests")

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

	if argUserAgent == "random" || argUserAgent == "" {
		argUserAgent = randomUserAgent()
	}

	// parse header string to map of key and values
	// "k1:v1; k2:v2\n k3=v3 " (note "\n" and "=" characters)
	if argHeaderString != "" {
		// Replace new line "\n" to semicolon
		argHeaderString = strings.Replace(argHeaderString, "\\n", ";", -1)

		// Split to pairs
		pairs := strings.Split(argHeaderString, ";") // ["k1=v1", "k2=val2", "k3=v3"] (3)

		// reconstruct string from only valid parts
		argHeaderString = ""

		for _, pair := range pairs {

			// If pair doesn't hold colon ":" try to replace "=" to it
			if !strings.Contains(pair, ":") {
				pair = strings.Replace(pair, "=", ":", 1)
			}

			// Make sure there is two parts: key and value
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) != 2 {
				parts = append(parts, "")
			}

			hKey := strings.TrimSpace(parts[0])
			hVal := strings.TrimSpace(parts[1])
			mHeaders[hKey] = hVal
			argHeaderString += fmt.Sprintf("%s:%s; ", hKey, hVal)
		}

	}

}

// Validate arguments
func validateArgs() error {

	// Does source path exists
	if _, err := os.Stat(argSourcePath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		return fmt.Errorf("Source path [-s, --src]: \n\t%v", err)
	}
	argSourcePath, _ = filepath.Abs(argSourcePath)
	argSourcePath += "/"

	// NB! Do not check here if URL is available!
	// Because of different configurations given base URL could not be "200 OK"
	// Also there could be configurations where only valid files gives different response and others fails

	// // Trailing slash - URL must end with slash
	// argEndpoint = strings.TrimSuffix(argEndpoint, "/") + "/"

	// Method uppercase - necessary only for visual appearance
	argMethod = strings.ToUpper(argMethod)

	// Depth
	if argDepth < 0 {
		argDepth = 0
	}

	// Delay
	argDelay = int(math.Abs(float64(argDelay)))

	// Timeout
	argTimeout = int(math.Abs(float64(argTimeout)))

	// Skpi files/dirs
	argSkip = normalizeArgSlice(argSkip)

	// Skiped extensions
	var exts []string
	argSkipExts = normalizeArgSlice(argSkipExts)
	for _, ext := range argSkipExts {
		ext = strings.Trim(ext, " .")
		ext = strings.ToLower(ext)
		ext = "." + ext // must be prefixed with dot
		if ext != "" {
			exts = append(exts, ext)
		}
	}
	argSkipExts = exts

	// Skiped sizes
	var sizes []string
	argSkipSizes = normalizeArgSlice(argSkipSizes)
	for _, s := range argSkipSizes {

		// Range definitions 100-200
		if strings.Contains(s, "-") {
			parts := strings.Split(s, "-")
			n1, _ := strconv.Atoi(parts[0])
			n2, _ := strconv.Atoi(parts[1])
			n1 = int(math.Abs(float64(n1)))
			n2 = int(math.Abs(float64(n2)))
			if n1 > n2 {
				n2 = n1 + 10
			} else if n2 > n1 {
				n1 = n2 - 10
				if n1 < 0 {
					n1 = 0
				}
			}
			for n := n1; n <= n2; n++ {
				sizes = append(sizes, fmt.Sprintf("%d", n))
			}
			continue
		}

		sizes = append(sizes, s)
	}
	argSkipSizes = sizes

	// No errors
	return nil
}

func printUsedArgs() {
	fmt.Println(strings.Repeat("-", 80))
	color.Cyan("%20s: %s", "Source path", argSourcePath)
	color.Cyan("%20s: %s", "URL", argEndpoint)
	color.Cyan("%20s: %s", "Method", argMethod)
	color.Cyan("%20s: %v", "Depth scan", argDepth)
	color.Cyan("%20s: %v", "Dir only", argDirOnly)
	color.Cyan("%20s: %d (ms)", "Delay", argDelay)
	color.Cyan("%20s: %d (s)", "Timeout", argTimeout)
	color.Cyan("%20s: %s", "Output", argOutput)
	color.Cyan("%20s: %v", "Ignore dir/files", argSkip)
	color.Cyan("%20s: %v", "Ignore extensions", argSkipExts)
	color.Cyan("%20s: %v", "Ignore by HTTP Code", argSkipCodes)
	color.Cyan("%20s: %v", "Ignore by size", argSkipSizes)
	color.Cyan("%20s: %v", "Mutation options", argBackups)
	color.Cyan("%20s: %v", "User-Agent", argUserAgent)
	color.Cyan("%20s: %v", "Cookie", argCookieString)
	color.Cyan("%20s: %v", "Headers", argHeaderString)
	fmt.Println(strings.Repeat("-", 80))
}

func normalizeArgSlice(arr []string) []string {
	s := strings.Join(arr, ",")

	// all to one separator
	s = strings.Replace(s, ";", ",", -1)
	s = strings.Replace(s, "/", ",", -1)
	s = strings.Replace(s, "|", ",", -1)

	// back to slice and items that added in as cli also now separated
	arr = strings.Split(s, ",")
	return arr
}
