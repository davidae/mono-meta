package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	file        string
	local       string
	remote      string
	buildCMD    string
	clonePath   string
	servicePath string
	exclude     []string
)

func main() {
	commonFlagSet := pflag.NewFlagSet("common", pflag.ExitOnError)
	commonFlagSet.StringVarP(&file, "file", "f", "", "load configuration file, it will override all other flags set")
	commonFlagSet.StringVarP(&local, "local", "l", "", "local URI to git repository")
	commonFlagSet.StringVarP(&remote, "remote", "r", "", "remote URI to git repository")
	commonFlagSet.StringVarP(&buildCMD, "cmd", "c", "go build -o $1", "build command when building services, a '$1' variable of the outputted binary is required")
	commonFlagSet.StringVarP(&clonePath, "clone-path", "cp", "~/.mono-meta", "local path where a remote git repository is cloned into")
	commonFlagSet.StringVarP(&servicePath, "services", "s", "", "directory/path pattern of where the services resides in the git repositry")
	exclude = *commonFlagSet.StringSliceP("exclude", "e", []string{}, "exclude any directories inside the service path with a comma seperated list")

	root := &cobra.Command{Use: "mono-meta"}
	root.AddCommand(Diff, Services)
}
