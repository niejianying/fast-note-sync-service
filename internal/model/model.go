package model

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, key string) error {
	if db == nil {
		return nil
	}
	switch key {

	case "AuthToken":
		return db.AutoMigrate(AuthToken{})

	case "AuthTokenLog":
		return db.AutoMigrate(AuthTokenLog{})

	case "BackupConfig":
		return db.AutoMigrate(BackupConfig{})

	case "BackupHistory":
		return db.AutoMigrate(BackupHistory{})

	case "File":
		return db.AutoMigrate(File{})

	case "Folder":
		return db.AutoMigrate(Folder{})

	case "FriendRelationship":
		return db.AutoMigrate(FriendRelationship{})

	case "GitSyncConfig":
		return db.AutoMigrate(GitSyncConfig{})

	case "GitSyncHistory":
		return db.AutoMigrate(GitSyncHistory{})

	case "Note":
		return db.AutoMigrate(Note{})

	case "NoteHistory":
		return db.AutoMigrate(NoteHistory{})

	case "NoteLink":
		return db.AutoMigrate(NoteLink{})

	case "Setting":
		return db.AutoMigrate(Setting{})

	case "SharedFile":
		return db.AutoMigrate(SharedFile{})

	case "SharedFolder":
		return db.AutoMigrate(SharedFolder{})

	case "SharedNote":
		return db.AutoMigrate(SharedNote{})

	case "SharedVault":
		return db.AutoMigrate(SharedVault{})

	case "Storage":
		return db.AutoMigrate(Storage{})

	case "VaultMember":
		return db.AutoMigrate(VaultMember{})

	case "User":
		return db.AutoMigrate(User{})

	case "UserOIDCIdentity":
		return db.AutoMigrate(UserOIDCIdentity{})

	case "UserShare":
		return db.AutoMigrate(UserShare{})

	case "Vault":
		return db.AutoMigrate(Vault{})

	case "InboxItem":
		return db.AutoMigrate(InboxItem{})
	}
	return nil
}
