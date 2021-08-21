package proto

import (
	"errors"
)

// http response error code and error message definitions
const (
	ErrCodeSuccess = iota
	ErrCodeInternalError
	ErrCodeParamError
	ErrCodeInvalidCfg
	ErrCodePersistenceByRaft
	ErrCodeMarshalData
	ErrCodeUnmarshalData
	ErrCodeVolNotExists
	ErrCodeBucketNotExists
	ErrCodeMetaPartitionNotExists
	ErrCodeMetaNodeNotExists
	ErrCodeDuplicateVol
	ErrCodeActiveMetaNodesTooLess
	ErrCodeInvalidMpStart
	ErrCodeReshuffleArray
	ErrCodeMissingReplica
	ErrCodeHasOneMissingReplica
	ErrCodeNoMetaNodeToWrite
	ErrCodeCannotBeOffLine
	ErrCodeNoNodeSetToCreateMetaPartition
	ErrCodeNoMetaNodeToCreateMetaPartition
	ErrCodeIllegalMetaReplica
	ErrCodeNoEnoughReplica
	ErrCodeNoLeader
	ErrCodeVolAuthKeyNotMatch
	ErrCodeAuthKeyStoreError
	ErrCodeAuthAPIAccessGenRespError
	ErrCodeAuthRaftNodeGenRespError
	ErrCodeAuthOSCapsOpGenRespError
	ErrCodeAuthReqRedirectError
	ErrCodeAccessKeyNotExists
	ErrCodeInvalidTicket
	ErrCodeExpiredTicket
	ErrCodeMasterAPIGenRespError
	ErrCodeVolMountClientNotExists
)

//err
var (
	ErrSuc                     = errors.New("success")
	ErrInternalError           = errors.New("internal error")
	ErrParamError              = errors.New("parameter error")
	ErrInvalidCfg              = errors.New("bad configuration file")
	ErrPersistenceByRaft       = errors.New("persistence by raft occurred error")
	ErrMarshalData             = errors.New("marshal data error")
	ErrUnmarshalData           = errors.New("unmarshal data error")
	ErrVolNotExists            = errors.New("vol not exists")
	ErrVolMountClientNotExists = errors.New("vol mount client not exists")
	ErrBucketNotExists         = errors.New("bucket not exists")
	ErrMetaPartitionNotExists  = errors.New("meta partition not exists")
	ErrMetaNodeNotExists       = errors.New("meta node not exists")
	ErrDuplicateVol            = errors.New("duplicate vol")
	ErrActiveMetaNodesTooLess  = errors.New("no enough active meta node")
	ErrInvalidMpStart          = errors.New("invalid meta partition start value")
	ErrReshuffleArray          = errors.New("the array to be reshuffled is nil")
	ErrMissingReplica          = errors.New("a missing data replica is found")
	ErrHasOneMissingReplica    = errors.New("there is a missing replica")
	ErrNoMetaNodeToWrite       = errors.New("No meta node available for creating a meta partition")

	ErrCannotBeOffLine                 = errors.New("cannot take the data replica offline")
	ErrNoNodeSetToCreateMetaPartition  = errors.New("no node set available for creating a meta partition")
	ErrNoMetaNodeToCreateMetaPartition = errors.New("no enough meta nodes for creating a meta partition")
	ErrIllegalMetaReplica              = errors.New("illegal meta replica")
	ErrNoEnoughReplica                 = errors.New("no enough replicas")
	ErrNoLeader                        = errors.New("no leader")
	ErrVolAuthKeyNotMatch              = errors.New("client and server auth key do not match")
	ErrAuthKeyStoreError               = errors.New("auth keystore error")
	ErrAuthAPIAccessGenRespError       = errors.New("auth API access response error")
	ErrAuthOSCapsOpGenRespError        = errors.New("auth Object Storage Node API response error")
	ErrKeyNotExists                    = errors.New("key not exists")
	ErrDuplicateKey                    = errors.New("duplicate key")
	ErrAccessKeyNotExists              = errors.New("access key not exists")
	ErrInvalidTicket                   = errors.New("invalid ticket")
	ErrExpiredTicket                   = errors.New("expired ticket")
	ErrMasterAPIGenRespError           = errors.New("master API generate response error")
)

// Err2CodeMap error map to code
var Err2CodeMap = map[error]int32{
	ErrSuc:                             ErrCodeSuccess,
	ErrInternalError:                   ErrCodeInternalError,
	ErrParamError:                      ErrCodeParamError,
	ErrInvalidCfg:                      ErrCodeInvalidCfg,
	ErrPersistenceByRaft:               ErrCodePersistenceByRaft,
	ErrMarshalData:                     ErrCodeMarshalData,
	ErrUnmarshalData:                   ErrCodeUnmarshalData,
	ErrVolNotExists:                    ErrCodeVolNotExists,
	ErrBucketNotExists:                 ErrCodeBucketNotExists,
	ErrMetaPartitionNotExists:          ErrCodeMetaPartitionNotExists,
	ErrMetaNodeNotExists:               ErrCodeMetaNodeNotExists,
	ErrDuplicateVol:                    ErrCodeDuplicateVol,
	ErrActiveMetaNodesTooLess:          ErrCodeActiveMetaNodesTooLess,
	ErrInvalidMpStart:                  ErrCodeInvalidMpStart,
	ErrReshuffleArray:                  ErrCodeReshuffleArray,
	ErrMissingReplica:                  ErrCodeMissingReplica,
	ErrHasOneMissingReplica:            ErrCodeHasOneMissingReplica,
	ErrNoMetaNodeToWrite:               ErrCodeNoMetaNodeToWrite,
	ErrCannotBeOffLine:                 ErrCodeCannotBeOffLine,
	ErrNoNodeSetToCreateMetaPartition:  ErrCodeNoNodeSetToCreateMetaPartition,
	ErrNoMetaNodeToCreateMetaPartition: ErrCodeNoMetaNodeToCreateMetaPartition,
	ErrIllegalMetaReplica:              ErrCodeIllegalMetaReplica,
	ErrNoEnoughReplica:                 ErrCodeNoEnoughReplica,
	ErrNoLeader:                        ErrCodeNoLeader,
	ErrVolAuthKeyNotMatch:              ErrCodeVolAuthKeyNotMatch,
	ErrAuthKeyStoreError:               ErrCodeAuthKeyStoreError,
	ErrAuthAPIAccessGenRespError:       ErrCodeAuthAPIAccessGenRespError,
	ErrAuthOSCapsOpGenRespError:        ErrCodeAuthOSCapsOpGenRespError,
	ErrAccessKeyNotExists:              ErrCodeAccessKeyNotExists,
	ErrInvalidTicket:                   ErrCodeInvalidTicket,
	ErrExpiredTicket:                   ErrCodeExpiredTicket,
	ErrMasterAPIGenRespError:           ErrCodeMasterAPIGenRespError,
	ErrVolMountClientNotExists:         ErrCodeVolMountClientNotExists,
}
