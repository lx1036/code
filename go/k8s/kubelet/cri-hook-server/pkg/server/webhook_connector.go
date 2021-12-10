package server



// hookerConnector used to forward docker request to backend
type webhookConnector struct {
	name          string
	endpoint      string
	failurePolicy v1.FailurePolicyType
	client        *http.Client
}

func newWebhookConnector(name, endpoint string, failurePolicy v1.FailurePolicyType)  {

}
