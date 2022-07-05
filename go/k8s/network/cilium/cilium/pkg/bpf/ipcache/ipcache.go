package ipcache


// INFO: ipcache BPF map 主要是用来？？？

var (
	// IPIdentityCache caches the mapping of endpoint IPs to their corresponding
	// security identities across the entire cluster in which this instance of
	// Cilium is running.
	IPIdentityCache = NewIPCache()
)
