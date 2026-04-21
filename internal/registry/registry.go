package registry

type Service interface {
	Name() string
	Close()
}

type ServiceRegistry struct {
	services []Service
}

func New() *ServiceRegistry {
	return &ServiceRegistry{}
}

func (r *ServiceRegistry) Add(services ...Service) {
	r.services = append(r.services, services...)
}

func (r *ServiceRegistry) Close() {
	for _, s := range r.services {
		s.Close()
	}
}

func (r *ServiceRegistry) Names() []string {
	names := make([]string, len(r.services))
	for i, s := range r.services {
		names[i] = s.Name()
	}
	return names
}
