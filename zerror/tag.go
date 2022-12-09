package zerror

type TagKind string

const (
	// Empty error
	None TagKind = ""
	// Internal errors, This means that some invariants expected by the underlying system have been broken
	Internal TagKind = "INTERNAL"
	// The operation was cancelled, typically by the caller
	Cancelled TagKind = "CANCELLED"
	// The client specified an invalid argument
	InvalidInput TagKind = "INVALID_INPUT"
	// Some requested entity was not found
	NotFound TagKind = "NOT_FOUND"
	// The caller does not have permission to execute the specified operation
	PermissionDenied TagKind = "PERMISSION_DENIED"
	// The request does not have valid authentication credentials for the operation
	Unauthorized TagKind = "UNAUTHORIZED"
)

func (t TagKind) Wrap(err error, text string) error {
	return With(err, text, WrapTag(t))
}

type withTag struct {
	wrapErr error
	tag     TagKind
}

func (e *withTag) Error() string {
	return e.wrapErr.Error()
}

func WrapTag(tag TagKind) External {
	return func(err error) error {
		return &withTag{
			wrapErr: err,
			tag:     tag,
		}
	}
}

func GetTag(err error) TagKind {
	if err == nil {
		return None
	}

	for err != nil {
		if f, ok := err.(*withTag); ok {
			return f.tag
		}

		if e, ok := err.(*Error); ok {
			err = e.wrapErr
			if err == nil {
				err = e.err
			}
		} else {
			break
		}
	}

	return None
}
