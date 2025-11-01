package analytics_test

import (
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ee/api/analytics"
	"github.com/stretchr/testify/suite"
)

type BotsSuite struct {
	suite.Suite
}

func (s *BotsSuite) Test_IsBot_DetectsKnownBots() {
	botUserAgents := []string{
		"Googlebot/2.1 (+http://www.google.com/bot.html)",
		"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
		"Twitterbot/1.0",
		"LinkedInBot/1.0 (compatible; Mozilla/5.0; Apache-HttpClient +http://www.linkedin.com/)",
		"Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)",
		"WhatsApp/2.19.81 A",
		"TelegramBot (like TwitterBot)",
		"DiscordBot (https://discordapp.com)",
		"crawler",
		"spider",
		"bot",
		"scraper",
	}

	for _, userAgent := range botUserAgents {
		s.True(analytics.IsBot(userAgent), "Should detect '%s' as a bot", userAgent)
	}
}

func (s *BotsSuite) Test_IsBot_AllowsRealUsers() {
	realUserAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Android 11; Mobile; rv:89.0) Gecko/89.0 Firefox/89.0",
		"PostmanRuntime/7.28.0",
	}

	for _, userAgent := range realUserAgents {
		s.False(analytics.IsBot(userAgent), "Should not detect '%s' as a bot", userAgent)
	}
}

func (s *BotsSuite) Test_IsBot_CaseInsensitive() {
	testCases := []string{
		"GoogleBot/2.1",
		"GOOGLEBOT/2.1",
		"googlebot/2.1",
		"Bot test",
		"BOT TEST",
		"bot test",
	}

	for _, userAgent := range testCases {
		s.True(analytics.IsBot(userAgent), "Should detect '%s' as a bot (case insensitive)", userAgent)
	}
}

func TestBotsSuite(t *testing.T) {
	suite.Run(t, &BotsSuite{})
}
