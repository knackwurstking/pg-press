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

func (c *Cookies) Get(value string) *Cookie {
	for _, cookie := range c.list {
		if cookie.Value == value {
			return cookie
		}
	}

	return nil
}

func (c *Cookies) Add(cookie *Cookie) {
	c.list = append(c.list, cookie)
}

func (c *Cookies) Contains(value string) bool {
	for _, cookie := range c.list {
		if cookie.Value == value {
			return true
		}
	}

	return false
}

func (c *Cookies) Remove(value string) {
	for i, cookie := range c.list {
		if cookie.Value == value {
			c.list = slices.Delete(c.list, i, i+1)
		}
	}
}
