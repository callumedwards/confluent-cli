package apikey

import (
	"time"

	"github.com/spf13/cobra"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/featureflags"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type out struct {
	IsCurrent    bool   `human:"Current,omitempty" serialized:"is_current,omitempty"`
	Key          string `human:"Key" serialized:"key"`
	Description  string `human:"Description" serialized:"description"`
	OwnerId      string `human:"Owner" serialized:"owner_id"`
	OwnerEmail   string `human:"Owner Email" serialized:"owner_email"`
	ResourceType string `human:"Resource Type" serialized:"resource_type"`
	ResourceId   string `human:"Resource" serialized:"resource_id"`
	Created      string `human:"Created" serialized:"created"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an API key.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	c.setKeyStoreIfNil()

	apiKey, httpResp, err := c.V2Client.GetApiKey(args[0])
	if err != nil {
		return errors.CatchApiKeyForbiddenAccessError(err, getOperation, httpResp)
	}

	var ownerId string
	var email string

	if apiKey.Spec.HasOwner() {
		allUsers, err := c.getAllUsers()
		if err != nil {
			return err
		}
		resourceIdToUserIdMap := mapResourceIdToUserId(allUsers)
		usersMap := getUsersMap(allUsers)

		serviceAccounts, err := c.V2Client.ListIamServiceAccounts()
		if err != nil {
			return err
		}
		serviceAccountsMap := getServiceAccountsMap(serviceAccounts)

		ownerId = apiKey.Spec.Owner.GetId()
		auditLogServiceAccountId := c.getAuditLogServiceAccountId()
		email = c.getEmail(ownerId, auditLogServiceAccountId, resourceIdToUserIdMap, usersMap, serviceAccountsMap)
	}

	resources := []apikeysv2.ObjectReference{apiKey.Spec.GetResource()}

	// Check if multicluster keys are enabled, and if so check the resources field
	if featureflags.Manager.BoolVariation("cli.multicluster-api-keys.enable", c.Context, config.CliLaunchDarklyClient, true, false) {
		resources = apiKey.Spec.GetResources()
	}

	list := output.NewList(cmd)
	// Note that if more resource types are added with no logical clusters, then additional logic
	// needs to be added here to determine the resource type.
	for _, res := range resources {
		list.Add(&out{
			Key:          apiKey.GetId(),
			Description:  apiKey.Spec.GetDescription(),
			OwnerId:      ownerId,
			OwnerEmail:   email,
			ResourceType: resourceKindToType[res.GetKind()],
			ResourceId:   getApiKeyResourceId(res.GetId()),
			Created:      apiKey.Metadata.GetCreatedAt().Format(time.RFC3339),
		})
	}
	return list.Print()
}
