package serviceaccount

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	rcache "k8s-lx1036/k8s/network/calico/kube-controllers/pkg/cache"
	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/calico"
	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/controller"
	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/converter"
	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/kube"

	log "github.com/sirupsen/logrus"

	api "github.com/projectcalico/calico/libcalico-go/lib/apis/v3"
	kdd "github.com/projectcalico/calico/libcalico-go/lib/backend/k8s/conversion"
	//client "github.com/projectcalico/calico/libcalico-go/lib/clientv3"
	"github.com/projectcalico/calico/libcalico-go/lib/errors"
	"github.com/projectcalico/calico/libcalico-go/lib/options"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

const (
	// maxRetries is the number of times a service will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the
	// sequence of delays between successive queuings of a service.
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

// serviceAccountController implements the Controller interface for managing Kubernetes service account
// and syncing them to the Calico datastore as Profiles.
type serviceAccountController struct {
	informer      cache.Controller
	resourceCache rcache.ResourceCache
	calicoClient  client.Interface
	ctx           context.Context
	//cfg           config.GenericControllerConfig
}

func NewServiceAccountController(ctx context.Context) controller.Controller {
	kubeClientset := kube.GetKubernetesClientset()
	calicoClient := calico.GetCalicoClientOrDie()

	serviceAccountConverter := converter.NewServiceAccountConverter()

	// Function returns map of profile_name:object stored by policy controller
	// in the Calico datastore. Identifies controller written objects by
	// their naming convention.
	listFunc := func() (map[string]interface{}, error) {
		log.Debugf("Listing profiles from Calico datastore: to check for ServiceAccount")
		filteredProfiles := make(map[string]interface{})

		// Get all profile objects from Calico datastore.
		profileList, err := calicoClient.Profiles().List(ctx, options.ListOptions{})
		if err != nil {
			return nil, err
		}

		// Filter out only objects that are written by policy controller.
		for _, profile := range profileList.Items {
			if strings.HasPrefix(profile.Name, kdd.ServiceAccountProfileNamePrefix) {
				// Update the profile's ObjectMeta so that it simply contains the name.
				// There is other metadata that we might receive (like resource version) that we don't want to
				// compare in the cache.
				profile.ObjectMeta = metav1.ObjectMeta{Name: profile.Name}
				key := serviceAccountConverter.GetKey(profile)
				filteredProfiles[key] = profile
			}
		}
		log.Debugf("Found %d ServiceAccount profiles in Calico datastore", len(filteredProfiles))
		return filteredProfiles, nil
	}
	resourceCache := rcache.NewResourceCache(rcache.ResourceCacheArgs{
		Name:             "serviceaccount",
		ListFunc:         listFunc,
		ObjectType:       reflect.TypeOf(api.Profile{}),
		LogTypeDesc:      "ServiceAccount",
		ReconcilerConfig: nil,
	})

	listWatcher := cache.NewListWatchFromClient(kubeClientset.CoreV1().RESTClient(), "serviceaccounts", metav1.NamespaceAll, fields.Everything())
	_, informer := cache.NewIndexerInformer(listWatcher, &v1.ServiceAccount{}, time.Minute*2, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Debugf("Got ADD event for ServiceAccount: %#v", obj)
			profile, err := serviceAccountConverter.Convert(obj)
			if err != nil {
				log.WithError(err).Errorf("Error while converting %#v to Calico profile.", obj)
				return
			}

			// Add to cache.
			k := serviceAccountConverter.GetKey(profile)
			resourceCache.Set(k, profile)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Debugf("Got UPDATE event for ServiceAccount, Old object: %#v, New object: %#v", oldObj, newObj)
			profile, err := serviceAccountConverter.Convert(newObj)
			if err != nil {
				log.WithError(err).Errorf("Error while converting %#v to Calico profile.", newObj)
				return
			}

			// Update in the cache.
			k := serviceAccountConverter.GetKey(profile)
			resourceCache.Set(k, profile)
		},
		DeleteFunc: func(obj interface{}) {
			// Convert the ServiceAccount into a Profile.
			log.Debugf("Got DELETE event for ServiceAccount: %#v", obj)
			profile, err := serviceAccountConverter.Convert(obj)
			if err != nil {
				log.WithError(err).Errorf("Error while converting %#v to Calico profile.", obj)
				return
			}

			k := serviceAccountConverter.GetKey(profile)
			resourceCache.Delete(k)
		},
	}, cache.Indexers{})

	return &serviceAccountController{
		informer:      informer,
		resourceCache: resourceCache,
		calicoClient:  calicoClient,
		ctx:           ctx,
	}

}

// Run starts the controller.
func (c *serviceAccountController) Run(workers int, stopCh chan struct{}) {
	defer utilruntime.HandleCrash()

	// Let the workers stop when we are done
	workqueue := c.resourceCache.GetQueue()
	defer workqueue.ShutDown()

	log.Info("Starting ServiceAccount/Profile controller")

	// Wait till k8s cache is synced
	log.Debug("Waiting to sync with Kubernetes API (ServiceAccount)")
	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		panic(fmt.Errorf("ServiceAccount cache sync failed"))
	}

	log.Debug("Finished syncing with Kubernetes API (ServiceAccount)")

	// Start Calico cache.
	c.resourceCache.Run((time.Minute * 5).String())

	// Start a number of worker threads to read from the queue.
	for i := 0; i < workers; i++ {
		go wait.Until(func() {
			for c.processNextItem() {
			}
		}, time.Second, stopCh)
	}
	log.Info("ServiceAccount/Profile controller is now running")

	<-stopCh
	log.Info("Stopping ServiceAccount/Profile controller")
}

// processNextItem waits for an event on the output queue from the resource cache and syncs
// any received keys to the datastore.
func (c *serviceAccountController) processNextItem() bool {
	// Wait until there is a new item in the work queue.
	workqueue := c.resourceCache.GetQueue()
	key, quit := workqueue.Get()
	if quit {
		return false
	}
	defer workqueue.Done(key)

	err := c.syncToDatastore(key.(string))
	c.handleErr(err, key.(string))

	return true
}

func (c *serviceAccountController) handleErr(err error, key interface{}) {
	workqueue := c.resourceCache.GetQueue()
	if err == nil {
		workqueue.Forget(key)
		return
	}

	if workqueue.NumRequeues(key) < maxRetries {
		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		log.WithError(err).Errorf("Error syncing Profile %v: %v", key, err)
		workqueue.AddRateLimited(key)
		return
	}

	workqueue.Forget(key)

	utilruntime.HandleError(err)
	log.WithError(err).Errorf("Dropping Profile %q out of the queue: %v", key, err)
}

// syncToDatastore syncs the given update to the Calico datastore. The provided key can be used to
// find the corresponding resource within the resource cache.
// 1. If the resource for the provided key exists in the cache, then the value should be written to the datastore.
// 2. If it does not exist in the cache, then it should be deleted from the datastore.
// 1. 不存在cache中存在datastore中，则删除profile
// 2. 存在cache中不存在datastore中，则创建profile
// 3. 存在cache中存在datastore中，则更新profile
func (c *serviceAccountController) syncToDatastore(key string) error {
	clog := log.WithField("key", key)

	// Check if it exists in the controller's cache.
	obj, exists := c.resourceCache.Get(key)
	if !exists {
		// The object no longer exists - delete from the datastore.
		clog.Infof("Deleting ServiceAccount Profile from Calico datastore")
		_, name := converter.NewServiceAccountConverter().DeleteArgsFromKey(key)
		_, err := c.calicoClient.Profiles().Delete(c.ctx, name, options.DeleteOptions{})
		if _, ok := err.(errors.ErrorResourceDoesNotExist); !ok {
			// We hit an error other than "does not exist".
			return err
		}
		return nil
	} else {
		// The object exists - update the datastore to reflect.
		clog.Info("Create/Update ServiceAccount Profile in Calico datastore")
		profile := obj.(api.Profile)

		// Lookup to see if this object already exists in the datastore.
		oldProfile, err := c.calicoClient.Profiles().Get(c.ctx, profile.Name, options.GetOptions{})
		if err != nil {
			if _, ok := err.(errors.ErrorResourceDoesNotExist); !ok {
				clog.WithError(err).Warning("Unexpected error for ServiceAccount profile from datastore")
				return err
			}

			// Doesn't exist - create it.
			_, err := c.calicoClient.Profiles().Create(c.ctx, &profile, options.SetOptions{})
			if err != nil {
				clog.WithError(err).Warning("Failed to create ServiceAccount profile")
				return err
			}
			clog.Info("Successfully created ServiceAccount profile")
			return nil
		}

		oldProfile.Spec = profile.Spec
		clog.Infof("Update ServiceAccount Profile in Calico datastore with resource version %s", oldProfile.ResourceVersion)
		_, err = c.calicoClient.Profiles().Update(c.ctx, oldProfile, options.SetOptions{})
		if err != nil {
			clog.WithError(err).Warning("Failed to update profile")
			return err
		}
		clog.Infof("Successfully updated profile")
		return nil
	}
}
