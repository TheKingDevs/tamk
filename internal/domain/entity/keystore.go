package entity

type Keystore struct {
	Path     string
	Password string
	Alias    string
	Validity int
}

func NewKeystore(path, password, alias string) *Keystore {
	return &Keystore{
		Path:     path,
		Password: password,
		Alias:    alias,
		Validity: 10000,
	}
}
