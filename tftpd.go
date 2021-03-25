/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"

	//"github.com/vmware/gotftp"
	"github.com/pin/tftp"
)

/*
type Handler struct {
	Path string
}

func (h Handler) ReadFile(c gotftp.Conn, filename string) (gotftp.ReadCloser, error) {
	log.Printf("Request from %s to read %s", c.RemoteAddr(), filename)
	return os.OpenFile(path.Join(h.Path, filename), os.O_RDONLY, 0)
}

func (h Handler) WriteFile(c gotftp.Conn, filename string) (gotftp.WriteCloser, error) {
	return nil, fmt.Errorf("not allowed")
}

func TFTPd() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pwd = pwd + "/tftp"

	h := Handler{Path: pwd}
	err = gotftp.ListenAndServe(h)
	panic(err)
}*/

func readHandler(filename string, rf io.ReaderFrom) error {

	// get the requesting ip-address
	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
	//strip the port
	ip, _, _ := net.SplitHostPort(raddr.String())

	//get the object that correlates with the ip
	var address models.Address
	db.DB.Preload(clause.Associations).First(&address, "ip = ?", ip)

	//get the image info that correlates with the pool the ip is in
	var image models.Image
	db.DB.First(&image, "id = ?", address.Group.ImageID)

	//if the filename is mboot.efi, we hijack it and serve the mboot.efi file that is part of that specific image, this guarantees that you always get an mboot file that works for the build.
	if filename == "mboot.efi" {
		logrus.WithFields(logrus.Fields{
			ip: "requesting mboot.efi",
		}).Info("tftpd")
		filename = image.Path + "/MBOOT.EFI"
	} else if filename == "/boot.cfg" {
		//if the filename is boot.cfg, we serve the boot cfg that belongs to that build.
		logrus.WithFields(logrus.Fields{
			ip: "requesting boot.cfg",
		}).Info("tftpd")
		filename = image.Path + "/BOOT.CFG"
	} else {
		if _, err := os.Stat("tftp/" + filename); err == nil {
			filename = "tftp/" + filename
		} else {
			dir, file := path.Split(filename)
			upperfile := strings.ToUpper(string(file))
			filename = "tftp/" + dir + upperfile
		}
		spew.Dump(filename)
	}

	//chroot the rest

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	logrus.WithFields(logrus.Fields{
		"file":  filename,
		"bytes": n,
	}).Info("tftpd")
	//fmt.Printf("%s sent\n", filename)
	//fmt.Printf("%d bytes sent\n", n)
	return nil
}

func TFTPd() {
	s := tftp.NewServer(readHandler, nil)
	s.SetTimeout(5 * time.Second)  // optional
	err := s.ListenAndServe(":69") // blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}
}
