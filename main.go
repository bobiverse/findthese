package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
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
var argUserAgent = "random"                                                         // assigned default value
var argReportPath = "./findthese.report"                                            // assigned default value
var argDelay = 150                                                                  // assigned default value
var argTimeout = 10                                                                 // assigned default value
var argDepth = 0                                                                    // assigned default value
var argSkip = []string{"jquery", "css", "img", "images", "i18n", "po"}              // assigned default value
var argSkipExts = []string{".png", ".jpeg", "jpg", "Gif", ".CSS", ".less", ".sass"} // assigned default value
var argSkipCodes = []string{"404"}                                                  // assigned default value
var argSkipSizes = []string{}                                                       // assigned default value
var argSkipContent string                                                           // assigned default value
var argDirOnly = false                                                              // assigned default value
var argCookieString = ""                                                            // assigned default value
var argHeaderString = ""                                                            // assigned default value

// Parse `argHeaderString` and fills this map
var mHeaders = map[string]string{}

// asterisk "*" replaced by filename
// if no asterisk found treat as suffix
var argBackups = []string{"~", ".swp", ".swo", ".tmp", ".dmp", ".bkp", ".backup", ".bak", ".zip", ".tar", ".old", "_*", "~*"} // assigned default value

// Walk mode. Before real check/fetch count ETA
const walkModeCount = 0
const walkModeProcess = 1

var walkMode = walkModeCount
var dirItemCount = 0
var totalScanCount = 0

func main() {
	parseArgs()

	// TODO: Count items in source path folder and calc ~ETA
	walkMode = walkModeCount
	filepath.Walk(argSourcePath, localFileVisit)
	// durETA := time.Duration(totalScanCount*(argDelay+200)) * time.Millisecond
	printUsedArgs()

	// Setup logging
	defer LogSetupAndDestruct(argReportPath)()

	// Walk local source directory
	log.Printf("(START) -- (%d items + %d mutations)", dirItemCount, totalScanCount)
	walkMode = walkModeProcess
	fmt.Println(strings.Repeat("-", 80))
	if err := filepath.Walk(argSourcePath, localFileVisit); err != nil {
		fmt.Printf("ERR: Local directory: %v\n", err)
	}
	fmt.Println("\n" + strings.Repeat("-", 80))
	log.Printf("(END)")

}

// Last line length to know how much to clean
var lastLineLength int // cleaning current line with previous line length

// callback
func localFileVisit(fpath string, f os.FileInfo, err error) error {
	fpath = strings.TrimPrefix(fpath, argSourcePath) // without local directory path
	depth := strings.Count(fpath, "/") + 1

	if fpath == "" {
		return nil
	}

	//  skip file if allowed to scan only directories
	if argDirOnly && !f.IsDir() {
		return nil
	}

	// Skip predefined dirs
	if f.IsDir() {

		// Skip by allowed depth
		if argDepth > 0 && depth > argDepth {
			return filepath.SkipDir
		}

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

	// counting mode
	if walkMode == walkModeCount {
		dirItemCount++
		totalScanCount += len(argBackups) - 1
		return nil
	}

	// generate mutations fpath list based on given fpath
	var fpaths []string
	fpaths = filePathMutations(fpath, argBackups)

	// Loop throw all fpath versions
	for _, fpath := range fpaths {
		fullURL := argEndpoint + fpath
		// fname := filepath.Base(fpath)

		// Delay after basic checks and right before call
		if argDelay > 0 {
			time.Sleep(time.Duration(argDelay) * time.Millisecond)
		}

		// Fetch
		resp, err := fetchURL(argMethod, fullURL)
		if err != nil {
			color.Red("ERR: %v", err)
			fmt.Println()
			continue
		}

		sCode := fmt.Sprintf("%d", resp.StatusCode)

		// try to read real body length if empty
		var buf []byte
		buf, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.ContentLength <= 0 {
			resp.ContentLength = int64(len(buf))
		}
		sLength := fmt.Sprintf("%d", resp.ContentLength)

		// Check for "skip" rules
		isSkipable := inSlice(sCode, argSkipCodes)

		// by size
		isSkipable = isSkipable || inSlice(sLength, argSkipSizes)

		// Skip content for specifix methods
		if !isSkipable && argMethod != "HEAD" {
			// by content
			if argSkipContent != "" {
				isSkipable = bytes.Contains(buf, []byte(argSkipContent))
			}
		}

		fmt.Printf("\r")
		fmt.Printf(strings.Repeat(" ", lastLineLength)) // cleaning
		fmt.Printf("\r")

		sMore := "" // add at the end of line
		switch {

		case isSkipable:
			sLine := fmt.Sprintf("-> %s%s \tCODE:%s", color.MagentaString(argEndpoint), fpath, sCode)

			if argMethod != "HEAD" {
				sLine += fmt.Sprintf("SIZE:%s ", sLength)
			}

			lastLineLength = len(sLine)
			fmt.Printf(sLine)
			continue

		case sCode == "200":
			sCode = color.HiGreenString(sCode)
			sMore += color.GreenString(fullURL)

		case sCode[:1] == "3": // 3xx codes
			sCode = color.CyanString(sCode)
			sMore += color.CyanString(fullURL)

		case sCode[:1] == "4": // 4xx codes
			sCode = color.RedString(sCode)
			sMore += color.RedString(fullURL)

		case sCode[:1] == "5": // 5xx codes
			sCode = color.BlueString(sCode)
			sMore += color.BlueString(fullURL)
		}

		// fmt.Printf("\r")

		msg := fmt.Sprintf("%s ", argMethod)
		msg += fmt.Sprintf("CODE:%-4s ", sCode)
		if argMethod != "HEAD" {
			msg += fmt.Sprintf("SIZE:%-10s ", sLength)
		}
		msg += sMore

		// color.Red("%d < %d", len(msg), cleanupLen)

		log.Println(msg)
	}

	return nil
}

// Fetches url content to dataTarget
func fetchURL(method, URL string) (*http.Response, error) {
	client := requestClient(URL)

	// Request
	req, _ := http.NewRequest(method, URL, nil)

	// User-Agent
	req.Header.Set("User-Agent", argUserAgent)

	// Cookies string
	req.Header.Set("Cookie", argCookieString)

	// Custom headers
	// Can override previously set headers
	for hKey, hVal := range mHeaders {
		req.Header.Set(hKey, hVal)
	}

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
		Timeout:   time.Duration(argTimeout) * time.Second,
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrNoLocation
		// },
	}

	return client
}
