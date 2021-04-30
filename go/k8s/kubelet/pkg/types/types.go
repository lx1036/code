package types

import "k8s.io/apimachinery/pkg/types"

// A pod UID which has been translated/resolved to the representation known to kubelets.
type ResolvedPodUID types.UID

// A pod UID for a mirror pod.
type MirrorPodUID types.UID
