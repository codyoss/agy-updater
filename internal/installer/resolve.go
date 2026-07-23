package installer

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// fetchWithRetry fetches a URL with GET, retrying up to 3 times on transient errors.
func fetchWithRetry(urlStr string) ([]byte, error) {
	var body []byte
	var err error
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
		var resp *http.Response
		resp, err = http.Get(urlStr)
		if err != nil {
			continue
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}
		return body, nil
	}
	return nil, err
}

// ResolveMainBundle fetches the download page and finds all JS bundle URLs.
func ResolveMainBundle(downloadPage string) (string, error) {
	htmlBytes, err := fetchWithRetry(downloadPage)
	if err != nil {
		return "", fmt.Errorf("failed to download download page %s: %w", downloadPage, err)
	}
	htmlStr := string(htmlBytes)

	// Match all JS scripts and asset files referenced in HTML
	reAny := regexp.MustCompile(`(?:src|href)=["']([^"']+\.js)["']`)
	matches := reAny.FindAllStringSubmatch(htmlStr, -1)

	baseURL, err := url.Parse(downloadPage)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL %s: %w", downloadPage, err)
	}

	var urls []string
	visited := make(map[string]bool)

	var addURL func(u string)
	addURL = func(u string) {
		if visited[u] {
			return
		}
		visited[u] = true
		urls = append(urls, u)

		// Fetch script to check for imported JS bundles or direct URLs
		jsBytes, err := fetchWithRetry(u)
		if err != nil {
			return
		}
		jsStr := string(jsBytes)

		// Find imported JS bundles: import ... from "./..." or import(...)
		reImport := regexp.MustCompile(`from\s*["']([^"']+\.js)["']|import\(["']([^"']+\.js)["']\)`)
		impMatches := reImport.FindAllStringSubmatch(jsStr, -1)
		currBase, err := url.Parse(u)
		if err != nil {
			return
		}
		for _, im := range impMatches {
			target := im[1]
			if target == "" {
				target = im[2]
			}
			refURL, err := url.Parse(target)
			if err == nil {
				resolved := currBase.ResolveReference(refURL).String()
				addURL(resolved)
			}
		}
	}

	for _, m := range matches {
		refURL, err := url.Parse(m[1])
		if err == nil {
			addURL(baseURL.ResolveReference(refURL).String())
		}
	}

	// Include the HTML page itself in case URLs are embedded directly in data attributes or inline scripts
	urls = append(urls, downloadPage)

	return strings.Join(urls, "|"), nil
}

// ResolveDownload resolves the version and download URL for a product and platform.
func ResolveDownload(jsBundleURLs, product, platform string) (string, string, error) {
	urlsToFetch := strings.Split(jsBundleURLs, "|")

	var marker, nextMarker, label string
	var filenamePatterns []string

	switch product {
	case "desktop":
		marker = `id:"antigravity-2"`
		nextMarker = `id:"antigravity-cli"`
		filenamePatterns = []string{`Antigravity\.tar\.gz`}
		label = "Antigravity 2.0"
	case "ide":
		marker = `id:"antigravity-ide"`
		nextMarker = `id:"antigravity-sdk"`
		filenamePatterns = []string{
			`Antigravity%20IDE\.tar\.gz`,
			`Antigravity\+IDE\.tar\.gz`,
			`Antigravity IDE\.tar\.gz`,
		}
		label = "Antigravity IDE"
	default:
		return "", "", fmt.Errorf("unknown product: %s", product)
	}

	for _, bundleURL := range urlsToFetch {
		jsBytes, err := fetchWithRetry(bundleURL)
		if err != nil {
			continue
		}

		jsStr := html.UnescapeString(string(jsBytes))
		// Normalize escaped slashes sometimes found in JS string literals
		jsStr = strings.ReplaceAll(jsStr, `\/`, `/`)

		var sections []string
		startIdx := strings.Index(jsStr, marker)
		if startIdx != -1 {
			endIdx := strings.Index(jsStr[startIdx:], nextMarker)
			if endIdx != -1 {
				sections = append(sections, jsStr[startIdx:startIdx+endIdx])
			} else {
				sections = append(sections, jsStr[startIdx:])
			}
		}
		sections = append(sections, jsStr)

		// Search each section for matches
		for _, section := range sections {
			for _, filePat := range filenamePatterns {
				patternStr := fmt.Sprintf(`https?://[^"'\s<>)']*/%s/%s`, regexp.QuoteMeta(platform), filePat)
				re := regexp.MustCompile(patternStr)
				matches := re.FindAllString(section, -1)
				if len(matches) > 0 {
					urlStr := matches[len(matches)-1]
					ver := versionFromURL(urlStr)
					return ver, urlStr, nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("could not find official %s tarball for %s in Google download bundle", label, platform)
}

var versionRegexes = []*regexp.Regexp{
	regexp.MustCompile(`/antigravity-hub/([^/]+)/`),
	regexp.MustCompile(`/stable/([^/]+)/`),
	regexp.MustCompile(`/(\d+\.\d+\.\d+(?:-[^/]+)?)/`),
}

func versionFromURL(urlStr string) string {
	decoded, err := url.QueryUnescape(urlStr)
	if err != nil {
		decoded = urlStr // fallback to original
	}
	for _, re := range versionRegexes {
		m := re.FindStringSubmatch(decoded)
		if len(m) > 1 {
			parts := strings.Split(m[1], "-")
			return parts[0]
		}
	}
	return "unknown"
}
