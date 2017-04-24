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
    "flag"
    "strconv"
    "log"
    "fmt"
    "os"
    "io"
)

func main() {
    var port int
    flag.IntVar(&port, "port", 8765, "Web Server Address")
    flag.Parse()
    StartEndpoint(":" + strconv.Itoa(port))
}

// example upload to custom folder with sub/a.txt path
// curl --data-binary @"a.txt" http://server:8765/upload/custom/sub/a.txt
func uploadHandler(w http.ResponseWriter, r *http.Request) {

    //Split request into /<upload>/<folder></remainderPath>
    seg := strings.SplitN(r.URL.Path, "/", 4)

    //'' is seg[0]
    //'upload' is seg[1]
    folder := seg[2]
    path := seg[3] //contains leading "/"

    w.Write([]byte(fmt.Sprintf("Will save %s into folder '%s'\n", path, folder)))

    folders := getFolders()

    //should always be ok because handler is registered as /upload/<folderName>.
    folderCfg, ok := folders[folder]

    if !ok {
    panic("'" + folder + "' does not exist")
    }

    //TODO MS Windows convert separators in path.
    //TODO pre handler appears to clean /../../ paths by sending to 404. double check security!
    storagePath := folderCfg.RawPath + path

    parent := filepath.Dir(storagePath)

    err := os.MkdirAll(parent, 0777)
    if err != nil {
        panic(err)
    }

    file, err := os.Create(storagePath)
    if err != nil {
        panic(err)
    }

    n, err := io.Copy(file, r.Body)
    if err != nil {
        panic(err)
    }

    //Inform sender
    w.Write([]byte(fmt.Sprintf("%d bytes successfully received.\n", n)))
    //Log
    log.Printf("UPLOAD %s %s -> %s (%d bytes)\n", r.RemoteAddr, r.URL.Path, storagePath, n)
}

//
// StartEndpoint Adds another simple http server of the shared folders
//
func StartEndpoint(portString string){

    log.Printf("Starting simple http file server. \n")

    folders := getFolders()
    //
    // Essentially the following, but for each folder
    // http.Handle("/tmpfiles/", http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp"))))
    //
    for folder, cfg := range folders {

        id := folder
        if cfg.Label != "" {
            id = cfg.Label
        }
        path := filepath.Clean(cfg.Path())
        fs := http.FileServer(http.Dir(path))
        prefix := "/" + id + "/" //The last slash is absolutely required for sub paths to work.  weird.
        log.Printf("%s -> http://127.0.0.1%s%s\n", path, portString, prefix)
        h := DiscourageDownloadEverything(http.StripPrefix(prefix, fs))
        http.Handle(prefix, h)

        http.HandleFunc("/upload" + prefix, uploadHandler)
    }

    err := http.ListenAndServe(portString, nil)
    log.Panic(err)
}


//
// DiscourageDownloadEverything is a special handler to discourage mirroring everything.
// TODO
//  * This is Security through obscurity
//  * Active directory or single sign on. (AD SSO)
//
func DiscourageDownloadEverything(h http.Handler) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        log.Printf("DOWNLOAD %s %s\n", r.RemoteAddr, r.URL.Path)

        //No Directory Listings, note /dir gets redirected to /dir/
        if strings.HasSuffix(r.URL.Path, "/") {
            log.Printf("Slowing a directory listing, will return 404\n")
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
        log.Printf("Replacing 404 with 204\n")
        h.ResponseWriter.WriteHeader(204)
        return
    }
/*
    if 404 == code {
        log.Printf("Slowing a 404\n")
        time.Sleep(60 * time.Second)
    }
*/
    h.ResponseWriter.WriteHeader(code)
}
