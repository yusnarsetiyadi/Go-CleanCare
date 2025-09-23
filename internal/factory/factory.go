package factory

import (
	"iss_cleancare/internal/repository"
	"iss_cleancare/pkg/constant"
	"iss_cleancare/pkg/database"
	"iss_cleancare/pkg/gdrive"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/drive/v3"
	"gorm.io/gorm"
)

type Factory struct {
	Db *gorm.DB

	DbRedis *redis.Client

	GDrive GoogleDrive

	Repository_initiated
}

type Repository_initiated struct {
	UserRepository repository.User
	RoleRepository repository.Role
}

type GoogleDrive struct {
	Service   *drive.Service
	FolderIss *drive.File
}

func NewFactory() *Factory {
	f := &Factory{}
	f.SetupDb()
	f.SetupDbRedis()
	f.SetupGoogleDrive()
	f.SetupRepository()
	return f
}

func (f *Factory) SetupDb() {
	db, err := database.Connection("MYSQL")
	if err != nil {
		panic("Failed setup db, connection is undefined")
	}

	// sqlDB, err := db.DB()
	// if err != nil {
	// 	panic(err)
	// }
	// sqlDB.SetMaxIdleConns(5)
	// sqlDB.SetMaxOpenConns(20)
	// sqlDB.SetConnMaxLifetime(time.Hour)

	f.Db = db
}

func (f *Factory) SetupDbRedis() {
	dbRedis := database.InitRedis()
	f.DbRedis = dbRedis
}

func (f *Factory) SetupGoogleDrive() {
	service, err := gdrive.InitService()
	if err != nil {
		panic("Failed setup gdrive, connection is undefined")
	}
	folderIss, err := gdrive.InitFolder(service, constant.DRIVE_FOLDER, "root")
	if err != nil {
		logrus.Infof("Failed setup folder %s, cause: %s", constant.DRIVE_FOLDER, err.Error())
	}
	f.GDrive.Service = service
	f.GDrive.FolderIss = folderIss
}

func (f *Factory) SetupRepository() {
	if f.Db == nil {
		panic("Failed setup repository, db is undefined")
	}

	f.UserRepository = repository.NewUser(f.Db)
	f.RoleRepository = repository.NewRole(f.Db)
}
