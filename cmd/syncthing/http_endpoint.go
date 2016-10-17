// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"net/http"
	"path/filepath"
	"time"
	"strings"

	"github.com/syncthing/syncthing/lib/config"
)

//
//Add another simple http server of the shared folders
//
func Start_HTTP_endpoint(portString string, folders map[string]config.FolderConfiguration){
	l.Infoln("Starting simple http file server.")
	go func() {
		//
		// Essentially the following, but for each folder
		// http.Handle("/tmpfiles/", http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp"))))
		//
		for folder, cfg := range folders {

			id := folder
			if cfg.Label != "" {
				id = cfg.Label
			}
			path := filepath.Clean(cfg.Path());
			fs := http.FileServer(http.Dir(path))
			prefix := "/" + id + "/" //The last slash is absolutely required for sub paths to work.  weird.
			l.Infoln(path + " is available at http://127.0.0.1" + portString + prefix)
			h := DiscourageDownloadEverything(http.StripPrefix(prefix, fs))
			http.Handle(prefix, h)
		}

		http.ListenAndServe(portString, nil)
	}()
}


//
//Special Handler to discourage mirroring everything.
//TODO
// * This is Security through obsurity
// * Active directory or single sign on. (AD SSO)
// * logging
//
func DiscourageDownloadEverything(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//No Directory Listings, note /dir gets redirected to /dir/
		if strings.HasSuffix(r.URL.Path, "/") {
			l.Infoln("Slowing a directory listing, will return 404")
			time.Sleep(60 * time.Second)
			http.NotFound(w, r)
			return
		}

		//Other ideas
		//* Limit an IP to X concurrent requests.

		//* If file does not exist add extra wait time.
		slowWriter := &slow404{ ResponseWriter: w }

		//Slows things down even when a legit file is fetched.
		time.Sleep(100 * time.Millisecond)

		h.ServeHTTP(slowWriter, r)
	})
}

//
// slow404
//
type slow404 struct {
	http.ResponseWriter
}

func (h *slow404) WriteHeader(code int) {
	if 404 == code {
		l.Infoln("Slowing a 404")
		time.Sleep(60 * time.Second)
	}

	h.ResponseWriter.WriteHeader(code)
}
