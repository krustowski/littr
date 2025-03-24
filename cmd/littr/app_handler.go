package main

import (
	"os"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/config"
)

// appHandler holds the pointer to the very main FE app handler.
var appHandler = &app.Handler{
	Name:                    "littr nanoblogger",
	ShortName:               "littr",
	Title:                   "littr nanoblogger",
	Description:             "A simple nanoblogging platform",
	Author:                  "krusty",
	Domain:                  config.ServerUrl,
	BackgroundColor:         "#000000",
	ThemeColor:              "#000000",
	LoadingLabel:            "loading...",
	WasmContentLengthHeader: "X-Uncompressed-Content-Length",
	Lang:                    "en",
	Keywords: []string{
		"blog",
		"blogging",
		"board",
		"go-app",
		"microblog",
		"microblogging",
		"nanoblog",
		"nanoblogging",
		"platform",
		"social network",
	},
	Icon: app.Icon{
		Maskable: "/web/android-chrome-192x192.png",
		Default:  "/web/android-chrome-192x192.png",
		SVG:      "/web/android-chrome-512x512.svg",
		Large:    "/web/android-chrome-512x512.png",
		//AppleTouch: "/web/apple-touch-icon.png",
	},
	Image: "/web/android-chrome-512x512.png",
	// Ensure the default light theme is dark.
	Body: func() app.HTMLBody {
		return app.Body().Class("")
	},
	Version: os.Getenv("APP_VERSION") + "-" + time.Now().Format("2006-01-02_15:04:05"),
	// Environment constants to be transferred to the app context.
	Env: map[string]string{
		"APP_ENVIRONMENT":      os.Getenv("APP_ENVIRONMENT"),
		"APP_URL_MAIN":         os.Getenv("APP_URL_MAIN"),
		"APP_VERSION":          os.Getenv("APP_VERSION"),
		"REGISTRATION_ENABLED": os.Getenv("REGISTRATION_ENABLED"),
		"VAPID_PUB_KEY":        os.Getenv("VAPID_PUB_KEY"),
	},

	Preconnect: []string{
		//"https://cdn.vxn.dev/",
	},
	// Web fonts.
	Fonts: []string{
		"https://cdn.vxn.dev/css/material-symbols-outlined.woff2",
		//"https://cdn.jsdelivr.net/npm/beercss@3.5.0/dist/cdn/material-symbols-outlined.woff2",
	},
	// CSS styles files.
	Styles: []string{
		//"https://cdn.vxn.dev/css/beercss-3.7.0-custom.min.css",
		"https://cdn.vxn.dev/css/beercss-3.9.7-custom.min.css",
		"https://cdn.vxn.dev/css/sortable.min.css",
		"/web/littr.css",
	},
	// JS script files.
	Scripts: []string{
		"https://cdn.vxn.dev/js/beer-3.9.7.min.js",
		//"https://cdn.jsdelivr.net/npm/beercss@3.7.0/dist/cdn/beer.min.js",
		"https://cdn.vxn.dev/js/material-dynamic-colors.min.js",
		//"https://cdn.jsdelivr.net/npm/material-dynamic-colors@1.1.2/dist/cdn/material-dynamic-colors.min.js",
		"https://cdn.vxn.dev/js/sortable.min.js",
		"https://cdn.vxn.dev/js/eventsource.min.js",
		"/web/littr.js",
		//"https://cdn.vxn.dev/js/littr.js",
	},
	ServiceWorkerTemplate: config.EnchartedSW,
}
