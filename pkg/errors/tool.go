package errors

func GatherErrors(op ...func() error) []error {
	errors := make([]error, 0, len(op))

	for _, f := range op {
		if err := f(); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
