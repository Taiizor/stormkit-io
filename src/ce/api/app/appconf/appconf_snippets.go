package appconf

import (
	"strings"
)

type SnippetInjection struct {
	HeadAppend  string `json:"headAppend,omitempty"`
	HeadPrepend string `json:"headPrepend,omitempty"`
	BodyAppend  string `json:"bodyAppend,omitempty"`
	BodyPrepend string `json:"bodyPrepend,omitempty"`
}

type SnippetFilters struct {
	RequestPath string
}

// IsEmpty returns true if all locations are empty.
func (sn *SnippetInjection) IsEmpty() bool {
	return sn.HeadAppend == "" && sn.HeadPrepend == "" && sn.BodyAppend == "" && sn.BodyPrepend == ""
}

// From reads the given snippet object, filters the disabled ones
// and joins the snippets.
func SnippetsHTML(sn Snippets, filters ...SnippetFilters) *SnippetInjection {
	if len(sn) == 0 {
		return nil
	}

	f := SnippetFilters{}

	if len(filters) > 0 {
		f = filters[0]
	}

	s := &SnippetInjection{}

	headAppend := []string{}
	headPrepend := []string{}
	bodyAppend := []string{}
	bodyPrepend := []string{}

	for _, snippet := range sn {
		if snippet.Rules != nil && snippet.Rules.PathCompiled != nil {
			if match, _ := snippet.Rules.PathCompiled.MatchString(f.RequestPath); !match {
				continue
			}
		}

		if snippet.Location == "head" {
			if snippet.Prepend {
				headPrepend = append(headPrepend, snippet.Content)
			} else {
				headAppend = append(headAppend, snippet.Content)
			}
		} else if snippet.Location == "body" {
			if snippet.Prepend {
				bodyPrepend = append(bodyPrepend, snippet.Content)
			} else {
				bodyAppend = append(bodyAppend, snippet.Content)
			}
		}
	}

	s.HeadAppend = strings.Join(headAppend, "")
	s.HeadPrepend = strings.Join(headPrepend, "")
	s.BodyAppend = strings.Join(bodyAppend, "")
	s.BodyPrepend = strings.Join(bodyPrepend, "")

	return s
}
