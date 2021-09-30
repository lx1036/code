package backend

var (
	testBucketName = []byte("test")

	keyBucketName  = []byte("key")
	metaBucketName = []byte("meta")
)

var (
	Test = bucket{id: 100, name: testBucketName, safeRangeBucket: false}
	Key  = bucket{id: 1, name: keyBucketName, safeRangeBucket: true}
	Meta = bucket{id: 2, name: metaBucketName, safeRangeBucket: false}
)

type bucket struct {
	id              BucketID
	name            []byte
	safeRangeBucket bool
}

func (b bucket) ID() BucketID {
	return b.id
}

func (b bucket) Name() []byte {
	return b.name
}

func (b bucket) String() string {
	return string(b.Name())
}

func (b bucket) IsSafeRangeBucket() bool {
	return b.safeRangeBucket
}
