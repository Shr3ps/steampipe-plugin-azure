package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

//// TABLE DEFINITION ////

func tableAzureComputeDiskEncryptionSet(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_compute_disk_encryption_set",
		Description: "Azure Compute Disk Encryption Set",
		Get: &plugin.GetConfig{
			KeyColumns:        plugin.AllColumns([]string{"name", "resource_group"}),
			Hydrate:           getAzureComputeDiskEncryptionSet,
			ShouldIgnoreError: isNotFoundError([]string{"ResourceGroupNotFound", "ResourceNotFound", "404"}),
		},
		List: &plugin.ListConfig{
			Hydrate: listAzureComputeDiskEncryptionSets,
		},
		Columns: []*plugin.Column{
			{
				Name:        "name",
				Description: "The friendly name that identifies the disk encryption set",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The unique id identifying the resource in subscription",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "type",
				Description: "The type of the resource in Azure",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "provisioning_state",
				Description: "The disk encryption set provisioning state",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("EncryptionSetProperties.ProvisioningState"),
			},
			{
				Name:        "active_key_source_vault_id",
				Description: "Resource id of the KeyVault containing the key or secret",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("EncryptionSetProperties.ActiveKey.SourceVault.ID"),
			},
			{
				Name:        "active_key_url",
				Description: "Url pointing to a key or secret in KeyVault",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("EncryptionSetProperties.ActiveKey.KeyURL"),
			},
			{
				Name:        "encryption_type",
				Description: "Contains the type of the encryption",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("EncryptionSetProperties.EncryptionType").Transform(transform.ToString),
			},
			{
				Name:        "identity_principal_id",
				Description: "The object id of the Managed Identity Resource",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Identity.PrincipalID"),
			},
			{
				Name:        "identity_tenant_id",
				Description: "The tenant id of the Managed Identity Resource",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Identity.TenantID"),
			},
			{
				Name:        "identity_type",
				Description: "The type of Managed Identity used by the DiskEncryptionSet",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Identity.Type").Transform(transform.ToString),
			},
			{
				Name:        "previous_keys",
				Description: "A list of key vault keys previously used by this disk encryption set while a key rotation is in progress",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("EncryptionSetProperties.PreviousKeys"),
			},

			// Standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "tags",
				Description: resourceInterfaceDescription("tags"),
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ID").Transform(idToAkas),
			},
			{
				Name:        "region",
				Description: "The Azure region in which the resource is located",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Location"),
			},
			{
				Name:        "resource_group",
				Description: "Name of the resource group in which the resource is located",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID").Transform(extractResourceGroupFromID),
			},
			{
				Name:        "subscription_id",
				Description: "The Azure Subscription ID in which the resource is located",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID").Transform(idToSubscriptionID),
			},
		},
	}
}

//// LIST FUNCTION ////

func listAzureComputeDiskEncryptionSets(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("listAzureComputeDiskEncryptionSets")
	session, err := GetNewSession(ctx, d.ConnectionManager, "MANAGEMENT")
	if err != nil {
		return nil, err
	}

	subscriptionID := session.SubscriptionID
	client := compute.NewDiskEncryptionSetsClient(subscriptionID)
	client.Authorizer = session.Authorizer
	pagesLeft := true

	for pagesLeft {
		result, err := client.List(context.Background())
		if err != nil {
			return nil, err
		}

		for _, diskEncryptionSet := range result.Values() {
			d.StreamListItem(ctx, diskEncryptionSet)
		}
		result.NextWithContext(context.Background())
		pagesLeft = result.NotDone()
	}

	return nil, nil
}

//// HYDRATE FUNCTION ////

func getAzureComputeDiskEncryptionSet(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAzureComputeDiskEncryptionSet")

	name := d.KeyColumnQuals["name"].GetStringValue()
	resourceGroup := d.KeyColumnQuals["resource_group"].GetStringValue()

	session, err := GetNewSession(ctx, d.ConnectionManager, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID
	client := compute.NewDiskEncryptionSetsClient(subscriptionID)
	client.Authorizer = session.Authorizer

	op, err := client.Get(context.Background(), resourceGroup, name)
	if err != nil {
		return nil, err
	}

	// In some cases resource does not give any notFound error
	// instead of notFound error, it returns empty data
	if op.ID != nil {
		return op, nil
	}

	return nil, nil
}
