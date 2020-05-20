package jwe

import (
	"k8s-lx1036/dashboard/backend/sync"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func getKeyHolder() KeyHolder {
	client := fake.NewSimpleClientset()
	syncManager := sync.NewSynchronizerManager(client)
	return NewRSAKeyHolder(syncManager.Secret("", ""))
}

func TestNewRSAKeyHolder(test *testing.T) {
	holder := getKeyHolder()
	if holder == nil {
		test.Fatalf("NewRSAKeyHolder(): Expected key holder not to be nil")
	}
}

func TestRsaKeyHolder_Encrypter(test *testing.T) {
	holder := getKeyHolder()
	if holder.Encrypter() == nil {
		test.Fatalf("Encrypter(): Expected encrypter not to be nil")
	}
}

func TestRsaKeyHolder_Key(test *testing.T) {
	holder := getKeyHolder()
	if holder.Key() == nil {
		test.Fatalf("Key(): Expected key not to be nil")
	}
}
