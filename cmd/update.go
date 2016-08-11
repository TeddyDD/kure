package cmd

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/bitly/go-simplejson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	netkan  = false
	clean   = false
	noClean = false
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local netkan archive and check for netkan.exe updates",
	Long:  `All netkans from repo are cached. This command will update cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if c := checkWorkspace(); c != nil {
			return c
		}
		if netkan {
			return downloadNetkan()
		}
		if clean {
			if verbose {
				fmt.Println("Cleaning up cache/repo/")
			}
			err := cleanRepo()
			if err != nil {
				return errors.New("Error while cleaning repo cache: " + err.Error())
			}
			return err
		}

		return downloadRepos()
		// return nil
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&netkan, "netkan", "n", false, "Update netkan tool")
	updateCmd.Flags().BoolVarP(&clean, "clean", "c", false, "Remove cached netkan packages.")
	updateCmd.Flags().BoolVarP(&noClean, "no-clean", "C", false, "Disable automatic cleaning befroe update")
}

func downloadNetkan() error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.New("Cannot get working directory")
	}
	path := filepath.Join(pwd, "cache", "bin", "netkan.exe")
	url := viper.GetString("netkan_exe")
	err = downloadFile(url, path, true)
	if err != nil {
		return err
	}
	return nil
}

func downloadRepos() error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.New("Cannot get working directory")
	}
	file, err := os.Open("kure.json")
	if err != nil {
		return err
	}
	json, err := simplejson.NewFromReader(file)
	if err != nil {
		return err
	}
	// names of already downloaded repos
	var done []string

	if !noClean {
		if verbose {
			Warn("Cleaning up old cached\n")
		}
		err = cleanRepo()
		if err != nil {
			Warn("Could not clean repo cache.\n")
			fmt.Println("Update will proceed but you should clean repo cache manually and update again.")
		}
	}

	for _, v := range json.Get("repos").MustArray() {
		vv, found := v.(map[string]interface{})
		if !found {
			return errors.New("Not found repos array in config file")
		}

		repoName, found := vv["name"].(string)
		if !found {
			return errors.New("Not found type of repo")
		}

		repoType, found := vv["type"].(string)
		if !found {
			return errors.New("Not found type of repo")
		}

		url, found := vv["url"].(string)
		if !found {
			return errors.New("Not found url of repo")
		}

		//check if not downloaded (enforce uniqe repo names)
		if !contains(done, repoName) {
			if verbose {
				fmt.Printf("Downloading %s (%s) repo\nUrl: %s\n", repoName, repoType, url)
			}
			// download tar.gz
			path := filepath.Join(pwd, "cache", "repo", repoType+"_"+repoName+".tar.gz")
			err = downloadFile(url, path, false)
			if err != nil {
				return err
			}

			//unpack tar.gz
			final := filepath.Join(pwd, "cache", "repo", repoName)
			if verbose {
				fmt.Printf("Unpackging %s to %s\n", path, final)
			}
			var tgz *os.File
			tgz, err = os.Open(path)
			defer tgz.Close()
			err = unTarGz(final, tgz)
			if err != nil {
				return err
			}

			done = append(done, repoName)
		} else {
			fmt.Printf("Warning: repo name `%s` is not uniqe. Ignoring `%s` url.\n", repoName, url)
		}
	}
	Done("Update finished\n")
	return nil
}

func downloadFile(url, path string, exe bool) error {
	var out *os.File
	var err error
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("Creating file %s\n", path)
		}
		out, err = os.Create(path)
	} else {
		if verbose {
			fmt.Printf("Overwrite %s\n", filepath.Base(path))
		}
		out, err = os.OpenFile(path, os.O_WRONLY, 0700)
	}
	defer out.Close()
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("Downloading file\nfrom: %s\nto: %s\n", url, path)
	}
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	if exe {
		err = os.Chmod(path, 0700)
	} else {
		err = os.Chmod(path, 0600)
	}
	if err != nil {
		return err
	}
	return nil
}

func cleanRepo() error {
	pwd, err := getPwd()
	if err != nil {
		return err
	}
	repoPath := filepath.Join(pwd, "cache", "repo")
	err = os.RemoveAll(repoPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(repoPath, DirPerm)
	return err
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func unTarGz(targetdir string, reader io.ReadCloser) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		target := path.Join(targetdir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			setAttrs(target, header)
			break

		case tar.TypeReg:
			w, err := os.Create(target)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, tarReader)
			if err != nil {
				return err
			}
			w.Close()

			setAttrs(target, header)
			break
		case tar.TypeXGlobalHeader:
			if verbose {
				fmt.Println("header type X ignored")
			}

		default:
			fmt.Printf("unsupported type: %v\n", header.Typeflag)
			break
		}
	}

	return nil
}

func setAttrs(target string, header *tar.Header) {
	os.Chmod(target, os.FileMode(header.Mode))
	os.Chtimes(target, header.AccessTime, header.ModTime)
}

// return current working directory
func getPwd() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", errors.New("Cannot get working directory")
	}
	return pwd, nil
}
