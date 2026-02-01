package tinyrbac

type config struct {
	Description string
	Roles       []role
}

type role struct {
	Name        string
	Description string
	Resources   []resource
}

type resource struct {
	Name    string
	Actions []string
}
