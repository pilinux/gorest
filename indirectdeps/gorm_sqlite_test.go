package indirectdeps_test

// SQLite path: database.InitDB (driver sqlite3) -> gorm.io/driver/sqlite ->
// go-sqlite3 (cgo); gorm names tables and columns with inflection and
// x/text/cases.
//
// Indirect modules:
//   - github.com/mattn/go-sqlite3
//   - github.com/jinzhu/inflection
//   - golang.org/x/text
//
//	Function                  Purpose
//	------------------------  -------
//	initSQLite                config with driver sqlite3 and a temp DB file, then InitDB
//	TestSQLiteMigrateAndCRUD  migrate.StartMigration creates auths, two_fas, two_fa_backups and
//	                          temp_emails (inflection plurals) with email_hash, email_cipher,
//	                          verify_email and deleted_at; create fills AuthID and CreatedAt,
//	                          lookup by email_hash works, delete is soft (hidden, still counted
//	                          by Unscoped), and sqlite_version() runs
//	TestGormJoinTableNaming   AutoMigrate of example User/Post/Hobby creates user_hobbies with
//	                          user_user_id and hobby_hobby_id, and users.id_auth (gorm builds
//	                          both names with x/text/cases); a User with a Hobby saves and
//	                          preloads back

import (
	"path/filepath"
	"testing"

	gconfig "github.com/pilinux/gorest/config"
	gdb "github.com/pilinux/gorest/database"
	gmigrate "github.com/pilinux/gorest/database/migrate"
	gmodel "github.com/pilinux/gorest/database/model"
	exmodel "github.com/pilinux/gorest/example/database/model"

	"gorm.io/gorm"
)

func initSQLite(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := setupConfig(t)
	cfg.Database.RDBMS.Env.Driver = "sqlite3"
	cfg.Database.RDBMS.Access.DbName = filepath.Join(t.TempDir(), "gorest.db")
	cfg.Database.RDBMS.Log.LogLevel = 1
	t.Cleanup(func() { _ = gdb.CloseSQL() })

	db := gdb.InitDB()
	if db.Error != nil {
		t.Fatalf("InitDB sqlite3: %v", db.Error)
	}
	return db
}

// TestSQLiteMigrateAndCRUD covers mattn/go-sqlite3 and jinzhu/inflection:
// migrate.StartMigration relies on pluralized table names, and handlers rely
// on create, lookup by hash and soft delete.
func TestSQLiteMigrateAndCRUD(t *testing.T) {
	logIndirect(t)
	db := initSQLite(t)
	if err := gmigrate.StartMigration(*gconfig.GetConfig()); err != nil {
		t.Fatalf("StartMigration: %v", err)
	}

	for _, table := range []string{"auths", "two_fas", "two_fa_backups", "temp_emails"} {
		if !db.Migrator().HasTable(table) {
			t.Errorf("table %q missing after migration", table)
		}
	}
	for _, col := range []string{"email_hash", "email_cipher", "verify_email", "deleted_at"} {
		if !db.Migrator().HasColumn(&gmodel.Auth{}, col) {
			t.Errorf("auths.%s missing", col)
		}
	}

	auth := gmodel.Auth{EmailHash: "hash-1", Password: "argon2-hash", VerifyEmail: 1} // ggignore
	if err := db.Create(&auth).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	if auth.AuthID == 0 || auth.CreatedAt.IsZero() {
		t.Fatalf("primary key / timestamps not filled: %+v", auth)
	}

	var found gmodel.Auth
	if err := db.Where("email_hash = ?", "hash-1").First(&found).Error; err != nil {
		t.Fatalf("lookup by email_hash: %v", err)
	}
	if found.AuthID != auth.AuthID {
		t.Errorf("found AuthID %d, want %d", found.AuthID, auth.AuthID)
	}

	if err := db.Delete(&found).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if err := db.Where("email_hash = ?", "hash-1").First(&gmodel.Auth{}).Error; err != gorm.ErrRecordNotFound {
		t.Errorf("soft-deleted row still visible: %v", err)
	}
	var count int64
	db.Unscoped().Model(&gmodel.Auth{}).Where("email_hash = ?", "hash-1").Count(&count)
	if count != 1 {
		t.Errorf("unscoped count = %d, want 1", count)
	}

	var version string
	if err := db.Raw("select sqlite_version()").Scan(&version).Error; err != nil || version == "" {
		t.Errorf("sqlite_version: %q, %v", version, err)
	}
}

// TestGormJoinTableNaming covers golang.org/x/text/cases: gorm title-cases
// the schema name to build many2many join columns for the example models.
func TestGormJoinTableNaming(t *testing.T) {
	logIndirect(t)
	db := initSQLite(t)
	if err := db.AutoMigrate(&exmodel.User{}, &exmodel.Post{}, &exmodel.Hobby{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	for _, col := range []string{"user_user_id", "hobby_hobby_id"} {
		if !db.Migrator().HasColumn("user_hobbies", col) {
			t.Errorf("user_hobbies.%s missing", col)
		}
	}
	if !db.Migrator().HasColumn(&exmodel.User{}, "id_auth") {
		t.Error("users.id_auth missing (initialism naming)")
	}

	user := exmodel.User{FirstName: "Ada", Hobbies: []exmodel.Hobby{{Hobby: "chess"}}}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create with association: %v", err)
	}
	var loaded exmodel.User
	if err := db.Preload("Hobbies").First(&loaded, user.UserID).Error; err != nil {
		t.Fatalf("preload: %v", err)
	}
	if len(loaded.Hobbies) != 1 || loaded.Hobbies[0].Hobby != "chess" {
		t.Errorf("hobbies = %+v", loaded.Hobbies)
	}
}
