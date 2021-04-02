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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"

	//"github.com/vmware/gotftp"
	"github.com/pin/tftp"
)

func readHandler(filename string, rf io.ReaderFrom) error {

	// get the requesting ip-address and our source address
	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
	laddr := rf.(tftp.RequestPacketInfo).LocalIP()

	//strip the port
	ip, _, _ := net.SplitHostPort(raddr.String())

	//get the object that correlates with the ip
	var address models.Address
	db.DB.Preload(clause.Associations).First(&address, "ip = ?", ip)

	//get the image info that correlates with the pool the ip is in
	var image models.Image
	db.DB.First(&image, "id = ?", address.Group.ImageID)

	logrus.WithFields(logrus.Fields{
		"raddr":     raddr,
		"laddr":     laddr,
		"filename":  filename,
		"imageid":   image.ID,
		"addressid": address.ID,
	}).Debug("tftpd")

	//if the filename is mboot.efi, we hijack it and serve the mboot.efi file that is part of that specific image, this guarantees that you always get an mboot file that works for the build.
	if filename == "mboot.efi" {
		logrus.WithFields(logrus.Fields{
			ip: "requesting mboot.efi",
		}).Info("tftpd")
		filename, _ = mbootPath(image.Path)
	} else if strings.ToLower(filename) == "/boot.cfg" {
		//if the filename is boot.cfg, we serve the boot cfg that belongs to that build.
		logrus.WithFields(logrus.Fields{
			ip: "requesting boot.cfg",
		}).Info("tftpd")
		//filename = image.Path + "/BOOT.CFG"

		bc, err := ioutil.ReadFile(image.Path + "/BOOT.CFG")
		if err != nil {
			logrus.Warn(err)
			return err
		}

		// add kickstart path to kernelopt
		re := regexp.MustCompile("kernelopt=.*")
		o := re.Find(bc)
		bc = re.ReplaceAllLiteral(bc, append(o, []byte(" ks=http://"+laddr.String()+":8080/ks.cfg")...))

		// replace prefix with prefix=foldername
		split := strings.Split(image.Path, "/")
		re = regexp.MustCompile("prefix=")
		o = re.Find(bc)
		bc = re.ReplaceAllLiteral(bc, append(o, []byte(split[1])...))

		// strip slashes from paths in file
		re = regexp.MustCompile("/")
		o = re.Find(bc)
		bc = re.ReplaceAllLiteral(bc, append(o, []byte("")...))

		// Make a buffer to read from
		buff := bytes.NewBuffer(bc)
		spew.Dump(buff)
		// Send the data from the buffer to the client
		rf.(tftp.OutgoingTransfer).SetSize(int64(buff.Len()))
		n, err := rf.ReadFrom(buff)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return err
		}
		logrus.WithFields(logrus.Fields{
			"file":  filename,
			"bytes": n,
		}).Info("tftpd")
		return nil
	} else {
		spew.Dump(filename)
		if _, err := os.Stat("tftp/" + filename); err == nil {
			filename = "tftp/" + filename
		} else {
			dir, file := path.Split(filename)
			upperfile := strings.ToUpper(string(file))
			filename = "tftp/" + dir + upperfile
		}
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

/*
func modifyBootCfg(path string) bool {
	//determine if the boot.cfg file is stored as lowercase or uppercase
	if _, err := os.Stat(path); err == nil {
		filename := "tftp/" + filename
	} else {
		dir, file := path.Split(filename)
		upperfile := strings.ToUpper(string(file))
		filename := "tftp/" + dir + upperfile
	}
}
*/

func mbootPath(imagePath string) (string, error) {
	//check these paths if the file exists.
	paths := []string{"/EFI/BOOT/BOOTX64.EFI", "/MBOOT.EFI", "/mboot.efi", "/efi/boot/bootx64.efi"}

	for _, v := range paths {
		if _, err := os.Stat(imagePath + v); err == nil {
			return imagePath + v, nil
		}
	}
	//couldn't find the file
	return "", fmt.Errorf("could not locate a mboot.efi")

}

func checkUpperLower(myfile string) bool {
	rune := []rune(myfile)
	res := unicode.IsLower(rune[0])
	return res
}
