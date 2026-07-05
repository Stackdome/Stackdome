package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

// convertSecretReference converts API SecretReference to domain model
func convertSecretReference(in openapi.SecretReference) models.SecretReference {
	return models.SecretReference{
		SecretID: in.GetSecretId(),
		Key:      in.GetKey(),
	}
}

// presentSecretReference converts domain model to API SecretReference
func presentSecretReference(in models.SecretReference) openapi.SecretReference {
	ref := openapi.SecretReference{}
	ref.SetSecretId(in.SecretID)
	ref.SetKey(in.Key)
	return ref
}

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
	res.SetTeamId(in.TeamID)
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
			AccessKeyID:     convertSecretReference(in.S3Credentials.GetAccessKeyId()),
			SecretAccessKey: convertSecretReference(in.S3Credentials.GetSecretAccessKey()),
			Region:          in.S3Credentials.GetRegion(),
			Endpoint:        in.S3Credentials.GetEndpointUrl(),
		}
	}

	if in.AzureCredentials != nil {
		res.AzureCredentials = &models.AzureCredentials{
			ConnectionString:   convertSecretReference(in.AzureCredentials.GetConnectionString()),
			StorageAccountName: in.AzureCredentials.GetStorageAccountName(),
		}
	}

	if in.GcsCredentials != nil {
		res.GCSCredentials = &models.GCSCredentials{
			ServiceAccountCredentials: convertSecretReference(in.GcsCredentials.GetServiceAccountCredentials()),
		}
	}

	return res
}

func presentObjectStoreConfiguration(in models.ObjectStoreConfiguration) openapi.ObjectStoreConfiguration {
	res := openapi.ObjectStoreConfiguration{}

	if in.S3Credentials != nil {
		s3 := openapi.S3Credentials{}
		s3.SetAccessKeyId(presentSecretReference(in.S3Credentials.AccessKeyID))
		s3.SetSecretAccessKey(presentSecretReference(in.S3Credentials.SecretAccessKey))
		s3.SetRegion(in.S3Credentials.Region)
		if in.S3Credentials.Endpoint != "" {
			s3.SetEndpointUrl(in.S3Credentials.Endpoint)
		}
		res.SetS3Credentials(s3)
	}

	if in.AzureCredentials != nil {
		azure := openapi.AzureCredentials{}
		azure.SetConnectionString(presentSecretReference(in.AzureCredentials.ConnectionString))
		if in.AzureCredentials.StorageAccountName != "" {
			azure.SetStorageAccountName(in.AzureCredentials.StorageAccountName)
		}
		res.SetAzureCredentials(azure)
	}

	if in.GCSCredentials != nil {
		gcs := openapi.GCSCredentials{}
		gcs.SetServiceAccountCredentials(presentSecretReference(in.GCSCredentials.ServiceAccountCredentials))
		res.SetGcsCredentials(gcs)
	}

	return res
}
