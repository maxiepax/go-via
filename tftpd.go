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
	"io"
	"log"
	"net/http"
	"os"

	//"github.com/davecgh/go-spew/spew"
	//"github.com/vmware/gotftp"
	"github.com/zwh8800/tftp"
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

func TFTPd() {
	log.Panic(tftp.ListenAndServe(":69", tftp.ReadonlyFileServer(http.Dir(pwd+"/tftp"))))
	pwd, err := os.Getwd()
	fmt.println(pwd + "/tftp")
	tftp.HandleFunc("test", func(w io.WriteCloser, req *tftp.Request) error {
		log.Println("incoming read operation for test:", req)

		f, _ := os.Open("tftp/test")
		io.Copy(w, f)
		f.Close() // don't forget close the writer

		return nil
	}, nil)
}
