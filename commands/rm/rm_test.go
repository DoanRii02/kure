package rm

import (
	"bytes"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestRm(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	name := "test"
	createEntries(t, db, name)

	buf := bytes.NewBufferString("y")
	cmd := NewCmd(db, buf)
	cmd.SetArgs([]string{name})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed removing the entry: %v", err)
	}

	// Check if the entry was removed successfully
	if _, err := entry.Get(db, name); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
}

func TestRmDir(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	// Create the entries inside a folder to remove them
	names := []string{"test/entry1", "test/entry2"}
	createEntries(t, db, names...)

	buf := bytes.NewBufferString("y")
	cmd := NewCmd(db, buf)
	cmd.SetArgs([]string{"test/"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed removing the entry: %v", err)
	}

	// Check if the entries were removed successfully
	if _, err := entry.Get(db, names[0]); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
	if _, err := entry.Get(db, names[1]); err == nil {
		t.Error("Expected Get() to fail but it didn't")
	}
}

func TestRmAbort(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	name := "may the force be with you"
	createEntries(t, db, name)

	buf := bytes.NewBufferString("n") // Abort operation
	cmd := NewCmd(db, buf)
	cmd.SetArgs([]string{name})

	if err := cmd.Execute(); err != nil {
		t.Errorf("Rm() failed: %v", err)
	}
}

func TestRmErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntries(t, db, "fail.txt")

	cases := []struct {
		desc  string
		name  string
		input string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc:  "Not exists",
			name:  "non-existent",
			input: "y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			if err := cmd.Execute(); err == nil {
				t.Fatal("Expected Rm() to fail but it didn't")
			}
		})
	}
}

func createEntries(t *testing.T, db *bolt.DB, names ...string) {
	t.Helper()

	for _, n := range names {
		if err := entry.Create(db, &pb.Entry{Name: n}); err != nil {
			t.Fatal(err)
		}
	}
}
