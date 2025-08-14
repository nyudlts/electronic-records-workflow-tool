package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	projectInitCmd.Flags().StringVarP(&collectionCode, "collection-code", "c", "", "the collection code to use for adoc")
	projectInitCmd.Flags().StringVarP(&sourceLoc, "source-location", "s", "", "the source location for the collection")
	projectCmd.AddCommand(projectInitCmd)
	projectArchiveCmd.Flags().StringVarP(&projectLoc, "project-location", "p", "", "Project name")
	projectCmd.AddCommand(projectArchiveCmd)
	rootCmd.AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "ewt project commands",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a EWT project",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.InitProject(collectionCode, sourceLoc); err != nil {
			panic(err)
		}
	},
}

var projectArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a EWT Project",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ArchiveProject(projectLoc); err != nil {
			panic(err)
		}
	},
}
