package flink

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

const serviceAccountWarning = "[WARN] No service account provided. To ensure that your statements run continuously, " +
	"use a service account instead of your user identity with `confluent iam service-account use` or `--service-account`. " +
	"Otherwise, statements will stop running after 4 hours."

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "flink",
		Short:       "Manage Apache Flink.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newComputePoolCommand())
	cmd.AddCommand(c.newRegionCommand())
	cmd.AddCommand(c.newShellCommand(prerunner))
	cmd.AddCommand(c.newStatementCommand())

	return cmd
}

func (c *command) addComputePoolFlag(cmd *cobra.Command) {
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "compute-pool", c.autocompleteComputePools)
}

func (c *command) autocompleteComputePools(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	computePools, err := c.V2Client.ListFlinkComputePools(environmentId, "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(computePools))
	for i, computePool := range computePools {
		suggestions[i] = fmt.Sprintf("%s\t%s", computePool.GetId(), computePool.Spec.GetDisplayName())
	}
	return suggestions
}

func (c *command) addDatabaseFlag(cmd *cobra.Command) {
	cmd.Flags().String("database", "", "The database which will be used as the default database. When using Kafka, this is the cluster ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "database", c.autocompleteDatabases)
}

func (c *command) autocompleteDatabases(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}

func (c *command) addRegionFlag(cmd *cobra.Command) {
	cmd.Flags().String("region", "", `Cloud region for compute pool (use "confluent flink region list" to see all).`)
	pcmd.RegisterFlagCompletionFunc(cmd, "region", c.autocompleteRegions)
}

func (c *command) autocompleteRegions(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	cloud, _ := cmd.Flags().GetString("cloud")

	regions, err := c.V2Client.ListFlinkRegions(strings.ToUpper(cloud))
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(regions))
	for i, region := range regions {
		suggestions[i] = region.GetRegionName()
	}
	return suggestions
}
