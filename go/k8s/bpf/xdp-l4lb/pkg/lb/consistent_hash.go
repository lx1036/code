package lb

const (
	DefaultChRingSize = 65537
)

/**
 * struct which describes backend, each backend would have unique number,
 * weight (the measurment of how often we would see this endpoint
 * on CH ring) and hash value, which will be used as a seed value
 * (it should be unique value per endpoint for CH to work as expected)
 */
type Endpoint struct {
	num    uint32
	weight uint32
	hash   uint64
}

func compareEndpoints(a, b Endpoint) bool {
	return a.hash < b.hash
}

type ConsistentHash interface {
	generateHashRing(endpoints []Endpoint, ringSize uint32) []int
}
