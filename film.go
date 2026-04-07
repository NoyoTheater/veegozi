package veegozi

import (
	"context"
	"fmt"
	"time"
)

type FilmStatus string

const (
	FilmStatusActive   FilmStatus = "Active"
	FilmStatusDeleted  FilmStatus = "Deleted"
	FilmStatusInactive FilmStatus = "Inactive"
)

type FilmFormat string

const (
	FilmFormatFilm2D    FilmFormat = "2D Film"
	FilmFormatDigital2D FilmFormat = "2D Digital"
	FilmFormatDigital3D FilmFormat = "3D Digital"
	FilmFormatHFR3D     FilmFormat = "3D HFR"
	FilmFormatNotAFilm  FilmFormat = "Not a Film"
)

type Person struct {
	ID        string
	FirstName string
	LastName  string
	Role      string
}

type Film struct {
	ID                     string
	Title                  string
	ShortName              string
	Synopsis               string
	Genre                  string
	SignageText            string
	Distributor            string
	OpeningDate            time.Time
	Rating                 string
	Status                 FilmStatus
	Content                string
	Duration               uint32
	DisplaySequence        uint32
	NationalCode           string
	Format                 FilmFormat
	IsRestricted           bool
	People                 []Person
	AudioLanguage          string
	GovernmentFilmTitle    string
	FilmPosterURL          string
	FilmPosterThumbnailURL string
	BackdropImageURL       string
	FilmTrailerURL         string
}
type FilmList struct {
	Films []Film
}

func (f Film) Sessions(ctx context.Context, c *Client) (SessionList, error) {
	sessions, err := c.GetSessions(ctx)
	if err != nil {
		return SessionList{}, err
	}

	return sessions.FilterByFilm(f.ID), nil
}

func (f Film) WebSessions(ctx context.Context, c *Client) (SessionList, error) {
	sessions, err := c.GetWebSessions(ctx)
	if err != nil {
		return SessionList{}, err
	}

	return sessions.FilterByFilm(f.ID), nil
}

func (f Film) FormattedDuration() string {
	hours := f.Duration / 60
	minutes := f.Duration % 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	} else {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}
