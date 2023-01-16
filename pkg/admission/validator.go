package admission

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
)

// ResourceValidator struct holds a resource information required to preform the validations
type ResourceValidator struct {
	Logger     *logrus.Entry
	arRequest  *admissionv1.AdmissionReview
	arResponse *admissionv1.AdmissionReview
}

//ServerHandler listen to admission requests and serve responses
type ServerHandler struct {
}

func (gs *ServerHandler) serve(w http.ResponseWriter, r *http.Request) {
	namespace := options.Namespace
	log := logrus.WithField("admission validator", namespace)
	var arResponse admissionv1.AdmissionReview
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		log.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}
	log.Info("Received request")

	if r.URL.Path != "/validate" {
		log.Error("no validate")
		http.Error(w, "no validate", http.StatusBadRequest)
		return
	}

	arRequest := admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, &arRequest); err != nil {
		log.Error("incorrect body")
		http.Error(w, "incorrect body", http.StatusBadRequest)
		return
	}

	switch arRequest.Request.Resource.Resource {
	case "backingstores":
		arResponse = NewBackingStoreValidator(arRequest).ValidateBackingstore()
	case "namespacestores":
		arResponse = NewNamespaceStoreValidator(arRequest).ValidateNamespaceStore()
	case "bucketclasses":
		arResponse = NewBucketClassValidator(arRequest).ValidateBucketClass()
	case "noobaaaccounts":
		arResponse = NewNoobaaAccountValidator(arRequest).ValidateNoobaAaccount()
	case "noobaas":
		arResponse = NewNoobaaValidator(arRequest).ValidateNoobaa()
	default:
		log.Error("failed to identify resource type")
		http.Error(w, "incorrect resource", http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(arResponse)
	if err != nil {
		log.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}
	log.Infof("Ready to write response ...")
	if _, err := w.Write(resp); err != nil {
		log.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		return
	}
}

// SetValidationResult responsible of assinging the return values of a validation into the response appropriate fields
func (rv *ResourceValidator) SetValidationResult(isAllowed bool, message string) {
	rv.arResponse.Response.Allowed = isAllowed
	rv.arResponse.Response.Result.Message = message
}
