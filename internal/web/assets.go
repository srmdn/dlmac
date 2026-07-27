package web

import "embed"

//go:embed assets/templates/index.html assets/static/*
var assetFS embed.FS
