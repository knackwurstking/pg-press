package cookies

import "slices"

type Cookies struct {
	list []*Cookie
}

func NewCookies() *Cookies {
	return &Cookies{}
}

func New() *Cookies {
	return NewCookies()
}

func (c *Cookies) Add(userAgent, value string) {
	c.list = append(c.list, &Cookie{
		UserAgent: userAgent,
		Value:     value,
	})
}

func (c *Cookies) Contains(userAgent, value string) bool {
	for _, cookie := range c.list {
		if cookie.UserAgent == userAgent && cookie.Value == value {
			return true
		}
	}

	return false
}

func (c *Cookies) Remove(userAgent, value string) {
	for i, cookie := range c.list {
		if cookie.UserAgent == userAgent && cookie.Value == value {
			c.list = slices.Delete(c.list, i, i+1)
		}
	}
}
