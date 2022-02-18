package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// URL represents the url of a wallet or account
type URL struct {
	ProtocolScheme string
	Path           string
}

// String implements the stringer interface.
func (u URL) String() string {
	if u.ProtocolScheme != "" {
		return fmt.Sprintf("%s://%s", u.ProtocolScheme, u.Path)
	}

	return u.Path
}

// MarshalJSON implements the json.Marshaller interface.
func (u URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON parses url.
func (u *URL) UnmarshalJSON(input []byte) error {
	var txt string

	err := json.Unmarshal(input, txt)
	if err != nil {
		return err
	}

	url, err := parseURL(txt)
	if err != nil {
		return err
	}

	u.ProtocolScheme = url.ProtocolScheme
	u.Path = url.Path

	return nil
}

func (u URL) Cmp(url URL) int {
	if u.ProtocolScheme == url.ProtocolScheme {
		return strings.Compare(u.Path, url.Path)
	}

	return strings.Compare(u.ProtocolScheme, url.ProtocolScheme)
}

// parseURL converts url string into the URL specific structure.
func parseURL(url string) (URL, error) {
	splits := strings.Split(url, "://")

	if len(splits) != 2 || splits[0] == "" {
		return URL{}, errors.New("protocol scheme missing")
	}

	return URL{
		ProtocolScheme: splits[0],
		Path:           splits[1],
	}, nil
}
