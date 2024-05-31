package api

// AgentNotification specifies the type of agent notification
type AgentNotification uint32

const (
    AgentNotifyUnspec AgentNotification = iota
    AgentNotifyGeneric
    AgentNotifyStart
    AgentNotifyEndpointRegenerateSuccess
    AgentNotifyEndpointRegenerateFail
    AgentNotifyPolicyUpdated
    AgentNotifyPolicyDeleted
    AgentNotifyEndpointCreated
    AgentNotifyEndpointDeleted
    AgentNotifyIPCacheUpserted
    AgentNotifyIPCacheDeleted
    AgentNotifyServiceUpserted
    AgentNotifyServiceDeleted
)

// AgentNotifyMessage is a notification from the agent. It is similar to
// AgentNotify, but the notification is an unencoded struct. See the *Message
// constructors in this package for possible values.
type AgentNotifyMessage struct {
    Type         AgentNotification
    Notification interface{}
}
