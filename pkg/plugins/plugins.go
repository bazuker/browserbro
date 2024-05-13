package plugins

type Plugin interface {
	Name() string
	Run(Params map[string]any) (map[string]any, error)
}
