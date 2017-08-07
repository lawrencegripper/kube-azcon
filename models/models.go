package models

type SecretsMap map[string]string

type Output struct {
	Secrets     SecretsMap
	Port        int
	Endpoint    string
	ServiceName string
	Namespace   string
}

func (s *Output) GetSecretMap() map[string][]byte {
	res := make(map[string][]byte)
	for k, v := range s.Secrets{
		res[k] = []byte(v)
	}
	return res
}
