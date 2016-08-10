package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verboseNetkan    = false
	prereleaseNetkan = false
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [package]",
	Short: "generate ckan packages from your local netkan files",
	Long: `This command use netkan.exe tool to generate ckan packages from your local/netkan metadata.
	Generated packages are saved to local/ckan. You must have netkan.exe tool in cache/bin. You can download it
	using "kure update -n". Without arguments this command will generate ckan files from all local/netkan packages`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if c := checkWorkspace(); c != nil {
			return c
		}
		pwd, _ := getPwd()
		var ckanArgs []string
		// Args for all files
		if verbose {
			ckanArgs = append(ckanArgs, "--verbose")
		}
		// cacheDir, _ := filepath.Abs(viper.GetString("cachedir"))
		cacheDir := viper.GetString("cachedir")
		cacheDir, _ = filepath.Abs(cacheDir)
		fmt.Println("PWD ", pwd)
		// cacheDir := filepath.Join(pwd, "cache", "download")
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			Warn("Directory %s not exist. Using default cache.\n", cacheDir)
		}
		outputDir, _ := filepath.Abs("./local/ckan")
		ckanArgs = append(ckanArgs, fmt.Sprintf("--outputdir=\"%s\"", outputDir))
		ckanArgs = append(ckanArgs, fmt.Sprintf("--cachedir=\"%s\"", cacheDir))
		ckanExe := filepath.Join(pwd, "cache", "bin", "netkan.exe")

		if len(args) == 0 {
			err := filepath.Walk(filepath.Join(pwd, "local", "netkan"),
				func(path string, f os.FileInfo, err error) error {
					if !f.IsDir() {
						finalArgs := append(ckanArgs, path)
						fmt.Printf("File %s Args: %v\n", f.Name(), ckanArgs)
						cmd := exec.Command(ckanExe, finalArgs...)
						o, e := cmd.Output()
						Done("netkan output for %s:\n%s\n", f.Name(), string(o))
						return e
					}
					return nil
				})
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVarP(&verboseNetkan, "verbose-netkan", "V", false, "Print verbose output of netkan.exe tool")
	buildCmd.Flags().BoolVarP(&prereleaseNetkan, "prerelease", "p", false, "netkan.exe tool will index github prereleases")
}
