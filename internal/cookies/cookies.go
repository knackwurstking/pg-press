package cookies

import "slices"

type Cookies []*Cookie

func NewCookies() Cookies {
	return make(Cookies, 0)
}

func New() Cookies {
	return NewCookies()
}

func (c Cookies) Add(userAgent, value string) {
	c = append(c, &Cookie{
		UserAgent: userAgent,
		Value:     value,
	})
}

func (c Cookies) Contains(userAgent, value string) bool {
	for _, cookie := range c {
		if cookie.UserAgent == userAgent && cookie.Value == value {
			return true
		}
	}

	return false
}

func (c Cookies) Remove(userAgent, value string) {
	for i, cookie := range c {
		if cookie.UserAgent == userAgent && cookie.Value == value {
			c = slices.Delete(c, i, i+1)
		}
	}
}
