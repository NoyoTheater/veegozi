package veegozi

import "context"

type Screen struct {
	ID              string
	Name            string
	ScreenNumber    string
	HasCustomLayout bool
	TotalSeats      uint32
	HouseSeats      uint32
}
type ScreenList struct {
	Screens []Screen
}

func (s Screen) Sessions(ctx context.Context, c *Client) (SessionList, error) {
	sessions, err := c.GetSessions(ctx)
	if err != nil {
		return SessionList{}, err
	}

	return sessions.FilterByScreen(s.ID), nil
}
