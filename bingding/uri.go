package binding

type uriBinding struct{}

func (uriBinding) Name() string {
	return "uri"
}

func (uriBinding) Bind(m map[string][]string, obj interface{}) error {
	if err := mapFormByTag(obj, m, "uri"); err != nil {
		return err
	}
	return validate(obj)
}
