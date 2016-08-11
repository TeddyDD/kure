// Copyright © 2016 Daniel Lewan @TeddyDD
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mholt/archiver"
	"github.com/spf13/cobra"
)

var port string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Pack your ckans into tar.gz and start local web server that will host this file.",
	Long: `Local web server allows you to load your ckans directly into ckan client, like from
	ordinary ckan repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if c := checkWorkspace(); c != nil {
			return c
		}
		pwd, _ := os.Getwd()
		// pack repository
		localPath := filepath.Join("local", "ckan")
		if verbose {
			fmt.Printf("Adding files from %s\n", localPath)
		}
		files, err := ioutil.ReadDir(localPath)
		if err != nil {
			return err
		}
		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, filepath.Join(localPath, f.Name()))
		}
		if verbose {
			fmt.Println("Creating tar.gz")
		}
		archiver.TarGz(filepath.Join("cache", "server", "main.tar.gz"), fileNames)

		// serve repository
		Done("Starting server. CTRL-C to stop. Addres:\n")
		fmt.Printf("%s | localhost:%s/main.tar.gz\n", filepath.Base(pwd), port)
		Done("Paste it into CKAN-Settings>New\n")

		fs := http.FileServer(http.Dir("./cache/server"))
		http.Handle("/", fs)
		log.Fatal(http.ListenAndServe(":"+port, nil))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&port, "port", "p", "8000", "localhost port")
}
