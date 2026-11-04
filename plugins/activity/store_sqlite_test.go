package activity_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/magic-lib/go-plat-utils/plugins/activity"
)

// Migrate 建出的表要能被 Save/List 正常读写
func TestSQLActivityStoreMigrateSaveList(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	store := &activity.SQLActivityStore{DB: db}
	if err = store.Migrate(ctx, "activity_store_test"); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	act := &activity.Activity{Id: "a1", ActivityType: "test_type", ArgTemplate: "{{a}}"}
	if err = store.Save(ctx, act); err != nil {
		t.Fatalf("Save(insert): %v", err)
	}
	act.ArgTemplate = "{{b}}"
	if err = store.Save(ctx, act); err != nil {
		t.Fatalf("Save(update): %v", err)
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Id != "a1" || list[0].ArgTemplate != "{{b}}" {
		t.Fatalf("List = %+v", list)
	}
}
