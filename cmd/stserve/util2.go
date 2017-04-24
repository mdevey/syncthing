// Copyright (C) 2015 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
    "fmt"
    "os"
    "crypto/tls"
    "path/filepath"

    "github.com/syncthing/syncthing/lib/osutil"
    "github.com/syncthing/syncthing/lib/protocol"
    "github.com/syncthing/syncthing/lib/config"
)

func getFolders() map[string]config.FolderConfiguration{
    myID := getMyID()
    path := defaultConfigDir() + "/config.xml"

    //Side step the file mutex in config.Load.
    path2 := path + ".stendpoint"
    osutil.Copy(path, path2);

    cfg, err := config.Load(path2, myID)
    if err != nil{
        fmt.Println(err)
        os.Exit(1)
    }

    return cfg.Folders()
}

func getMyID() protocol.DeviceID {
    dir := defaultConfigDir()
    certFile, keyFile := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil{
        fmt.Println(err)
        os.Exit(1)
    }
    myID := protocol.NewDeviceID(cert.Certificate[0])
    return myID
}
