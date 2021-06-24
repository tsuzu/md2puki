package urlutil

import "net/url"

func EscapeURL(s string) string {
	u, err := url.Parse(s)

	if err != nil {
		return s
	}

	return u.String()
}
