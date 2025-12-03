package uidev

import (
	"fmt"
	"maps"
	"strings"

	"github.com/a-h/templ"
)

type HxOptions struct {
	Method          string
	Href            templ.SafeURL
	Trigger         string
	Swap            string
	SwapOOB         string
	Preserve        string
	OnResponseError string
	OnBeforeRequest string
	OnAfterRequest  string
}

func (o *HxOptions) Attributes() templ.Attributes {
	a := templ.Attributes{}

	if o.Href != "" {
		method := "get"
		if o.Method != "" {
			method = o.Method
		}
		method = fmt.Sprintf("hx-%s", o.Method)
		a[method] = string(o.Href)
	}

	if o.Trigger != "" {
		a["hx-trigger"] = o.Trigger
	}

	if o.Swap != "" {
		a["hx-swap"] = o.Swap
	}

	if o.SwapOOB != "" {
		a["hx-swap-oob"] = o.SwapOOB
	}

	if o.Preserve != "" {
		a["hx-preserve"] = o.Preserve
	}

	if o.OnResponseError != "" {
		a["hx-on::response-error"] = o.OnResponseError
	}

	if o.OnBeforeRequest != "" {
		a["hx-on::before-request"] = o.OnBeforeRequest
	}

	if o.OnAfterRequest != "" {
		a["hx-on::after-request"] = o.OnAfterRequest
	}

	return a
}

// ButtonOptions defines the configuration options for the button component
type ButtonOptions struct {
	HxOptions
	Href     templ.SafeURL
	Anchor   bool
	Variant  ButtonVariant
	Color    Color
	Size     ButtonSize
	Icon     string
	Disabled bool
	Active   bool
	Classes  map[string]bool
	Styles   map[string]string
}

func (o *ButtonOptions) Attributes(classes map[string]bool) templ.Attributes {
	if classes == nil {
		classes = map[string]bool{}
	}

	a := templ.Attributes{}

	if o.Anchor {
		a["role"] = "button"
	}

	if o.Href != "" && o.HxOptions.Href == "" {
		a["href"] = string(o.Href)
	}

	if o.Disabled {
		a["disabled"] = ""
	}

	c := []string{}
	for k, v := range classes {
		if v {
			c = append(c, k)
		}
	}

	for k, v := range o.Classes {
		if v {
			c = append(c, k)
		}
	}

	if o.Variant != "" {
		c = append(c, string(o.Variant))
	}

	if o.Color != "" {
		c = append(c, string(o.Color))
	}

	if o.Size != "" {
		c = append(c, string(o.Size))
	}

	a["class"] = strings.Join(c, " ")

	maps.Copy(a, o.HxOptions.Attributes())

	return a
}
