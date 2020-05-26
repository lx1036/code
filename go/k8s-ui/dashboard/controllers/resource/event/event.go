package event

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"strings"
)

var FailedReasonPartials = []string{"failed", "err", "exceeded", "invalid", "unhealthy",
	"mismatch", "insufficient", "conflict", "outof", "nil", "backoff"}

func GetPodsEventWarnings(events []corev1.Event, pods []corev1.Pod) []Event {

}

func ListResourceEventsByQuery(
	k8sClient kubernetes.Interface,
	namespaceName string,
	resourceName string,
	dataSelect *dataselect.DataSelectQuery) (EventList, error) {
	fieldSelector, err := fields.ParseSelector(fmt.Sprintf("involvedObject.name=%s", resourceName))
	if err != nil {
		return EventList{}, err
	}

	channel := common.ResourceChannels{
		EventListChannel: GetEventListChannelWithOptions(
			k8sClient,
			namespaceName,
			metav1.ListOptions{
				FieldSelector: fieldSelector.String(),
				LabelSelector: labels.Everything().String(),
			},
			1),
	}
	eventList := <-channel.EventListChannel.List
	err = <-channel.EventListChannel.Error
	if err != nil {
		return EventList{}, err
	}
	return toEventList(FillEventsType(eventList.Items)), nil
}

func ListNamespaceEventsByQuery(
	k8sClient kubernetes.Interface,
	namespaceName string,
	dataSelect *dataselect.DataSelectQuery) (EventList, error) {
	rawEventList, err := k8sClient.CoreV1().Events(namespaceName).List(context.TODO(), common.ListEverything)
	if err != nil {
		return EventList{}, err
	}
	
	eventList := toEventList(FillEventsType(rawEventList.Items))
	
	return eventList, nil
}

func FillEventsType(events []corev1.Event) []corev1.Event {
	for i := range events {
		if len(events[i].Type) == 0 { // type is empty
			if isFailedReason(events[i].Reason, FailedReasonPartials...) {
				events[i].Type = corev1.EventTypeWarning
			} else {
				events[i].Type = corev1.EventTypeNormal
			}
		}
	}
	
	return events
}

func toEventList(rawEvents []corev1.Event) EventList {
	eventList := EventList{
		ListMeta: common.ListMeta{
			TotalItems: len(rawEvents),
		},
	}

	var events []Event
	for _, event := range rawEvents {
		 events = append(events,  Event{
			ObjectMeta: common.NewObjectMeta(event.ObjectMeta),
			TypeMeta: common.NewTypeMeta(common.ResourceKindEvent),
			Message: event.Message,
			SourceComponent: event.Source.Component,
			SourceHost: event.Source.Host,
			SubObject: event.InvolvedObject.FieldPath,
			Count: event.Count,
			FirstSeen: event.FirstTimestamp,
			LastSeen: event.LastTimestamp,
			Reason: event.Reason,
			Type: event.Type,
		})
	}

	eventList.Events = events

	return eventList
}

func isFailedReason(reason string, partials ...string) bool {
	for _, partial := range partials {
		if strings.Contains(strings.ToLower(reason), partial) {
			return true
		}
	}
	
	return false
}
