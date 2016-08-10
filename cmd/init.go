package cmd

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ungerik/go-dry"
)

// DirPerm is default directory permission
const DirPerm = 0700

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init directory",
	Short: "Create new directory with KURE workspace",
	Long: `Create new directory with KURE workspace
	KURE workspace contains your local netkan files, generated ckans and config file
	As argument you have to provide directory name. Directory should exist and be empty`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return newCommandError("No directory specified", "init", true)
		}

		pwd, err := os.Getwd()
		if err != nil {
			return newCommandError(err.Error(), "init", false)
		}

		//create directory
		path := filepath.Join(pwd, args[0])
		if _, err := os.Stat(path); err == nil {
			return newCommandError("Directory already exist", "init", true)
		}

		//create directory
		err = os.Mkdir(path, DirPerm)
		if err != nil {
			return newCommandError(err.Error(), "init", false)
		}
		os.Mkdir(filepath.Join(path, "local"), DirPerm)
		os.MkdirAll(filepath.Join(path, "local", "netkan"), DirPerm)
		os.MkdirAll(filepath.Join(path, "local", "ckan"), DirPerm)
		os.MkdirAll(filepath.Join(path, "cache", "repo"), DirPerm)
		os.MkdirAll(filepath.Join(path, "cache", "server"), DirPerm)
		os.MkdirAll(filepath.Join(path, "cache", "download"), DirPerm)
		os.MkdirAll(filepath.Join(path, "cache", "bin"), DirPerm)

		dry.FileAppendString(filepath.Join(path, "kure.json"), `
{
    "netkan_exe": "https://ckan-travis.s3.amazonaws.com/netkan.exe",
    "repos": [
        {
            "name": "Offical Netkan",
            "type": "netkan",
            "url": "https://github.com/KSP-CKAN/NetKAN/archive/master.tar.gz"
        },
        {
            "name": "Offical Ckan",
            "type": "ckan",
            "url": "https://github.com/KSP-CKAN/CKAN-Meta/archive/master.tar.gz"
        }
    ]
}
		`)
		//Download netkan
		Done("Workspace %s created\n", args[0])
		fmt.Println("Run `kure update -n` and `kure update` from workspace to download netkan.exe and packages.")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
