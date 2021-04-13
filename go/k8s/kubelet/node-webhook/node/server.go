package node

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// admitFunc is the type we use for all of our validators and mutators
type admitFunc func(v1.AdmissionReview) *v1.AdmissionResponse

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
func toAdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// serve handles the http portion of a request prior to handing to an admit function
func Serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	requestedAdmissionReview := v1.AdmissionReview{}
	responseAdmissionReview := &v1.AdmissionReview{}
	if _, gvk, err := codecs.UniversalDeserializer().Decode(body, nil, &requestedAdmissionReview); err != nil {
		klog.Errorf("failed to deserialize request body to AdmissionReview obj with err: %v", err)
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Error(err)
	}

	klog.Infof("sending response: %s", string(respBytes))

	if _, err = w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}
