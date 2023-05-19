package server

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/mudkipme/inazuma/config"
	"golang.org/x/text/language"
)

var (
	utmParameters = map[string]bool{
		"utm_source":   true,
		"utm_medium":   true,
		"utm_campaign": true,
		"utm_term":     true,
		"utm_content":  true,
	}
)

func containsNonUTMQueryParameters(u *url.URL) bool {
	queryParameters := u.Query()
	for key := range queryParameters {
		if _, isUTM := utmParameters[key]; !isUTM {
			return true
		}
	}
	return false
}

func hasBypassCacheCookie(r *http.Request, cookieToBypassCache []string) bool {
	for _, cookieName := range cookieToBypassCache {
		if cookie, _ := r.Cookie(cookieName); cookie != nil {
			return true
		}
	}
	return false
}

func getCacheKey(r *http.Request, cachePatterns []config.CachePattern) (cacheKey string, purge bool) {
	for _, pattern := range cachePatterns {
		re, err := regexp.Compile(pattern.Path)
		if err != nil {
			continue
		}
		matches := re.FindStringSubmatchIndex(r.URL.Path)
		if len(matches) == 0 {
			continue
		}

		matchReplacemnts := make(map[int]string)
		for _, rule := range pattern.ReplaceRules {
			if len(matches)/2 <= rule.Match {
				continue
			}

			tags, qValues, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
			highestQ := float32(-1)
			var replacement string

			for _, item := range rule.Replacements {
				if item.Default && replacement == "" {
					replacement = item.Replacement
					continue
				}

				for i, tag := range tags {
					if containsStringSlice(item.LanguageMatches, tag.String()) && qValues[i] > highestQ {
						replacement = item.Replacement
						highestQ = qValues[i]
					}
				}
			}

			if replacement != "" {
				matchReplacemnts[rule.Match] = replacement
			}
		}

		cacheKey = ""
		currentPos := 0
		for i := 2; i < len(matches); i += 2 {
			cacheKey += r.URL.Path[currentPos:matches[i]]
			if replacement, ok := matchReplacemnts[i/2]; ok {
				cacheKey += replacement
			} else {
				cacheKey += r.URL.Path[matches[i]:matches[i+1]]
			}
			currentPos = matches[i+1]
		}
		cacheKey += r.URL.Path[currentPos:]
		purge = pattern.Purge
		return
	}
	return
}

func containsStringSlice(slice []string, value string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
