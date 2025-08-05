package metadata

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/manteia/zhulong/biz/model/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbConn *gorm.DB
	ms     MetadataServiceInterface
)

// TestMain a special function that is called before any tests are run
func TestMain(m *testing.M) {
	// Load .env file for tests
	if err := godotenv.Load("/home/manteia/workspace/zhulong/config/.env"); err != nil {
		fmt.Println("Warning: Error loading .env file for tests:", err)
	}

	// Set up the database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_USER", "postgres"),
		getEnv("POSTGRES_PASSWORD", "postgres"),
		getEnv("POSTGRES_DBNAME", "zhulong_test"),
		getEnv("POSTGRES_PORT", "5432"),
	)
	var err error
	dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Drop and recreate table for a clean test environment
	dbConn.Migrator().DropTable(&db.VideoMetadata{})
	dbConn.AutoMigrate(&db.VideoMetadata{})

	// Set up the metadata service
	ms, err = NewMetadataService(dbConn)
	if err != nil {
		panic(fmt.Sprintf("failed to create metadata service: %v", err))
	}

	// Run the tests
	code := m.Run()

	// Clean up the database
	dbConn.Migrator().DropTable(&db.VideoMetadata{})

	// Exit
	os.Exit(code)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// TestMetadataService_SaveAndGetMetadata tests saving and retrieving metadata
func TestMetadataService_SaveAndGetMetadata(t *testing.T) {
	// Create a new metadata object
	metadata := &FileMetadata{
		FileID:      "test-file-id",
		BucketName:  "test-bucket",
		ObjectName:  "test-object-name",
		FileName:    "test-file-name",
		Title:       "test-title",
		Description: "test-description",
		ContentType: "video/mp4",
		FileSize:    12345,
		Duration:    60,
		Resolution:  "1920x1080",
		Thumbnail:   "test-thumbnail",
		Tags:        []string{"tag1", "tag2"},
		CreatedBy:   "test-user",
	}

	// Save the metadata
	err := ms.SaveMetadata(context.Background(), metadata)
	require.NoError(t, err)

	// Get the metadata
	retrievedMetadata, err := ms.GetMetadata(context.Background(), "test-file-id")
	require.NoError(t, err)

	// Assert that the retrieved metadata is correct
	assert.Equal(t, metadata.FileID, retrievedMetadata.FileID)
	assert.Equal(t, metadata.BucketName, retrievedMetadata.BucketName)
	assert.Equal(t, metadata.ObjectName, retrievedMetadata.ObjectName)
	assert.Equal(t, metadata.FileName, retrievedMetadata.FileName)
	assert.Equal(t, metadata.Title, retrievedMetadata.Title)
	assert.Equal(t, metadata.Description, retrievedMetadata.Description)
	assert.Equal(t, metadata.ContentType, retrievedMetadata.ContentType)
	assert.Equal(t, metadata.FileSize, retrievedMetadata.FileSize)
	assert.Equal(t, metadata.Duration, retrievedMetadata.Duration)
	assert.Equal(t, metadata.Resolution, retrievedMetadata.Resolution)
	assert.Equal(t, metadata.Thumbnail, retrievedMetadata.Thumbnail)
	assert.Equal(t, metadata.Tags, retrievedMetadata.Tags)
	assert.Equal(t, metadata.CreatedBy, retrievedMetadata.CreatedBy)
}

// TestMetadataService_DeleteMetadata tests deleting metadata
func TestMetadataService_DeleteMetadata(t *testing.T) {
	// Create a new metadata object
	metadata := &FileMetadata{
		FileID:    "test-delete-file-id",
		Title:     "test-delete-title",
		CreatedBy: "test-user",
	}

	// Save the metadata
	err := ms.SaveMetadata(context.Background(), metadata)
	require.NoError(t, err)

	// Delete the metadata
	err = ms.DeleteMetadata(context.Background(), "test-delete-file-id")
	require.NoError(t, err)

	// Try to get the deleted metadata
	_, err = ms.GetMetadata(context.Background(), "test-delete-file-id")
	assert.Error(t, err)
}

// TestMetadataService_ListMetadata tests listing metadata
func TestMetadataService_ListMetadata(t *testing.T) {
	// Clear table before test to ensure clean state
	dbConn.Exec("DELETE FROM video_metadata")

	// Create some metadata objects
	for i := 0; i < 5; i++ {
		metadata := &FileMetadata{
			FileID:    fmt.Sprintf("test-list-file-id-%d", i),
			Title:     fmt.Sprintf("test-list-title-%d", i),
			CreatedBy: "test-user",
		}
		err := ms.SaveMetadata(context.Background(), metadata)
		require.NoError(t, err)
	}

	// List the metadata
	listResp, err := ms.ListMetadata(context.Background(), &ListMetadataRequest{
		Offset: 0,
		Limit:  10,
		SortBy: "created_at",
		Order:  "desc",
	})
	require.NoError(t, err)

	// Assert that the list is correct
	assert.Equal(t, 5, listResp.Total)
	assert.Len(t, listResp.Items, 5)
}
