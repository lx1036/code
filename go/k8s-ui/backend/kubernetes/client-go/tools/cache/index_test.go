package cache

import (
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"strings"
	"testing"
)

/**
@see https://blog.csdn.net/weixin_42663840/article/details/81530606
 */

func testIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*v1.Pod)
	return []string{pod.Labels["foo"]}, nil
}

func TestGetIndexFuncValues(test *testing.T) {
	index := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{"testmodes": testIndexFunc})

	pod1 := &v1.Pod{
		TypeMeta:   metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{Name: "one", Labels: map[string]string{"foo": "bar"}},
		Spec:       v1.PodSpec{},
		Status:     v1.PodStatus{},
	}

	pod2 := &v1.Pod{
		TypeMeta:   metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{Name: "two", Labels: map[string]string{"foo": "bar"}},
		Spec:       v1.PodSpec{},
		Status:     v1.PodStatus{},
	}

	pod3 := &v1.Pod{
		TypeMeta:   metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{Name: "three", Labels: map[string]string{"foo": "biz"}},
		Spec:       v1.PodSpec{},
		Status:     v1.PodStatus{},
	}

	_ = index.Add(pod1)
	_ = index.Add(pod2)
	_ = index.Add(pod3)

	keys := index.ListIndexFuncValues("testmodes")
	for _, key := range keys {
		if key != "bar" && key != "biz" {
			test.Errorf("want bar or biz, got %s", key)
		}
	}
}

func testUsersIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*v1.Pod)
	usersAnnotations := pod.Annotations["users"]
	return strings.Split(usersAnnotations, ","), nil
}

func TestMultiIndexKeys(test *testing.T) {
	index := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{"byUser": testUsersIndexFunc})

	pod1 := &v1.Pod{
		TypeMeta:   metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{Name: "one", Annotations: map[string]string{"users": "user1,user2"}},
		Spec:       v1.PodSpec{},
		Status:     v1.PodStatus{},
	}

	pod2 := &v1.Pod{
		TypeMeta:   metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{Name: "two", Annotations: map[string]string{"users": "user2,user3"}},
		Spec:       v1.PodSpec{},
		Status:     v1.PodStatus{},
	}

	pod3 := &v1.Pod{
		TypeMeta:   metaV1.TypeMeta{},
		ObjectMeta: metaV1.ObjectMeta{Name: "three", Annotations: map[string]string{"users": "user1,user4"}},
		Spec:       v1.PodSpec{},
		Status:     v1.PodStatus{},
	}

	_ = index.Add(pod1)
	_ = index.Add(pod2)
	_ = index.Add(pod3)

	expected := map[string]sets.String{}
	expected["user1"] = sets.NewString("one", "three")
	expected["user2"] = sets.NewString("one", "two")
	expected["user3"] = sets.NewString("two")
	expected["user4"] = sets.NewString("three")
	expected["user4"] = sets.NewString()

	{
		for k, v := range expected {
			found := sets.String{}
			results, err :=index.ByIndex("byUser", k)
			if err != nil {
				test.Errorf("error: %v", err)
			}
			for _, result := range results {
				found.Insert(result.(*v1.Pod).Name)
			}
			items := v.List()
			if !found.HasAll(items...) {
				test.Errorf("got %s, want %s", found.List(), items)
			}
		}
	}

	// delete pod3
	_ = index.Delete(pod3)
	user1Pods, err :=index.ByIndex("byUser", "user1")
	if err != nil {
		test.Errorf("error: %v", err)
	}
	if len(user1Pods) != 1 {
		test.Errorf("got %d, want %d", len(user1Pods), 1)
	}
	for _, pod := range user1Pods {
		if pod.(*v1.Pod).Name != "one" {
			test.Errorf("got %s, want %s", pod.(*v1.Pod).Name, "one")
		}
	}
	user4Pods, err :=index.ByIndex("byUser", "user4")
	if err != nil {
		test.Errorf("error: %v", err)
	}
	if len(user4Pods) != 0 {
		test.Errorf("got %d, want %d", len(user4Pods), 0)
	}
	
	// update pod2
	copyOfPod2 := pod2.DeepCopy() // 这里不是添加了一个新的pod，而是去更新pod2的annotation值为"user3"
	copyOfPod2.Annotations["users"] = "user3"
	_ = index.Update(copyOfPod2)
	user2Pods, err := index.ByIndex("byUser", "user2")
	if err != nil {
		test.Errorf("error: %v", err)
	}
	if len(user2Pods) != 1 {
		test.Errorf("got %d, want %d", len(user2Pods), 1)
	}
	for _, pod := range user2Pods {
		if pod.(*v1.Pod).Name != "one" {
			test.Errorf("got %s, want %s", pod.(*v1.Pod).Name, "one")
		}
	}
	if len(index.List()) != 2 {
		test.Errorf("got %d, want %d", len(index.List()), 2)
	}
	for _, item := range index.List() {
		if item.(*v1.Pod).Name != "one" && item.(*v1.Pod).Name != "two" {
			test.Errorf("got %s, want %s", item.(*v1.Pod).Name, "one,two")
		}
	}
}
