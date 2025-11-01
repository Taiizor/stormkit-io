package analytics

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// These values are extracted from our analytics database
var BotList = []string{
	"amazon-kendra",
	"Apache-HttpClient/",
	"Atlassian",
	"Baiduspider",
	"Spider",
	"Baidu",
	"crawler",
	"dcrawl",
	"Expanse",
	"facebookexternalhit",
	"GoodLinks",
	"Google",
	"Go-http-client",
	"Grammarly",
	"quic-go-HTTP",
	"github",
	"gitlab",
	"Honeybadger Uptime Check",
	"LinkValidator",
	"lychee", // https://github.com/lycheeverse/lychee
	"Lenns.io",
	"Java/",
	"ManicTime",
	"Mozlila", // https://trunc.org/learning/the-mozlila-user-agent-bot
	"msray-plus",
	"NetworkingExtension",
	"Pulsetic.com",
	"Pandalytics",
	"python-",
	"Python/",
	"SEOlizer",
	"search.marginalia.nu",
	"Scrapy",
	"Sogou web spider",
	"Tines",
	"upptime.js.org",
	"Xpanse",
	"WhatsApp",
}

func init() {
	for i := range BotList {
		BotList[i] = strings.ToLower(BotList[i])
	}
}

// Additional patterns for more sophisticated bot detection
var botPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(bot|crawler|spider|scraper|fetch|monitor|check|test)\b`),
	regexp.MustCompile(`(?i)\b(curl|wget|http|client|java|python|go-http|ruby|php)\b`),
	regexp.MustCompile(`(?i)\b(headless|phantom|selenium|playwright)\b`),
	regexp.MustCompile(`(?i)\b(uptime|monitor|ping|health|status)\b`),
}

// Suspicious user agent patterns
func hasSuspiciousPatterns(userAgent string) bool {
	// Too short user agents are often bots
	if len(userAgent) < 10 {
		return true
	}

	// Missing common browser indicators
	if !strings.Contains(userAgent, "mozilla") &&
		!strings.Contains(userAgent, "webkit") &&
		!strings.Contains(userAgent, "gecko") {
		return true
	}

	return false
}

func IsBot(userAgent string) bool {
	if userAgent == "" {
		return true
	}

	if !hasSuspiciousPatterns(userAgent) {
		return false
	}

	// Check against regex patterns
	for _, pattern := range botPatterns {
		if pattern.MatchString(userAgent) {
			return true
		}
	}

	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "bot") {
		return true
	}

	for _, bot := range BotList {
		if strings.Contains(userAgent, bot) {
			return true
		}
	}

	return false
}

func IsUtf8(str string) bool {
	data := []byte(str)

	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)

		if r == utf8.RuneError && size == 1 {
			return false
		}

		data = data[size:]
	}

	return true
}
