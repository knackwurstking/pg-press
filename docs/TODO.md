Use this:
`c.Response().Header().Set("HX-Trigger", "pageLoaded")`

Instead of:

```
hx-on:htmx:after-request="
	if(event.detail.successful) {
		window.dispatchEvent(new Event('visibilitychange'));
	}
"
```
