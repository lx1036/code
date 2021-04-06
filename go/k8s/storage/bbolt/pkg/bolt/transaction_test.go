package bolt

import "testing"

func TestTxCommit(test *testing.T) {
	db := MustOpenDB()
	defer db.MustClose()
	tx, err := db.Begin(true)
	if err != nil {
		test.Fatal(err)
	}

}
