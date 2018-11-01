package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

var appname = "findthese"
var version = "v0.1"

// flags
var argSourcePath string
var argEndpoint string
var argMethod = "HEAD"                                                              // assigned default value
var argOutput = "./findthese.report"                                                // assigned default value
var argDelay = 100                                                                  // assigned default value
var argSkip = []string{"jquery", "css", "img", "images", "i18n", "po"}              // assigned default value
var argSkipExts = []string{".png", ".jpeg", "jpg", "Gif", ".CSS", ".less", ".sass"} // assigned default value
var argDirOnly = false                                                              // assigned default value

// asterisk "*" replaced by filename
// if no asterisk found treat as suffix
var argBackups = []string{"~", ".swp", ".swo", ".tmp", ".TMP", ".lock", ".bkp", ".backup", ".bak", ".old", "_*", "~*"} // assigned default value

func main() {
	parseArgs()
	printUsedArgs()

	// Walk local source directory
	log.Printf("(START)")
	fmt.Println(strings.Repeat("-", 80)) // cleans \r
	if err := filepath.Walk(argSourcePath, localFileVisit); err != nil {
		fmt.Printf("ERR: Local directory: %v\n", err)
	}
	fmt.Println(strings.Repeat("-", 80)) // cleans \r
	log.Printf("(END)")

}

// callback
func localFileVisit(fpath string, f os.FileInfo, err error) error {
	fpath = strings.TrimPrefix(fpath, argSourcePath) // without local directory path
	depth := strings.Count(string(os.PathSeparator), fpath)

	if fpath == "" {
		return nil
	}

	//  skip file if allowed to scan only directories
	if argDirOnly && !f.IsDir() {
		return nil
	}

	// Skip predefined dirs
	if f.IsDir() {
		if inSlice(f.Name(), []string{".", "..", ".hg", ".git"}) {
			// fmt.Printf("-- SKIP ALWAYS [%s] --", f.Name())
			return filepath.SkipDir
		}
	}

	// Skip by name
	if inSlice(f.Name(), argSkip) {
		if f.IsDir() {
			return filepath.SkipDir // to skip whole tree
		}
		return nil // skip one item
	}

	// Skip by file extension
	if !f.IsDir() {
		ext := strings.ToLower(filepath.Ext(fpath))
		if inSlice(ext, argSkipExts) {
			// fmt.Printf("-- SKIP [%s] --", ext)
			return nil
		}
	}

	// generate mutations fpath list based on given fpath
	var fpaths []string
	fpaths = filePathMutations(fpath, argBackups)
	fpaths = append(fpaths, fpath) // keep original fpath too

	// Loop throw all fpath versions
	for _, fpath := range fpaths {
		fullURL := argEndpoint + fpath
		fname := filepath.Base(fpath)

		// Delay after basic checks and right before call
		time.Sleep(time.Duration(argDelay) * time.Millisecond)

		// Fetch
		resp, err := fetchURL(argMethod, fullURL)
		if err != nil {
			color.Red("ERR: %v", err)
			fmt.Println()
			continue
		}

		sCode := fmt.Sprintf("%d", resp.StatusCode)
		sLength := fmt.Sprintf("%d", resp.ContentLength)
		sMore := "" // add at the end of line
		switch {

		case sCode == "404" || sLength == "-1":
			// do not print out
			fmt.Printf("\r")
			fmt.Printf("-> %20s \tCODE:%s SIZE:%s ", fname, sCode, sLength)
			fmt.Printf(strings.Repeat(" ", 80)) // cleaning
			fmt.Printf("\r")
			continue

		case sCode == "200":
			sCode = color.GreenString(sCode)
			sMore += color.GreenString(fullURL)

		case sCode[:1] == "3": // 3xx codes
			sCode = color.CyanString(sCode)
			sMore += color.CyanString(fullURL)

		case sCode[:1] == "4": // 4xx codes
			sCode = color.YellowString(sCode)
			sMore += color.YellowString(fullURL)

		case sCode[:1] == "5": // 5xx codes
			sCode = color.RedString(sCode)
			sMore += color.RedString(fullURL)
		}

		fmt.Printf("depth=%d %20s | %-7s ", depth, fname, argMethod)
		fmt.Printf("CODE:%-4s SIZE:%-10s %-10s", sCode, sLength, sMore)
		fmt.Println()

	}

	return nil
}

// Fetches url content to dataTarget
func fetchURL(method, URL string) (*http.Response, error) {
	client := requestClient(URL)

	// Request
	req, _ := http.NewRequest(method, URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.75 Safari/537.36")

	// Make request
	resp, reqErr := client.Do(req)
	if reqErr != nil {
		log.Printf("ERROR: [FETCH] %s -- %v", URL, reqErr)
		return nil, reqErr
	}

	return resp, nil
}

// Common request http client for data fetch
func requestClient(URL string) *http.Client {
	u, _ := url.Parse(URL)

	tr := &http.Transport{}

	if u.Scheme == "https" {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client := &http.Client{
		Transport: tr,
		// Timeout:   time.Second * 5,
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrNoLocation
		// },
	}

	return client
}

// generate list of file mutations
// given argument can be single filename [file.txt]
// or path [path/to/file.txt]
func filePathMutations(fpath string, patterns []string) []string {
	var mutations []string

	fname := filepath.Base(fpath)
	for _, pattern := range patterns {
		smut := fname + pattern // as suffix

		// replace asterisk with fname
		if strings.Contains(pattern, "*") {
			smut = strings.Replace(pattern, "*", fname, 1)
		}

		mutations = append(mutations, smut)
	}

	// color.Red("MUT: %v", mutations)
	return mutations
}
