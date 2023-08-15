package mailer

type Configurer interface {
	Has(name string) bool
	UnmarshalKey(name string, out interface{}) error
}
