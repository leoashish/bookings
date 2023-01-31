package forms

type errors map[string][]string

// Adds an error message to the given field
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get returns the first error message.
func (e errors) Get(field string) string {
	es := e[field]

	if len(es) == 0 {
		return ""
	} else {
		return es[0]
	}
}
