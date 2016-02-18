package instancesmanager

type InstanceMap map[string]*Instance

type Instance struct {
	Backend        string
	Name           string
	Params         interface{}
	Run            bool
	Process        *chan string `json:"-"`
	ProcessRunning bool
}
