package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

// ConvertObjectStore converts API ObjectStore to domain model
func ConvertObjectStore(in *openapi.ObjectStore) *models.ObjectStore {
	if in == nil {
		return nil
	}

	res := &models.ObjectStore{
		Name:            in.GetName(),
		DestinationPath: in.Spec.GetDestinationPath(),
		RetentionPolicy: in.Spec.GetRetentionPolicy(),
		Configuration:   convertObjectStoreConfiguration(in.Spec.GetConfiguration()),
	}

	return res
}

// PresentObjectStore converts domain model to API ObjectStore
func PresentObjectStore(in *models.ObjectStore) openapi.ObjectStore {
	res := openapi.ObjectStore{}

	res.SetId(in.ID)
	res.SetOrganisationId(in.OrganisationID)
	res.SetName(in.Name)
	res.SetCreatedAt(in.CreatedAt)
	res.SetUpdatedAt(in.UpdatedAt)

	// Set spec
	spec := openapi.ObjectStoreSpec{
		Configuration:   presentObjectStoreConfiguration(in.Configuration),
		DestinationPath: in.DestinationPath,
	}

	if in.RetentionPolicy != "" {
		spec.SetRetentionPolicy(in.RetentionPolicy)
	}

	res.SetSpec(spec)

	return res
}

// PresentObjectStoreList converts list of domain models to API list
func PresentObjectStoreList(in []*models.ObjectStore) []openapi.ObjectStore {
	if len(in) == 0 {
		return []openapi.ObjectStore{}
	}

	result := make([]openapi.ObjectStore, len(in))
	for i, store := range in {
		result[i] = PresentObjectStore(store)
	}
	return result
}

// Helper functions for object store configuration

func convertObjectStoreConfiguration(in openapi.ObjectStoreConfiguration) models.ObjectStoreConfiguration {
	res := models.ObjectStoreConfiguration{}

	if in.S3Credentials != nil {
		res.S3Credentials = &models.S3Credentials{
			AccessKeyID:     in.S3Credentials.GetAccessKeyId(),
			SecretAccessKey: in.S3Credentials.GetSecretAccessKey(),
			Region:          in.S3Credentials.GetRegion(),
			Endpoint:        in.S3Credentials.GetEndpointUrl(),
		}
	}

	if in.AzureCredentials != nil {
		res.AzureCredentials = &models.AzureCredentials{
			ConnectionString: in.AzureCredentials.GetConnectionString(),
		}
	}

	if in.GcsCredentials != nil {
		res.GCSCredentials = &models.GCSCredentials{
			ServiceAccountKey: in.GcsCredentials.GetServiceAccountKey(),
		}
	}

	return res
}

func presentObjectStoreConfiguration(in models.ObjectStoreConfiguration) openapi.ObjectStoreConfiguration {
	res := openapi.ObjectStoreConfiguration{}

	if in.S3Credentials != nil {
		s3 := openapi.S3Credentials{}
		s3.SetAccessKeyId(in.S3Credentials.AccessKeyID)
		s3.SetSecretAccessKey(in.S3Credentials.SecretAccessKey)
		s3.SetRegion(in.S3Credentials.Region)
		if in.S3Credentials.Endpoint != "" {
			s3.SetEndpointUrl(in.S3Credentials.Endpoint)
		}
		res.SetS3Credentials(s3)
	}

	if in.AzureCredentials != nil {
		azure := openapi.AzureCredentials{}
		azure.SetConnectionString(in.AzureCredentials.ConnectionString)
		res.SetAzureCredentials(azure)
	}

	if in.GCSCredentials != nil {
		gcs := openapi.GCSCredentials{}
		gcs.SetServiceAccountKey(in.GCSCredentials.ServiceAccountKey)
		res.SetGcsCredentials(gcs)
	}

	return res
}
