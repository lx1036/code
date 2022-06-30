package endpoint

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/controller"
	"github.com/cilium/cilium/pkg/endpoint/regeneration"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/policy"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/defaults"
)

const (
	// StateWaitingForIdentity is used to set if the endpoint is waiting
	// for an identity from the KVStore.
	StateWaitingForIdentity = string(models.EndpointStateWaitingForIdentity)

	// StateReady specifies if the endpoint is ready to be used.
	StateReady = string(models.EndpointStateReady)

	// StateWaitingToRegenerate specifies when the endpoint needs to be regenerated, but regeneration has not started yet.
	StateWaitingToRegenerate = string(models.EndpointStateWaitingToRegenerate)

	// StateRegenerating specifies when the endpoint is being regenerated.
	StateRegenerating = string(models.EndpointStateRegenerating)

	// StateDisconnecting indicates that the endpoint is being disconnected
	StateDisconnecting = string(models.EndpointStateDisconnecting)

	// StateDisconnected is used to set the endpoint is disconnected.
	StateDisconnected = string(models.EndpointStateDisconnected)

	// StateRestoring is used to set the endpoint is being restored.
	StateRestoring = string(models.EndpointStateRestoring)

	// StateInvalid is used when an endpoint failed during creation due to
	// invalid data.
	StateInvalid = string(models.EndpointStateInvalid)
)

// ReadEPsFromDirNames returns a mapping of endpoint ID to endpoint of endpoints
// from a list of directory names that can possible contain an endpoint.
func ReadEPsFromDirNames(ctx context.Context, owner regeneration.Owner, basePath string, eptsDirNames []string) map[uint16]*Endpoint {
	completeEPDirNames, incompleteEPDirNames := partitionEPDirNamesByRestoreStatus(eptsDirNames)

	// delete incomplete dir
	if len(incompleteEPDirNames) > 0 {
		for _, epDirName := range incompleteEPDirNames {
			scopedLog := log.WithFields(log.Fields{
				logfields.EndpointID: epDirName,
			})
			fullDirName := filepath.Join(basePath, epDirName)
			scopedLog.Warning(fmt.Sprintf("Found incomplete restore directory %s. Removing it...", fullDirName))
			if err := os.RemoveAll(epDirName); err != nil {
				scopedLog.WithError(err).Warn(fmt.Sprintf("Error while removing directory %s. Ignoring it...", fullDirName))
			}
		}
	}

	possibleEPs := map[uint16]*Endpoint{}
	for _, epDirName := range completeEPDirNames {
		epDir := filepath.Join(basePath, epDirName)
		isHost := hasHostObjectFile(epDir)
		// We only check for the old header file. On v1.8+, if the old header
		// file is present, the new header file will be too. When upgrading
		// from pre-1.8, only the old header file is present and we will create
		// the new.
		// We can switch this to use the new header file once v1.8 is the
		// oldest supported version.
		cHeaderFile := filepath.Join(epDir, defaults.OldCHeaderFileName)
		if isHost {
			// Host endpoint doesn't have an old header file so that it's not
			// restored on downgrades.
			cHeaderFile = filepath.Join(epDir, defaults.CHeaderFileName)
		}
		scopedLog := log.WithFields(log.Fields{
			logfields.EndpointID: epDirName,
			logfields.Path:       cHeaderFile,
		})
		strEp, err := GetCiliumVersionString(cHeaderFile)
		if err != nil {
			scopedLog.WithError(err).Warn("Unable to read the C header file")
			continue
		}
		ep, err := parseEndpoint(ctx, owner, strEp)
		if err != nil {
			scopedLog.WithError(err).Warn("Unable to parse the C header file")
			continue
		}

		if _, ok := possibleEPs[ep.ID]; ok {
			// If the endpoint already exists then give priority to the directory
			// that contains an endpoint that didn't fail to be build.
			if strings.HasSuffix(ep.DirectoryPath(), epDirName) {
				possibleEPs[ep.ID] = ep
			}
		} else {
			possibleEPs[ep.ID] = ep
		}
	}

	return possibleEPs
}

// partitionEPDirNamesByRestoreStatus partitions the provided list of directory
// names that can possibly contain an endpoint, into two lists, containing those
// names that represent an incomplete endpoint restore and those that do not.
func partitionEPDirNamesByRestoreStatus(eptsDirNames []string) (complete []string, incomplete []string) {
	dirNames := make(map[string]struct{}, len(eptsDirNames))
	for _, epDirName := range eptsDirNames {
		dirNames[epDirName] = struct{}{}
	}

	incompleteSuffixes := []string{nextDirectorySuffix, nextFailedDirectorySuffix}
	incompleteSet := make(map[string]struct{})

	for _, epDirName := range eptsDirNames {
		for _, suff := range incompleteSuffixes {
			if strings.HasSuffix(epDirName, suff) {
				if _, exists := dirNames[epDirName[:len(epDirName)-len(suff)]]; exists {
					incompleteSet[epDirName] = struct{}{}
				}
			}
		}
	}

	for epDirName := range dirNames {
		if _, exists := incompleteSet[epDirName]; exists {
			incomplete = append(incomplete, epDirName)
		} else {
			complete = append(complete, epDirName)
		}
	}

	return
}

func hasHostObjectFile(epDir string) bool {
	hostObjFilepath := filepath.Join(epDir, defaults.HostObjFileName)
	_, err := os.Stat(hostObjFilepath)
	return err == nil
}

// GetCiliumVersionString returns the first line containing CiliumCHeaderPrefix.
func GetCiliumVersionString(epCHeaderFilePath string) (string, error) {
	f, err := os.Open(epCHeaderFilePath)
	if err != nil {
		return "", err
	}
	br := bufio.NewReader(f)
	defer f.Close()
	for {
		s, err := br.ReadString('\n')
		if err == io.EOF {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if strings.Contains(s, defaults.CiliumCHeaderPrefix) {
			return s, nil
		}
	}
}

// parseEndpoint parses the given strEp which is in the form of:
// common.CiliumCHeaderPrefix + common.Version + ":" + endpointBase64
// Note that the parse'd endpoint's identity is only partially restored. The
// caller must call `SetIdentity()` to make the returned endpoint's identity useful.
func parseEndpoint(ctx context.Context, owner regeneration.Owner, strEp string) (*Endpoint, error) {
	// TODO: Provide a better mechanism to update from old version once we bump
	// TODO: cilium version.
	strEpSlice := strings.Split(strEp, ":")
	if len(strEpSlice) != 2 {
		return nil, fmt.Errorf("invalid format %q. Should contain a single ':'", strEp)
	}
	ep := Endpoint{
		owner: owner,
	}

	if err := parseBase64ToEndpoint(strEpSlice[1], &ep); err != nil {
		return nil, fmt.Errorf("failed to parse restored endpoint: %s", err)
	}

	// Validate the options that were parsed
	//ep.SetDefaultOpts(ep.Options)

	// Initialize fields to values which are non-nil that are not serialized.
	ep.hasBPFProgram = make(chan struct{}, 0)
	ep.desiredPolicy = policy.NewEndpointPolicy(owner.GetPolicyRepository())
	ep.realizedPolicy = ep.desiredPolicy
	ep.controllers = controller.NewManager()
	ep.regenFailedChan = make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(ctx)
	ep.aliveCancel = cancel
	ep.aliveCtx = ctx

	// If host label is present, it's the host endpoint.
	ep.isHost = ep.HasLabels(labels.LabelHost)

	// We need to check for nil in Status, CurrentStatuses and Log, since in
	// some use cases, status will be not nil and Cilium will eventually
	// error/panic if CurrentStatus or Log are not initialized correctly.
	// Reference issue GH-2477
	if ep.status == nil || ep.status.CurrentStatuses == nil || ep.status.Log == nil {
		ep.status = NewEndpointStatus()
	}

	// Make sure the endpoint has an identity, using the 'init' identity if none.
	if ep.SecurityIdentity == nil {
		ep.SecurityIdentity = identity.LookupReservedIdentity(identity.ReservedIdentityInit)
	}
	ep.SecurityIdentity.Sanitize()

	ep.setState(StateRestoring, "Endpoint restoring")

	return &ep, nil
}

// parseBase64ToEndpoint parses the endpoint stored in the given base64 string.
/*
{
  "ID": 627,
  "ContainerName": "",
  "dockerID": "0c97cd6828821764212a51357c3f91ad197e5a48ede137b43f36123e7032b3ee",
  "DockerNetworkID": "",
  "DockerEndpointID": "",
  "DatapathMapID": 0,
  "IfName": "lxc6a6af5d3c4d5",
  "IfIndex": 53510,
  "OpLabels": {
    "Custom": {},
    "OrchestrationIdentity": {
      "app": {
        "key": "app",
        "value": "test-qianyi-nginx",
        "source": "k8s"
      },
      "io.cilium.k8s.namespace.labels.field.cattle.io/projectId": {
        "key": "io.cilium.k8s.namespace.labels.field.cattle.io/projectId",
        "value": "p-mlnsg",
        "source": "k8s"
      },
      "io.cilium.k8s.policy.cluster": {
        "key": "io.cilium.k8s.policy.cluster",
        "value": "default",
        "source": "k8s"
      },
      "io.cilium.k8s.policy.serviceaccount": {
        "key": "io.cilium.k8s.policy.serviceaccount",
        "value": "default",
        "source": "k8s"
      },
      "io.kubernetes.pod.namespace": {
        "key": "io.kubernetes.pod.namespace",
        "value": "lxx",
        "source": "k8s"
      },
      "nonce": {
        "key": "nonce",
        "value": "v",
        "source": "k8s"
      },
      "stark-app": {
        "key": "stark-app",
        "value": "test-qianyi",
        "source": "k8s"
      },
      "stark-ns": {
        "key": "stark-ns",
        "value": "lxx1",
        "source": "k8s"
      },
      "stark-priority": {
        "key": "stark-priority",
        "value": "LS",
        "source": "k8s"
      },
      "stark-res": {
        "key": "stark-res",
        "value": "Deployment",
        "source": "k8s"
      }
    },
    "Disabled": {},
    "OrchestrationInfo": {
      "pod-template-hash": {
        "key": "pod-template-hash",
        "value": "7ffdf5877c",
        "source": "k8s"
      }
    }
  },
  "LXCMAC": "0e:13:1d:f7:ec:75",
  "IPv6": "",
  "IPv4": "10.216.152.158",
  "NodeMAC": "a6:34:be:ba:ad:37",
  "SecLabel": {
    "id": 55086,
    "labels": {
      "app": {
        "key": "app",
        "value": "test-qianyi-nginx",
        "source": "k8s"
      },
      "io.cilium.k8s.namespace.labels.field.cattle.io/projectId": {
        "key": "io.cilium.k8s.namespace.labels.field.cattle.io/projectId",
        "value": "p-mlnsg",
        "source": "k8s"
      },
      "io.cilium.k8s.policy.cluster": {
        "key": "io.cilium.k8s.policy.cluster",
        "value": "default",
        "source": "k8s"
      },
      "io.cilium.k8s.policy.serviceaccount": {
        "key": "io.cilium.k8s.policy.serviceaccount",
        "value": "default",
        "source": "k8s"
      },
      "io.kubernetes.pod.namespace": {
        "key": "io.kubernetes.pod.namespace",
        "value": "lxx",
        "source": "k8s"
      },
      "nonce": {
        "key": "nonce",
        "value": "v",
        "source": "k8s"
      },
      "stark-app": {
        "key": "stark-app",
        "value": "test-qianyi",
        "source": "k8s"
      },
      "stark-ns": {
        "key": "stark-ns",
        "value": "lxx1",
        "source": "k8s"
      },
      "stark-priority": {
        "key": "stark-priority",
        "value": "LS",
        "source": "k8s"
      },
      "stark-res": {
        "key": "stark-res",
        "value": "Deployment",
        "source": "k8s"
      }
    },
    "labelsSHA256": ""
  },
  "Options": {
    "map": {
      "Conntrack": 1,
      "ConntrackAccounting": 1,
      "ConntrackLocal": 0,
      "Debug": 0,
      "DebugLB": 0,
      "DebugPolicy": 0,
      "DropNotification": 1,
      "MonitorAggregationLevel": 3,
      "NAT46": 0,
      "PolicyAuditMode": 0,
      "PolicyVerdictNotification": 1,
      "TraceNotification": 1
    }
  },
  "DNSHistory": [],
  "DNSZombies": {},
  "K8sPodName": "test-qianyi-nginx-v-7ffdf5877c-7kd26",
  "K8sNamespace": "lxx",
  "DatapathConfiguration": {}
}
*/
func parseBase64ToEndpoint(str string, ep *Endpoint) error {
	jsonBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonBytes, ep); err != nil {
		return fmt.Errorf("error unmarshaling serializableEndpoint from base64 representation: %s", err)
	}

	return nil
}
