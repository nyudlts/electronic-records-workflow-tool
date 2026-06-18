package cmd

import (
	"fmt"

	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	projectInitCmd.Flags().StringVarP(&collectionCode, "collection-code", "c", "", "the collection code to use for adoc")
	projectInitCmd.Flags().StringVarP(&sourceLoc, "source-location", "s", "", "the source location for the collection")
	projectInitCmd.Flags().StringVarP(&ewtConfig, "config", "j", "", "location of ewt config file (defaults to $HOME/.config/ewt.config)")
	projectInitCmd.MarkFlagRequired("collection-code")
	projectInitCmd.MarkFlagRequired("source-location")
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectArchiveCmd)
	rootCmd.AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "ewt project commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Help())
	},
}

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a EWT project",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.InitProject(collectionCode, sourceLoc, ewtConfig); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var projectArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a EWT Project",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ArchiveProject(args[0]); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}
