/* For license and copyright information please see LEGAL file in repository */

package http

import (
	"strings"

	"../convert"
	"../protocol"
)

// Cookies parses and returns the Cookie headers.
// By related RFC we just support one Cookie in header.
// https://tools.ietf.org/html/rfc6265#section-5.4
func (h *header) Cookies() (cookies []Cookie) {
	var cookie = h.Get(HeaderKeyCookie)
	if len(cookie) == 0 {
		return
	}
	var index int
	cookies = make([]Cookie, 0, 8)
	var c Cookie
	for {
		index = strings.IndexByte(cookie, ';')
		if index == -1 {
			c.Unmarshal(cookie)
			cookies = append(cookies, c)
			return
		}
		c.Unmarshal(cookie[:index])
		cookies = append(cookies, c)

		cookie = cookie[index+2:]
	}
}

// MarshalSetCookies parses and set them to Cookie header.
func (h *header) MarshalSetCookies(cookies []Cookie) {
	// TODO::: make buffer by needed size.
	var b strings.Builder
	var ln = len(cookies)
	var i int
	for ; ; i++ {
		b.WriteString(cookies[i].Name)
		b.WriteByte('=')
		b.WriteString(cookies[i].Value)
		if i < ln {
			b.WriteString(SemiColonSpace)
		} else {
			break
		}
	}
	h.Set(HeaderKeyCookie, b.String())
}

// Cookie represents an HTTP cookie as sent in the Cookie header of an HTTP request.
// implement by https://tools.ietf.org/html/rfc6265#section-4.2
type Cookie struct {
	Name  string
	Value string
}

// CheckAndSanitize check if the cookie is in standard by RFC and try to fix them. It returns last error!
func (c *Cookie) CheckAndSanitize() (err protocol.Error) {
	c.Name, err = sanitizeCookieName(c.Name)
	c.Value, err = sanitizeCookieValue(c.Value)
	return
}

// Marshal returns the serialization of the cookie.
func (c *Cookie) Marshal() string {
	return c.Name + "=" + c.Value
}

// Unmarshal parse given cookie value to c and return!
func (c *Cookie) Unmarshal(cookie string) {
	var equalIndex = strings.IndexByte(cookie, '=')
	// First check no equal(=) sign or empty name or value
	if equalIndex < 1 || equalIndex == len(cookie)-1 {
		return
	}
	c.Name = cookie[:equalIndex]
	c.Value = cookie[equalIndex+1:]
}

func sanitizeCookieName(n string) (name string, err protocol.Error) {
	var ln = len(n)
	var buf = make([]byte, 0, ln)
	var b byte
	for i := 0; i < ln; i++ {
		b = n[i]
		if b == '\n' || b == '\r' {
			buf = append(buf, '-')
			err = ErrCookieBadName
		} else {
			buf = append(buf, b)
		}
	}
	name = convert.UnsafeByteSliceToString(buf)
	return
}

// https://tools.ietf.org/html/rfc6265#section-4.1.1
// cookie-value      = *cookie-octet / ( DQUOTE *cookie-octet DQUOTE )
// cookie-octet      = %x21 / %x23-2B / %x2D-3A / %x3C-5B / %x5D-7E
//           ; US-ASCII characters excluding CTLs,
//           ; whitespace, DQUOTE, comma, semicolon,
//           ; and backslash
// Don't check for ; due to Unmarshal will panic for bad cookie!!
func sanitizeCookieValue(v string) (value string, err protocol.Error) {
	var ln = len(v)
	var buf = make([]byte, 0, ln)
	var b byte
	for i := 0; i < ln; i++ {
		b = v[i]
		if 0x20 <= b && b < 0x7f && b != ' ' && b != '"' && b != ',' && b != '\\' {
			buf = append(buf, b)
		} else {
			err = ErrCookieBadValue
		}
	}
	value = convert.UnsafeByteSliceToString(buf)
	return
}