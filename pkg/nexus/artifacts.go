package nexus

// Artifacts is a collection of artifacts, sortable with utiltity functions.
type Artifacts []*Artifact

// Latest resurns the latest version number in the list.
func (a Artifacts) Latest() *Artifact {
	var latest *Artifact
	for _, artifact := range a {
		if latest == nil || artifact.GreaterThan(latest) {
			latest = artifact
		}
	}
	return latest
}

// Len is implemnting sort.Interface
func (v Artifacts) Len() int {
	return len(v)
}

// Less is implementing sort.Interface
func (v Artifacts) Less(i, j int) bool {
	return v[i].LessThan(v[j])
}

// Swap is implementing sort.Interface
func (v Artifacts) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
