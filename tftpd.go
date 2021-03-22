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
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
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
	spew.dump(filename)

	// get the requesting ip-address
	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
	//strip the port
	ip, _, _ := net.SplitHostPort(raddr.String())

	//get the object that correlates with the ip
	var address models.Address
	db.DB.Preload(clause.Associations).First(&address, "ip = ?", ip)
	//spew.Dump(address)

	//get the image info that correlates with the pool the ip is in
	var image models.Image
	db.DB.First(&image, "id = ?", address.Group.ImageID)
	//spew.Dump(image)

	//if the filename is mboot.efi, we hijack it and serve the mboot.efi file that is part of that specific image, this guarantees that you always get an mboot file that works for the build.
	if filename == "mboot.efi" {
		fmt.Println("mboot.efi requested!")
		filename = image.Path + "/MBOOT.EFI"
		//spew.Dump(filename)
	}

	//if the filename is boot.cfg, we serve the boot cfg that belongs to that build.
	if filename == "boot.cfg" {
		fmt.Println("boot.cfg requested!")
		filename = image.Path + "/BOOT.CFG"
		spew.Dump(filename)
	}

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
	fmt.Printf("%s sent\n", filename)
	fmt.Printf("%d bytes sent\n", n)
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
