package execution

type Context struct {
	Outputs map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		Outputs: make(map[string]interface{}),
	}
}

func (c *Context) SetOutput(stepName string, output interface{}) {
	c.Outputs[stepName] = output
}

func (c *Context) GetOutput(stepName string) (interface{}, bool) {
	output, ok := c.Outputs[stepName]
	return output, ok
}
