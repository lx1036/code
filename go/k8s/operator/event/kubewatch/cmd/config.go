package cmd

type Config struct {
	Namespace string
	Handlers  []Handler
	Resources Resources
}
type Handler struct {
}
type Resources struct {
	Deployment bool
	Pod        bool
}
