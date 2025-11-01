package snippetshandlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appcache"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appconf"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

const LOCATION_HEAD = "head"
const LOCATION_BODY = "body"

var errorsMap = map[string]string{
	"invalid-location":    "Location must be either 'head' or 'body'.",
	"invalid-content":     "Snippet content is a required field.",
	"invalid-title":       "Snippet title is a required field.",
	"invalid-path-regexp": "Snippet path must be a valid regular expression.",
	"no-item":             "Nothing to update.",
}

// CalculateResetDomains will calculate the domains to be reset when
// a snippet gets updated.
//
// - If the returned string array is nil, all domains should be updated.
// - If it's empty an empty string array, no domain should be updated.
// - Otherwise only keys returned should be updated.
func CalculateResetDomains(appDisplayName string, snippets []*buildconf.Snippet) []string {
	reset := map[string]bool{}

	for _, snippet := range snippets {
		if snippet.Rules == nil || len(snippet.Rules.Hosts) == 0 {
			return nil
		}

		for _, host := range snippet.Rules.Hosts {
			// It's enough that one rule contains no host configuration,
			// we'll have to reset cache for all domains anyways.
			if len(host) == 0 {
				return nil
			}

			if strings.EqualFold(host, "*.dev") {
				reset[appcache.DevDomainCacheKey(appDisplayName)] = true
			} else {
				reset[host] = true
			}

		}
	}

	if len(reset) == 0 {
		return []string{}
	}

	slice := make([]string, len(reset))
	count := 0

	for k := range reset {
		slice[count] = k
		count++
	}

	return slice
}

func validateDomains(snippets []*buildconf.Snippet, envID types.ID) error {
	hosts := []string{}

	for _, snippet := range snippets {
		if snippet.Rules != nil && len(snippet.Rules.Hosts) > 0 {
			for _, host := range snippet.Rules.Hosts {
				if host != "*.dev" {
					hosts = append(hosts, host)
				}
			}
		}
	}

	if len(hosts) == 0 {
		return nil
	}

	missingHosts, err := buildconf.SnippetsStore().MissingHosts(context.Background(), hosts, envID)

	if err != nil {
		slog.Errorf("error while fetching missing hosts: %s", err.Error())
		return err
	}

	if len(missingHosts) == 0 {
		return nil
	}

	return fmt.Errorf("invalid or missing domain name(s): %s", strings.Join(missingHosts, ", "))
}

func validateSnippet(snippet *buildconf.Snippet, location string) error {
	if location != LOCATION_HEAD && location != LOCATION_BODY {
		return errors.New("invalid-location")
	}

	snippet.Title = strings.TrimSpace(snippet.Title)
	snippet.Content = strings.TrimSpace(snippet.Content)

	if snippet.Title == "" {
		return errors.New("invalid-title")
	}

	if snippet.Content == "" {
		return errors.New("invalid-content")
	}

	if snippet.Rules != nil && snippet.Rules.Path != "" {
		_, err := regexp2.Compile(snippet.Rules.Path, regexp2.None)

		if err != nil {
			return errors.New("invalid-path-regexp")
		}
	}

	return nil
}

// badRequest is a short-hand function to return a bad request response.
func badRequest(err error) *shttp.Response {
	msg := errorsMap[err.Error()]

	if msg == "" {
		msg = err.Error()
		msg = strings.ToUpper(msg[0:1]) + msg[1:]
	}

	return &shttp.Response{
		Status: http.StatusBadRequest,
		Data: map[string]string{
			"error": msg,
		},
	}
}

// normalizeRules will perform the following actions:
// - Remove all development endpoints in favor of *.dev
func normalizeRules(rules *buildconf.SnippetRule) {
	if rules == nil {
		return
	}

	hosts := []string{}
	added := false

	for _, host := range rules.Hosts {
		if appconf.IsStormkitDev(host) {
			if !added {
				hosts = append(hosts, "*.dev")
				added = true
			}
		} else {
			hosts = append(hosts, strings.ToLower(host))
		}
	}

	rules.Hosts = hosts
}
