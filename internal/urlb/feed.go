package urlb

import "github.com/a-h/templ"

func Feed() templ.SafeURL {
	return BuildURL("/feed")
}

func FeedList() templ.SafeURL {
	return BuildURL("/feed/list")
}
