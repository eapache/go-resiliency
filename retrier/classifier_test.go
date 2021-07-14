package retrier

import (
	"errors"
	"testing"
)

var (
	errFoo = errors.New("FOO")
	errBar = errors.New("BAR")
	errBaz = errors.New("BAZ")
)

func TestDefaultClassifier(t *testing.T) {
	c := DefaultClassifier{}

	if c.Classify(nil) != Succeed {
		t.Error("default misclassified nil")
	}

	if c.Classify(errFoo) != Retry {
		t.Error("default misclassified foo")
	}
	if c.Classify(errBar) != Retry {
		t.Error("default misclassified bar")
	}
	if c.Classify(errBaz) != Retry {
		t.Error("default misclassified baz")
	}
}

type wrappedErr struct {
	error
}

func (w wrappedErr) Error() string {
	return "there's an error happening during X: " + w.Error()
}

func (w wrappedErr) Unwrap() error {
	return w.error
}

func TestWhitelistClassifier(t *testing.T) {
	c := WhitelistClassifier{errFoo, errBar}

	if c.Classify(nil) != Succeed {
		t.Error("whitelist misclassified nil")
	}

	if c.Classify(errFoo) != Retry {
		t.Error("whitelist misclassified foo")
	}
	if c.Classify(errBar) != Retry {
		t.Error("whitelist misclassified bar")
	}
	if c.Classify(errBaz) != Fail {
		t.Error("whitelist misclassified baz")
	}

	if c.Classify(wrappedErr{error: errFoo}) != Retry {
		t.Error("whitelist misclassified foo")
	}
	if c.Classify(wrappedErr{error: errBar}) != Retry {
		t.Error("whitelist misclassified bar")
	}
	if c.Classify(wrappedErr{error: errBaz}) != Fail {
		t.Error("whitelist misclassified baz")
	}
}

func TestBlacklistClassifier(t *testing.T) {
	c := BlacklistClassifier{errBar}

	if c.Classify(nil) != Succeed {
		t.Error("blacklist misclassified nil")
	}

	if c.Classify(errFoo) != Retry {
		t.Error("blacklist misclassified foo")
	}
	if c.Classify(errBar) != Fail {
		t.Error("blacklist misclassified bar")
	}
	if c.Classify(errBaz) != Retry {
		t.Error("blacklist misclassified baz")
	}

	if c.Classify(wrappedErr{error: errFoo}) != Retry {
		t.Error("blacklist misclassified foo")
	}
	if c.Classify(wrappedErr{error: errBar}) != Fail {
		t.Error("blacklist misclassified bar")
	}
	if c.Classify(wrappedErr{error: errBaz}) != Retry {
		t.Error("blacklist misclassified baz")
	}
}
