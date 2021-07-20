package master

const (
	opSyncAddMetaNode          uint32 = 0x01
	opSyncAddVol               uint32 = 0x02
	opSyncAddMetaPartition     uint32 = 0x03
	opSyncUpdateMetaPartition  uint32 = 0x04
	opSyncDeleteMetaNode       uint32 = 0x05
	opSyncAllocMetaPartitionID uint32 = 0x06
	opSyncAllocCommonID        uint32 = 0x07
	opSyncPutCluster           uint32 = 0x08
	opSyncUpdateVol            uint32 = 0x09
	opSyncDeleteVol            uint32 = 0x0A
	opSyncDeleteMetaPartition  uint32 = 0x0B
	opSyncAddNodeSet           uint32 = 0x0C
	opSyncUpdateNodeSet        uint32 = 0x0D
	opSyncBatchPut             uint32 = 0x0E
	opSyncAddBucket            uint32 = 0x0F
	opSyncUpdateBucket         uint32 = 0x10
	opSyncDeleteBucket         uint32 = 0x11
	opSyncAddVolMountClient    uint32 = 0x12
	opSyncUpdateVolMountClient uint32 = 0x13
	opSyncDeleteVolMountClient uint32 = 0x14
)
