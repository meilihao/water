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

func (uriBinding) Bind2(m map[string][]string, obj interface{}) error {
	if err := MapForm(obj, m, nil, "uri"); err != nil {
		return err
	}
	return validate(obj)
}
