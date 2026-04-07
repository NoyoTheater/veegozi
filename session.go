package veegozi

import (
	"context"
	"time"
)

type Seating string

const (
	SeatingAllocated Seating = "Allocated"
	SeatingSelect    Seating = "Select"
	SeatingOpen      Seating = "Open"
)

type ShowType string

const (
	ShowTypePrivate ShowType = "Private"
	ShowTypePublic  ShowType = "Public"
)

type SessionStatus string

const (
	SessionStatusOpen    SessionStatus = "Open"
	SessionStatusClosed  SessionStatus = "Closed"
	SessionStatusPlanned SessionStatus = "Planned"
)

type SalesChannel string

const (
	SalesChannelKiosk SalesChannel = "KIOSK"
	SalesChannelPos   SalesChannel = "POS"
	SalesChannelWww   SalesChannel = "WWW"
	SalesChannelMx    SalesChannel = "MX"
	SalesChannelRsp   SalesChannel = "RSP"
)

type SessionList struct {
	Sessions []Session
}

func (sl SessionList) FilterByScreen(screenID string) SessionList {
	var filtered []Session
	for _, session := range sl.Sessions {
		if session.ScreenID == screenID {
			filtered = append(filtered, session)
		}
	}
	return SessionList{Sessions: filtered}
}

func (sl SessionList) FilterByFilm(filmID string) SessionList {
	var filtered []Session
	for _, session := range sl.Sessions {
		if session.FilmID == filmID {
			filtered = append(filtered, session)
		}
	}
	return SessionList{Sessions: filtered}
}

func (sl SessionList) FilterContainingAttribute(attributeID string) SessionList {
	var filtered []Session
	for _, session := range sl.Sessions {
		for _, attrID := range session.Attributes {
			if attrID == attributeID {
				filtered = append(filtered, session)
				break
			}
		}
	}
	return SessionList{Sessions: filtered}
}

func (sl SessionList) FilterByTimeRange(start, end time.Time) SessionList {
	var filtered []Session
	for _, session := range sl.Sessions {
		if session.FeatureStartTime.After(start) && session.FeatureEndTime.Before(end) {
			filtered = append(filtered, session)
		}
	}
	return SessionList{Sessions: filtered}
}

func (sl SessionList) GroupByDate() map[string]SessionList {
	grouped := make(map[string]SessionList)
	for _, session := range sl.Sessions {
		dateKey := session.FeatureStartTime.Format("2006-01-02")
		dateSessions := grouped[dateKey]
		dateSessions.Sessions = append(dateSessions.Sessions, session)
		grouped[dateKey] = dateSessions
	}
	return grouped
}

func (sl SessionList) Films(ctx context.Context, c *Client) (FilmList, error) {
	filmMap := make(map[string]Film)
	for _, session := range sl.Sessions {
		if _, exists := filmMap[session.FilmID]; !exists {
			film, err := session.Film(ctx, c)
			if err != nil {
				return FilmList{}, err
			}
			filmMap[session.FilmID] = film
		}
	}

	var films []Film
	for _, film := range filmMap {
		films = append(films, film)
	}
	return FilmList{Films: films}, nil
}

func (sl SessionList) Screens(ctx context.Context, c *Client) (ScreenList, error) {
	screenMap := make(map[string]Screen)
	for _, session := range sl.Sessions {
		if _, exists := screenMap[session.ScreenID]; !exists {
			screen, err := session.Screen(ctx, c)
			if err != nil {
				return ScreenList{}, err
			}
			screenMap[session.ScreenID] = screen
		}
	}

	var screens []Screen
	for _, screen := range screenMap {
		screens = append(screens, screen)
	}
	return ScreenList{Screens: screens}, nil
}

type Session struct {
	ID                        string
	FilmID                    string
	FilmPackageID             string
	Title                     string
	ScreenID                  string
	Seating                   Seating
	AreComplimentariesAllowed bool
	ShowType                  ShowType
	SalesVia                  []SalesChannel
	Status                    SessionStatus
	PreShowStartTime          time.Time
	SalesCutOffTime           time.Time
	FeatureStartTime          time.Time
	FeatureEndTime            time.Time
	CleanupEndTime            time.Time
	TicketsSoldOut            bool
	FewTicketsLeft            bool
	SeatsAvailable            uint32
	SeatsHeld                 uint32
	SeatsHouse                uint32
	SeatsSold                 uint32
	FilmFormat                FilmFormat
	PriceCardName             string
	Attributes                []string
	AudioLanguage             string
}

func (s Session) Film(ctx context.Context, c *Client) (Film, error) {
	return c.GetFilm(ctx, s.FilmID)
}

func (s Session) FilmPackage(ctx context.Context, c *Client) (FilmPackage, error) {
	return c.GetFilmPackage(ctx, s.FilmPackageID)
}

func (s Session) Screen(ctx context.Context, c *Client) (Screen, error) {
	return c.GetScreen(ctx, s.ScreenID)
}

func (s Session) GetAttributes(ctx context.Context, c *Client) ([]Attribute, error) {
	var attributes []Attribute
	for _, attrID := range s.Attributes {
		attr, err := c.GetAttribute(ctx, attrID)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attr)
	}
	return attributes, nil
}

func (s Session) IsOpenForSales() bool {
	now := time.Now()
	return (s.Status == SessionStatusOpen) && now.Before(s.SalesCutOffTime) && (s.SeatsAvailable > 0)
}
