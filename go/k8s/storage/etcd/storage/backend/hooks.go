package backend

// Hooks allow to add additional logic executed during transaction lifetime.
type Hooks interface {
	// OnPreCommitUnsafe is executed before Commit of transactions.
	// The given transaction is already locked.
	OnPreCommitUnsafe(tx BatchTx)
}
