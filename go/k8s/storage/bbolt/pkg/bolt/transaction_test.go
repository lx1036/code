package bolt

import "testing"

func TestTxCommit(test *testing.T) {
	db := MustOpenDB()
	defer db.MustClose()
	tx, err := db.Begin(true)
	if err != nil {
		test.Fatal(err)
	}

	if _, err := tx.CreateBucket([]byte("foo")); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		test.Fatal(err)
	}

	if err := tx.Commit(); err != ErrTxClosed {
		test.Fatalf("unexpected error: %s", err)
	}
}
