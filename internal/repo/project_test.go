package repo

import (
	"baas-api/config"
	"log"
	"os"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	db *gorm.DB
	pp ProjectRepository
	ep EntityRepository
)

func TestMain(m *testing.M) {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ////////// Init Gorm Database //////////
	db, err = gorm.Open(postgres.Open(config.Database.URL), &gorm.Config{
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

	cache := cache.New(15*time.Minute, 20*time.Minute)

	pp = NewProjectRepository(db)
	ep = NewEntityRepository(db, cache)

	os.Exit(m.Run())
}

var (
	NewProjectID  string
	NewProejctRef string
)

func TestProjectRepository_ByID(t *testing.T) {
	projectEntity, err := ep.GetByChineseName(t.Context(), "專案")
	if err != nil {
		t.Fatalf("Failed to get project entity: %v", err)
	}

	projectID, projectRef, err := pp.Create(t.Context(), "Test Project", nil, projectEntity.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	NewProjectID = *projectID
	NewProejctRef = *projectRef
	t.Logf("Created project with ID: %s", NewProjectID)

	_, err = pp.FindByID(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to get project by ID: %v", err)
	}

	err = pp.DeleteByIDSoft(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	err = pp.DeleteByID(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to permanently delete project: %v", err)
	}
}

func TestProjectRepository_ByRef(t *testing.T) {
	projectEntity, err := ep.GetByChineseName(t.Context(), "專案")
	if err != nil {
		t.Fatalf("Failed to get project entity: %v", err)
	}

	projectID, projectRef, err := pp.Create(t.Context(), "Test Project", nil, projectEntity.ID, nil)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	NewProjectID = *projectID
	NewProejctRef = *projectRef
	t.Logf("Created project with ID: %s and Ref: %s", NewProjectID, NewProejctRef)

	_, err = pp.FindByRef(t.Context(), NewProejctRef)
	if err != nil {
		t.Fatalf("Failed to get project by Ref: %v", err)
	}

	err = pp.DeleteByIDSoft(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	err = pp.DeleteByID(t.Context(), NewProjectID)
	if err != nil {
		t.Fatalf("Failed to permanently delete project: %v", err)
	}
}
