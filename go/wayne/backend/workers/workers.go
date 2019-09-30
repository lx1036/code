package workers

type Worker interface {
    Run() error
    Stop() error
}
