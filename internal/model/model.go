package model

type URL struct {
	Alias       string
	OriginalUrl string
}

type URLCreate struct {
	OriginalUrl string
	Alias       string
}

type URLUpdate struct {
	OldAlias string
	NewAlias string
}
