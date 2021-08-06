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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/maxiepax/go-via/config"
	"github.com/maxiepax/go-via/db"
	"github.com/maxiepax/go-via/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"

	"github.com/pin/tftp"
)

//func readHandler(filename string, rf io.ReaderFrom) error {
func readHandler(conf *config.Config) func(string, io.ReaderFrom) error {
	return func(filename string, rf io.ReaderFrom) error {

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
			logrus.WithFields(logrus.Fields{
				"id":           address.ID,
				"percentage":   10,
				"progresstext": "mboot.efi",
			}).Info("progress")
			filename, _ = mbootPath(image.Path)
			address.Progress = 10
			address.Progresstext = "mboot.efi"
			db.DB.Save(&address)
		} else if filename == "crypto64.efi" {
			logrus.WithFields(logrus.Fields{
				ip: "requesting crypto64.efi",
			}).Info("tftpd")
			logrus.WithFields(logrus.Fields{
				"id":           address.ID,
				"percentage":   12,
				"progresstext": "crypto64.efi",
			}).Info("progress")
			filename, _ = crypto64Path(image.Path)
			address.Progress = 12
			address.Progresstext = "crypto64.efi"
			db.DB.Save(&address)
		} else if (strings.ToLower(filename) == "boot.cfg") || (strings.ToLower(filename) == "/boot.cfg") {
			//if the filename is boot.cfg, or /boot.cfg, we serve the boot cfg that belongs to that build. unfortunately, it seems boot.cfg or /boot.cfg varies in builds.
			logrus.WithFields(logrus.Fields{
				ip: "requesting boot.cfg",
			}).Info("tftpd")
			logrus.WithFields(logrus.Fields{
				"id":           address.ID,
				"percentage":   15,
				"progresstext": "installation",
			}).Info("progress")
			address.Progress = 15
			address.Progresstext = "installation"
			db.DB.Save(&address)

			bc, err := ioutil.ReadFile(image.Path + "/BOOT.CFG")
			if err != nil {
				logrus.Warn(err)
				return err
			}

			// strip slashes from paths in file
			re := regexp.MustCompile("/")
			bc = re.ReplaceAllLiteral(bc, []byte(""))

			// add kickstart path to kernelopt
			re = regexp.MustCompile("kernelopt=.*")
			o := re.Find(bc)
			bc = re.ReplaceAllLiteral(bc, append(o, []byte(" ks=https://"+laddr.String()+":"+strconv.Itoa(conf.Port)+"/ks.cfg")...))

			options := models.GroupOptions{}
			json.Unmarshal(address.Group.Options, &options)

			// add allowLegacyCPU=true to kernelopt
			if options.AllowLegacyCPU {
				re = regexp.MustCompile("kernelopt=.*")
				o = re.Find(bc)
				bc = re.ReplaceAllLiteral(bc, append(o, []byte(" allowLegacyCPU=true")...))
			}

			// replace prefix with prefix=foldername
			split := strings.Split(image.Path, "/")
			re = regexp.MustCompile("prefix=")
			o = re.Find(bc)
			bc = re.ReplaceAllLiteral(bc, append(o, []byte(split[1])...))

			// Make a buffer to read from
			buff := bytes.NewBuffer(bc)

			// Send the data from the buffer to the client
			rf.(tftp.OutgoingTransfer).SetSize(int64(buff.Len()))
			n, err := rf.ReadFrom(buff)
			if err != nil {
				//fmt.Fprintf(os.Stderr, "%v\n", err)
				logrus.WithFields(logrus.Fields{
					"os.Stderr": err,
				}).Info("tftpd")
				return err
			}

			logrus.WithFields(logrus.Fields{
				"file":  filename,
				"bytes": n,
			}).Info("tftpd")
			return nil
		} else {
			//chroot the rest
			if _, err := os.Stat("tftp/" + filename); err == nil {
				filename = "tftp/" + filename
				logrus.WithFields(logrus.Fields{
					"lowercase file": filename,
				}).Debug("tftpd")
			} else {
				dir, file := path.Split(filename)
				upperfile := strings.ToUpper(string(file))
				filename = "tftp/" + dir + upperfile
				logrus.WithFields(logrus.Fields{
					"uppercase file": filename,
				}).Debug("tftpd")
			}
		}

		// get the filesize to send filelength
		fi, err := os.Stat(filename)
		if err != nil {
			return err
		}

		rf.(tftp.OutgoingTransfer).SetSize(fi.Size())

		file, err := os.Open(filename)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"could not open file": err,
			}).Debug("tftpd")
			return err
		}
		n, err := rf.ReadFrom(file)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"could not read from file": err,
			}).Debug("tftpd")
			return err
		}
		logrus.WithFields(logrus.Fields{
			"id":    address.ID,
			"ip":    address.IP,
			"host":  address.Hostname,
			"file":  filename,
			"bytes": n,
		}).Info("tftpd")
		return nil
	}
}

func TFTPd(conf *config.Config) {
	s := tftp.NewServer(readHandler(conf), nil)
	s.SetTimeout(5 * time.Second)  // optional
	err := s.ListenAndServe(":69") // blocks until s.Shutdown() is called
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"could not start tftp server:": err,
		}).Info("tftpd")
		os.Exit(1)
	}
}

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

func crypto64Path(imagePath string) (string, error) {
	//check these paths if the file exists.
	paths := []string{"/EFI/BOOT/CRYPTO64.EFI", "/efi/boot/crypto64.efi"}

	for _, v := range paths {
		if _, err := os.Stat(imagePath + v); err == nil {
			return imagePath + v, nil
		}
	}
	//couldn't find the file
	return "", fmt.Errorf("could not locate a crypto64.efi")

}
