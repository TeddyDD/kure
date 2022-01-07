package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
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
		var err error
		if len(args) > 0 {
			for _, p := range args {
				p, err := filepath.Abs(p)
				if err != nil {
					return err
				}
				err = updateNetkanFile(p)
			}
		} else {
			err = updateAll()
		}
		return err
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVarP(&verboseNetkan, "verbose-netkan", "V", false, "Print verbose output of netkan.exe tool")
	buildCmd.Flags().BoolVarP(&prereleaseNetkan, "prerelease", "p", false, "netkan.exe tool will index github prereleases")
}

func updateAll() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = filepath.Walk(filepath.Join(pwd, "local", "netkan"),
		func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() {
				Done("Building %s\n", filepath.Base(path))
				return updateNetkanFile(path)
			}
			return nil
		})

	return err
}

func updateNetkanFile(path string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	outputDir := filepath.Join("local", "ckan")
	netkanFile, err := filepath.Rel(pwd, path)
	if err != nil {
		return err
	}
	// netkan flags
	netkanVerboseFlag := ""
	if verboseNetkan {
		netkanVerboseFlag = "--verbose"
	}
	netkanPrereleaseFlag := ""
	if prereleaseNetkan {
		netkanPrereleaseFlag = "--prerelease"
	}

	mono, err := exec.LookPath("mono")
	if err != nil {
		return err
	}
	netkan, err := filepath.Abs(filepath.Join("cache", "bin", "netkan.exe"))
	if err != nil {
		return err
	}

	cmd := exec.Command(
		mono,
		netkan,
		"--outputdir="+outputDir,
		netkanPrereleaseFlag,
		netkanVerboseFlag,
		netkanFile,
	)
	fmt.Printf("%+v\n", cmd)

	out, err := cmd.CombinedOutput()

	fmt.Print(string(out))
	if err != nil {
		return err
	}
	return nil
}
