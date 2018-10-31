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
var argMethod = "HEAD"                                    // assigned default value
var argOutput = "./findthese.report"                      // assigned default value
var argDelay = 250                                        // assigned default value
var argSkipExts = []string{".png", ".jpeg", "jpg", "Gif"} // assigned default value

func main() {
	parseArgs()
	printUsedArgs()

	// Walk local source directory
	if err := filepath.Walk(argSourcePath, localFileVisit); err != nil {
		fmt.Printf("ERR: Local directory: %v\n", err)
	}

	log.Printf("FINISH")

}

// callback
func localFileVisit(fpath string, f os.FileInfo, err error) error {
	fpath = strings.TrimPrefix(fpath, argSourcePath) // without local directory path
	fullURL := argEndpoint + fpath

	defer fmt.Println()

	if fpath == "" {
		// docroot show full url
		fmt.Printf("== %20s | %-7s %s", fpath, argMethod, fullURL)
		return nil
	}

	fmt.Printf("== %20s | %-7s ", fpath, argMethod)

	// Skip predefined dirs
	if f.IsDir() {
		if inSlice(f.Name(), []string{".hg", ".git"}) {
			fmt.Printf("-- SKIP ALWAYS [%s] --", f.Name())
			return filepath.SkipDir

		}
	}

	// Skip by file extension
	if !f.IsDir() {
		ext := strings.ToLower(filepath.Ext(fpath))
		if inSlice(ext, argSkipExts) {
			fmt.Printf("-- SKIP [%s] --", ext)
			return nil
		}
	}

	// Delay after basic checks and right before call
	time.Sleep(time.Duration(argDelay) * time.Millisecond)

	// Fetch
	resp, err := fetchURL(argMethod, fullURL)
	if err != nil {
		color.Red("ERR: %v", err)
		return nil // do not stop walk
	}

	sCode := fmt.Sprintf("%d", resp.StatusCode)
	sLength := fmt.Sprintf("%d", resp.ContentLength)
	sMore := "" // add at the end of line
	switch {
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
	fmt.Printf("CODE:%-4s SIZE:%-10s %-10s", sCode, sLength, sMore)

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
