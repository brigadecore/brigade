package core

// Labels is a map of key/value pairs utilized mutually by Events in describing
// themselves and by EventSubscriptions in describing Events of interest to a
// Project.
type Labels map[string]string
