package utils

import "net/url"

// NormalizeURL нормализует URL, убирая фрагменты и нормализуя путь
func NormalizeURL(rawUrl string) (string, error) {
	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	parsed.Fragment = ""
	return parsed.String(), nil
}

// IsSameDomain проверяет, что два URL принадлежат одному домену
func IsSameDomain(url1, url2 string) bool {
	u1, err := url.Parse(url1)
	if err != nil {
		return false
	}

	u2, err := url.Parse(url2)
	if err != nil {
		return false
	}

	return u1.Host == u2.Host
}
