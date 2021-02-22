package mvcc

type revision struct {
	// main is the main revision of a set of changes that happen atomically.
	// 每一次事务id(transaction_id)
	main int64

	// sub is the sub revision of a change in a set of changes that happen
	// atomically. Each change has different increasing sub revision in that set.
	// 每一次事务内每一个操作id(sub_id)
	sub int64
}

func (a revision) GreaterThan(b revision) bool {
	if a.main > b.main {
		return true
	}
	if a.main < b.main {
		return false
	}
	return a.sub > b.sub
}
