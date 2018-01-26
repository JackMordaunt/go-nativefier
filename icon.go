package main

import "github.com/jackmordaunt/pageicon"

// IconInferrer abstracts how to retrieve an icon for a given url.
type IconInferrer interface {
	Infer(url string, prefs []string) (*Icon, error)
}

// IconInferrerFunc wraps functions to satisfy IconInferrer.
type IconInferrerFunc func(string, []string) (*Icon, error)

// Infer calls the func responsible for inferring the icon.
func (fn IconInferrerFunc) Infer(url string, prefs []string) (*Icon, error) {
	return fn(url, prefs)
}

// Icon contains metadata about an icon.
type Icon = pageicon.Icon
