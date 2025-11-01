package buildconf

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

type SnippetRule struct {
	Hosts        []string `json:"hosts,omitempty"`
	Path         string   `json:"path,omitempty"` // Accepts POSIX Regexp
	PathCompiled *regexp2.Regexp
}

// Scan implements the Scanner interface.
func (sr *SnippetRule) Scan(value any) error {
	if value != nil {
		return json.Unmarshal(value.([]byte), &sr)
	}

	return nil
}

// Value implements the Sql Driver interface.
func (sr *SnippetRule) Value() (driver.Value, error) {
	if sr == nil {
		return nil, nil
	}

	return json.Marshal(sr)
}

// Snippet represents a snippet.
type Snippet struct {
	// ID represents the internal id of the snippet. This is not a real auto_increment ID.
	// There is an internal counter stored alongside the Snippets object. Since we use
	// a `jsonb` column in the database, we cannot use the traditional IDs. We need these
	// IDs to facilitate REST operations.
	ID types.ID `json:"id,string"`

	// AppID contains the snippet application id.
	AppID types.ID `json:"appId,omitempty"`

	// EnvID contains the snippet environment id.
	EnvID types.ID `json:"envId,omitempty"`

	// Enabled specifies whether the snippet is enabled or not.
	Enabled bool `json:"enabled"`

	// Prepend specifies whether the snippet is prepended to the parent.
	// If this value is true, it will be inserted as the first child,
	// otherwise appended as the last child.
	Prepend bool `json:"prepend"`

	// Content is the snippet content.
	Content string `json:"content"`

	// Where to insert this snippet: head | body.
	Location string `json:"location"`

	// Rules contains the name of the domain that this snippet should be used.
	Rules *SnippetRule `json:"rules,omitempty"`

	// Title is the snippet title. It's used only by the Stormkit UI to provide
	// a meaningful description for the user.
	Title string `json:"title"`
}

// Snippets to be injected into the document.
type Snippets struct {
	Head   []Snippet `json:"head"`
	Body   []Snippet `json:"body"`
	LastID int64     `json:"lastId,omitempty"`
}

// Scan implements the Scanner interface.
func (s *Snippets) Scan(value any) error {
	if value != nil {
		if b, ok := value.([]byte); ok {
			json.Unmarshal(b, s)
		}
	}

	return nil
}

// Value implements the Sql Driver interface.
func (s *Snippets) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}

	if len(s.Body) == 0 && len(s.Head) == 0 {
		return nil, nil
	}

	return json.Marshal(s)
}

func (s *Snippet) ContentHash() string {
	h := sha256.New()
	h.Write([]byte(s.Content))
	return fmt.Sprintf("%x", h.Sum(nil))
}
