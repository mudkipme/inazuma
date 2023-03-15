package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/text/language"
)

func containsNonUTMQueryParameters(u *url.URL) bool {
	utmParameters := map[string]bool{
		"utm_source":   true,
		"utm_medium":   true,
		"utm_campaign": true,
		"utm_term":     true,
		"utm_content":  true,
	}

	queryParameters := u.Query()
	for key := range queryParameters {
		if _, isUTM := utmParameters[key]; !isUTM {
			return true
		}
	}
	return false
}

func hasBypassCacheCookie(r *http.Request) bool {
	if cookie, _ := r.Cookie(cookieToBypassCache); cookie != nil {
		return true
	}
	return false
}

func getCacheKeySuffix(r *http.Request) string {
	if !strings.HasPrefix(r.URL.Path, "/wiki/") {
		return ""
	}

	simplifiedTags := []string{"zh-hans", "zh-cn", "zh-sg", "zh-my"}
	traditionalTags := []string{"zh-hant", "zh-hk", "zh-tw", "zh-mo"}

	tags, qValues, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil {
		log.Printf("Failed to parse Accept-Language header: %v\n", err)
		return ""
	}

	highestQ := float32(-1)
	bestSuffix := ""
	for i, tag := range tags {
		if tag.Parent() == language.Chinese {
			if containsStringSlice(simplifiedTags, tag.String()) && qValues[i] > highestQ {
				highestQ = qValues[i]
				bestSuffix = "_zh-hans"
			} else if containsStringSlice(traditionalTags, tag.String()) && qValues[i] > highestQ {
				highestQ = qValues[i]
				bestSuffix = "_zh-hant"
			}
		}
	}

	return bestSuffix
}

func containsStringSlice(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
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
