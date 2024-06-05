package getters

// LinkGetter fetches local link information.
type LinkGetter interface {
    // GetIfNameCached returns the name of an interface (if it exists) by
    // looking it up in a regularly updated cache
    GetIfNameCached(ifIndex int) (string, bool)

    // Name returns the name of an interface, or returns a string
    // containing the ifindex if the link name cannot be determined.
    Name(ifIndex uint32) string
}
