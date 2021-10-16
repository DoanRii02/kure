package entry

import (
	"testing"

	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"
	"github.com/GGP1/kure/pb"

	"github.com/awnumar/memguard"
	bolt "go.etcd.io/bbolt"
)

func TestEntry(t *testing.T) {
	db := setContext(t)

	name := "test"
	e := &pb.Entry{
		Name:     name,
		Username: "testing",
		URL:      "golang.org",
		Expires:  "Never",
		Notes:    "",
	}

	t.Run("Create", create(db, e))
	t.Run("Get", get(db, name))
	t.Run("List", list(db))
	t.Run("List names", listNames(db))
	t.Run("Remove", remove(db, name))
	t.Run("Update", update(db))
}

func create(db *bolt.DB, e *pb.Entry) func(*testing.T) {
	return func(t *testing.T) {
		if err := Create(db, e); err != nil {
			t.Error(err)
		}
	}
}

func get(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Get(db, name)
		if err != nil {
			t.Error(err)
		}

		// They aren't DeepEqual
		if got.Name != name {
			t.Errorf("Expected %s, got %s", name, got.Name)
		}
	}
}

func list(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		got, err := List(db)
		if err != nil {
			t.Error(err)
		}

		if len(got) == 0 {
			t.Error("Expected one or more entries, got 0")
		}
	}
}

func listNames(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		entries, err := ListNames(db)
		if err != nil {
			t.Error(err)
		}

		expected := "test"
		got := entries[0]

		if got != expected {
			t.Errorf("Expected %s, got %s", expected, got)
		}

		if len(entries) == 0 {
			t.Error("Expected one or more entries, got 0")
		}
	}
}

func remove(db *bolt.DB, name string) func(*testing.T) {
	return func(t *testing.T) {
		if err := Remove(db, name); err != nil {
			t.Error(err)
		}
	}
}

func update(db *bolt.DB) func(*testing.T) {
	return func(t *testing.T) {
		oldEntry := &pb.Entry{Name: "old"}
		if err := Create(db, oldEntry); err != nil {
			t.Fatal(err)
		}

		newEntry := &pb.Entry{Name: "new"}
		if err := Update(db, oldEntry.Name, newEntry); err != nil {
			t.Fatal(err)
		}

		if _, err := Get(db, newEntry.Name); err != nil {
			t.Error(err)
		}
	}
}

func TestListExpired(t *testing.T) {
	db := setContext(t)

	e := &pb.Entry{
		Name:    "test expired",
		Expires: "Mon, 10 Jan 2020 15:04:05 -0700",
	}

	t.Run("Create", create(db, e))

	if _, err := List(db); err != nil {
		t.Errorf("List() failed: %v", err)
	}
}

func TestCreateErrors(t *testing.T) {
	db := setContext(t)

	cases := []struct {
		desc string
		name string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc: "Null characters",
			name: string('\x00'),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := Create(db, &pb.Entry{Name: tc.name}); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestGetErrors(t *testing.T) {
	db := setContext(t)

	if _, err := Get(db, "non-existent"); err == nil {
		t.Error("Expected an error, got nil")
	}
}

func TestCryptErrors(t *testing.T) {
	db := setContext(t)

	name := "test decrypt error"

	e := &pb.Entry{Name: name, Expires: "Never"}
	if err := Create(db, e); err != nil {
		t.Fatal(err)
	}

	// Try to get the entry with other password
	config.Set("auth.password", memguard.NewEnclave([]byte("invalid")))

	if _, err := Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestProtoErrors(t *testing.T) {
	db := setContext(t)

	name := "unformatted"
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbutil.EntryBucket))
		buf := make([]byte, 64)
		encBuf, _ := crypt.Encrypt(buf)
		return b.Put([]byte(name), encBuf)
	})
	if err != nil {
		t.Fatalf("Failed writing invalid type: %v", err)
	}

	if _, err := Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := List(db); err == nil {
		t.Error("Expected List() to fail but it didn't")
	}
}

func TestKeyError(t *testing.T) {
	db := setContext(t)

	if err := Create(db, &pb.Entry{Name: ""}); err == nil {
		t.Error("Create() didn't fail")
	}
}

func setContext(t testing.TB) *bolt.DB {
	return dbutil.SetContext(t, "../testdata/database", dbutil.EntryBucket)
}
