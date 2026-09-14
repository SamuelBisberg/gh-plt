// Package templates embeds gh-plt's template catalog (one directory per
// <lang>/[framework/]<kind> template, each holding a template.yaml manifest
// plus its content files) and the ecosystems.yaml language catalog,
// compiled into the binary at build time.
//
// Both live here, outside pkg/, so contributors can add or edit a workflow
// template or a language definition without touching any Go code.
package templates

import "embed"

//go:embed all:*
var FS embed.FS
