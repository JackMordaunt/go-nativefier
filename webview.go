package main

import (
	"runtime"

	"github.com/zserge/webview"
)

// NewWebview initialises a webview.
func NewWebview(title, url string, w, h int, resize, debug bool) webview.WebView {
	wv := webview.New(webview.Settings{
		Title:     title,
		URL:       url,
		Width:     w,
		Height:    h,
		Resizable: resize,
		Debug:     debug,
	})
	if runtime.GOOS == "windows" && debug {
		wv.Dispatch(func() {
			wv.Eval(loadFirebug())
		})
	}
	return wv
}

func loadFirebug() string {
	return `
function() {
	var script = document.createElement("script");
	script.type = "text/javascript";
	script.src = "https://getfirebug.com/firebug-lite.js";
	var head = document.head || document.getElementsByTagName('head')[0];
	head.appendChild(script);
}();
`
}
