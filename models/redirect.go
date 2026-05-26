package models

import "time"

type Redirect struct {
	ID           int64
	Slug         string
	TargetURL    string
	IsActive     bool
	RedirectType string
	UTMSource    string
	UTMMedium    string
	UTMCampaign  string
	UTMTerm      string
	UTMContent   string
	VisitCount   int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

const (
	RedirectTypeMetaRefresh = "meta_refresh"
	RedirectTypeDirect301   = "direct_301"
)
