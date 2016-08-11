package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"io/ioutil"

	"errors"

	"github.com/Songmu/prompter"
	"github.com/fatih/color"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ungerik/go-dry"
)

var (
	getCkan           = false
	getGlob           = false
	getShow           = false
	includeExtensions []string
)

type ckanPackage struct {
	ID         string   //ckan identifier
	Path       string   //filepath to package
	Depends    []string //dependecies. Change to []ckanPackage?
	Recommends []string //Recommended packages.
}

// searchMethod is custom type (enum)
type searchMethod int

const (
	Simple = iota //Simple search looks for any filename witch contains given string. Case insensitive.
	Glob          //Glob search uses standard globbing wildcards. If you don't use wildcards then you get strict search
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get identifier",
	Short: "Bring netkan of given identifier to your local repository",
	Long:  `Grab package from cache and move it to local/netkans directory. From there you can edit package and generate ckan packages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if c := checkWorkspace(); c != nil {
			return c
		}
		//get default extension for search if not set manually
		if includeExtensions == nil {
			includeExtensions = append(includeExtensions, viper.GetString("default_extension"))
		}
		//Looking for ckan is quiet common so this flag makes this faster
		//It means we looking for default + ckan!
		if getCkan {
			includeExtensions = append(includeExtensions, "ckan")
		}

		// var found []ckanPackage
		if len(args) == 0 {
			return errors.New("You need to provide id as argument!")
		}
		var idToFind = args[0]
		if verbose {
			fmt.Printf("Looking for extensions:\n%v\n", includeExtensions)
		}
		//search method
		var method searchMethod
		if getGlob {
			method = Glob // TODO: rename Glob? add regexp?
		} else {
			method = Simple
		}

		files, err := getFiles(idToFind, includeExtensions, method)
		if err != nil {
			return err
		}

		// show and select package
		var selectedPath string
		if len(files) == 0 { // not found
			Warn("Not found any packages\n")
			return nil
		} else if len(files) > 1 { // found many
			if len(args) == 2 { // user selected one
				n, err := strconv.Atoi(args[1])
				if err == nil && n <= len(files) { // second argument correct?
					selectedPath = files[n]
					Done("Found %s\n", filepath.Base(selectedPath))
				} else {
					Warn("Wrong second argument!\n")
					return errors.New("Select package by providing number (int) as second argument!")
				}
			} else { // show all found
				n := color.New(color.Bold).SprintfFunc()
				name := color.New(color.FgHiBlue).SprintfFunc()
				pwdc, err := getPwd()
				if err != nil {
					return err
				}
				pwdc = filepath.Join(pwdc, "cache", "repo")
				// result to display
				var result []string
				for i, e := range files {
					rel, _ := filepath.Rel(pwdc, e)
					result = append(result, fmt.Sprintf("%s | %s | %s\n", n("%d", i), name("%s", filepath.Base(e)), filepath.Dir(rel)))
				}
				fmt.Println(columnize.SimpleFormat(result))
				Done("Run the same command again with second argument to select package.\n")
				fmt.Printf("Example `kure get id 2` to get second pacage from the list\n")
				return nil
			}
		} else if len(files) == 1 { // only one found
			selectedPath = files[0]
			Done("Found %s\n", filepath.Base(selectedPath))
		} // todo add else
		// at this point package must be selected (selectedID)
		// handle -s flag - Show contenet of package
		if getShow {
			f, err := os.Open(selectedPath)
			if err != nil {
				return err
			}
			defer f.Close()
			if verbose {
				Done("Showing content of %s\n\n", selectedPath)
			}
			b := bytes.NewBufferString("")
			io.Copy(b, f)
			fmt.Println(b.String())
			return nil
		}
		// Lets get more info about this package.
		// save it to local/netkan? If no, then this will be saved to local/ckan

		var isNetkan bool
		if ext := filepath.Ext(selectedPath); ext == ".netkan" {
			isNetkan = true
		} else if ext == ".ckan" {
			isNetkan = false
		} else {
			Warn("If this package is netkan with chanded extension anwser yes. Otherwise pacage will be saved to local/ckan\n")
			isNetkan = prompter.YN("Is this valid netkan package?", false)
		}

		//copy
		pwd, err := getPwd()
		if err != nil {
			return err
		}

		if isNetkan {
			path := filepath.Join(pwd, "local", "netkan", filepath.Base(selectedPath))
			// handle remote netkans (netkan that reference another netkan)
			reg := regexp.MustCompile(`(?m)^\s*\"\$kref\"\s?\:\s?\"#\/ckan\/netkan\/(.+)\",?$`)
			f, e := ioutil.ReadFile(selectedPath)
			if e != nil {
				return errors.New("Could not read souce package while copying")
			}
			isRemote := reg.Match(f)
			if isRemote {
				res := reg.FindSubmatch(f)
				url := string(res[len(res)-1])
				Warn("Package %s is reference to remote package.\nURL: %s\n", filepath.Base(selectedPath), url)
				downloadIt := prompter.YN("Download it istead of copying package from cache?", true)
				if downloadIt {
					err = downloadFile(url, path, false)
					return err
				}
			}

			err = dry.FileCopy(selectedPath, path)
			if err != nil {
				return err
			}
			Done("Saved to local/netkan repository\n")
		} else {
			Done("Saved to local/ckan repository\n")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVarP(&getGlob, "glob", "g", false, "Search using glob pattern. Patterns are match against file name without extension.")
	getCmd.Flags().BoolVarP(&getCkan, "ckan", "c", false, "Search for ckan packages in cache. By default this means search for netkan and ckan files.")
	getCmd.Flags().BoolVarP(&getShow, "show", "s", false, "List source of given package.")
	getCmd.Flags().StringSliceVarP(&includeExtensions, "include", "i", nil, "Include given extensions to search. Default extension will not be included automaticly. Example `-i=txt,frozen`")
}

// get files with given extension and id.
func getFiles(id string, extensions []string, method searchMethod) ([]string, error) {
	pwd, err := getPwd()
	if err != nil {
		return nil, err
	}

	var result []string
	err = filepath.Walk(filepath.Join(pwd, "cache", "repo"),
		func(path string, f os.FileInfo, err error) error {
			// get extension
			ext := filepath.Ext(path)
			if !f.IsDir() && contains(extensions, strings.TrimPrefix(ext, ".")) {
				if method == Simple {
					if strings.Contains(strings.ToLower(f.Name()), strings.ToLower(id)) {
						result = append(result, path)
					}
				} else if method == Glob {
					if matched, _ := filepath.Match(id, strings.TrimSuffix(f.Name(), ext)); matched {
						result = append(result, path)
					}
				}
			}
			return nil
		})
	return result, err
}

func visit(path string, f os.FileInfo, err error) error {
	fmt.Printf("Visited: %s\n", path)
	return nil
}
