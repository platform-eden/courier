package registrar

type RegistrarOptions struct {
	NodeStorer          NodeStorer
	Port                string
	SubscribedSubjects  []string
	BroadcastedSubjects []string
}
