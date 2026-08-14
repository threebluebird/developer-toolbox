package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"developer-toolbox/backend/models"
)

func TestFileStorageRejectsPathTraversal(t *testing.T) {
	store := NewFileStorage(t.TempDir())
	if err := store.Set("../outside.json", []byte("x")); err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

func TestRepositoriesPersistAndLimitData(t *testing.T) {
	store := NewFileStorage(t.TempDir())
	favorites := NewFavoriteRepository(store)
	if err := favorites.Add("json"); err != nil {
		t.Fatal(err)
	}
	if err := favorites.Add("json"); err != nil {
		t.Fatal(err)
	}
	items, err := favorites.List()
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected favorites: %#v, %v", items, err)
	}

	history := NewHistoryRepository(store)
	for i := 0; i < 55; i++ {
		if err := history.Add(models.History{ToolID: "json", UsedAt: string(rune(i))}); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := history.List()
	if err != nil || len(entries) != 50 {
		t.Fatalf("unexpected history length: %d, %v", len(entries), err)
	}
}

func TestConfigRepositoryMissingFile(t *testing.T) {
	repo := NewConfigRepository(NewFileStorage(t.TempDir()))
	settings, err := repo.Load()
	if err != nil || settings != (models.Settings{}) {
		t.Fatalf("unexpected result: %#v, %v", settings, err)
	}
}

func TestDeleteMissingFileReturnsNotExist(t *testing.T) {
	store := NewFileStorage(t.TempDir())
	err := store.Delete("missing")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected not-exist, got %v", err)
	}
	_ = filepath.Separator
}

func TestWorkflowRepositoriesSupportV2AndLegacyEntries(t *testing.T) {
	store := NewFileStorage(t.TempDir())
	favorites := NewFavoriteRepository(store)
	if err := favorites.Add("json"); err != nil {
		t.Fatal(err)
	}
	item := models.Favorite{ID: "saved-1", ToolID: "http", Kind: "request", Name: "Users", Payload: map[string]any{"url": "https://example.test"}}
	if err := favorites.AddItem(item); err != nil {
		t.Fatal(err)
	}
	items, err := favorites.List()
	if err != nil || len(items) != 2 || items[0].ToolID != "json" || items[1].ID != "saved-1" {
		t.Fatalf("unexpected favorites: %#v, %v", items, err)
	}
	if err := favorites.RemoveItem("saved-1"); err != nil {
		t.Fatal(err)
	}

	history := NewHistoryRepository(store)
	if err := history.Add(models.History{ID: "history-1", ToolID: "json", UsedAt: "now", Input: map[string]any{"input": "{}"}, Output: "valid"}); err != nil {
		t.Fatal(err)
	}
	entries, err := history.List()
	if err != nil || len(entries) != 1 || entries[0].Input["input"] != "{}" {
		t.Fatalf("unexpected history: %#v, %v", entries, err)
	}
	if err := history.Clear(); err != nil {
		t.Fatal(err)
	}
	entries, _ = history.List()
	if len(entries) != 0 {
		t.Fatalf("history was not cleared: %#v", entries)
	}
}
