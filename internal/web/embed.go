// Package web provides embedded web assets for the SAML IDP.
package web

import "embed"

//go:embed templates/*
var Assets embed.FS
