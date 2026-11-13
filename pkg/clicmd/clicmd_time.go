package clicmd

import "time"

// timeValue is the pflag.Value of the time.Time flags and positional
// arguments: an RFC 3339 timestamp (e.g. "2023-06-07T10:00:00+02:00"). The
// empty string sets the zero time, i.e. unset, like an empty string flag.
//
// pflag can't hold a time.Time by itself: time.Time doesn't implement
// pflag.Value, and pflag v1.0.5 has no time flag which means we have
// to create a private pflag.Value to help manage time.Time value
type timeValue time.Time

func (v *timeValue) Set(s string) error {
	if s == "" {
		*v = timeValue{}
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*v = timeValue(t)
	return nil
}

func (v *timeValue) String() string {
	if time.Time(*v).IsZero() {
		return ""
	}

	return time.Time(*v).Format(time.RFC3339)
}

// Type is "string" and not "time": the help shows "--flag string", and an
// unset flag shows no default value.
func (v *timeValue) Type() string { return "string" }
