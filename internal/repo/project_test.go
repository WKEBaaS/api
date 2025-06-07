package repo

import (
	"baas-api/internal/configs"
	"log"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	db   *gorm.DB
	err  error
	repo ProjectRepository
)

func TestMain(m *testing.M) {
	config := configs.LoadConfig()

	// ////////// Init Gorm Database //////////
	db, err = gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			// SlowThreshold: time.Second,
			LogLevel: logger.Info,
			Colorful: true,
		}),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
			NoLowerCase:   false,
		},
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	repo = NewProjectRepository(db)

	os.Exit(m.Run())
}

var (
	NewProjectID  string
	NewProejctRef string
)

func TestProjectRepository_ByID(t *testing.T) {
	projectID, projectRef, err := repo.Create(t.Context(), "Test Project")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	NewProjectID = *projectID
	NewProejctRef = *projectRef
	t.Logf("Created project with ID: %s", NewProjectID)

	_, err = repo.GetByID(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to get project by ID: %v", err)
	}

	err = repo.DeleteByIDSoft(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	err = repo.DeleteByIDPermanently(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to permanently delete project: %v", err)
	}
}

func TestProjectRepository_ByRef(t *testing.T) {
	projectID, projectRef, err := repo.Create(t.Context(), "Test Project")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	NewProjectID = *projectID
	NewProejctRef = *projectRef
	t.Logf("Created project with ID: %s and Ref: %s", NewProjectID, NewProejctRef)

	_, err = repo.GetByRef(t.Context(), NewProejctRef)
	if err != nil {
		t.Fatalf("Failed to get project by Ref: %v", err)
	}

	err = repo.DeleteByIDSoft(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	err = repo.DeleteByIDPermanently(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to permanently delete project: %v", err)
	}
}
