package veegozi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

type Client struct {
	base                 *url.URL
	token                string
	http                 *http.Client
	sessionCache         *Cache[Session]
	sessionListCache     *Cache[SessionList]
	webSessionListCache  *Cache[SessionList]
	filmCache            *Cache[Film]
	filmListCache        *Cache[FilmList]
	filmPackageCache     *Cache[FilmPackage]
	filmPackageListCache *Cache[FilmPackageList]
	screenCache          *Cache[Screen]
	screenListCache      *Cache[ScreenList]
	attributeCache       *Cache[Attribute]
	attributeListCache   *Cache[AttributeList]
	siteCache            *Cache[Site]
}
type Attribute struct {
	ID                        string
	ShortName                 string
	Description               string
	FontColor                 string
	BackgroundColor           string
	ShowOnSessionsWithNoComps bool
}
type AttributeList struct {
	Attributes []Attribute
}

type ClientOption func(*Client)

func WithHTTPClient(http *http.Client) ClientOption {
	return func(c *Client) {
		c.http = http
	}
}

func WithSessionCache(ttl time.Duration, capacity int64) ClientOption {
	return func(c *Client) {
		c.sessionCache = newCache[Session](ttl, capacity)
		c.sessionListCache = newCache[SessionList](ttl, 1)
		c.webSessionListCache = newCache[SessionList](ttl, 1)
	}
}

func WithFilmCache(ttl time.Duration, capacity int64) ClientOption {
	return func(c *Client) {
		c.filmCache = newCache[Film](ttl, capacity)
		c.filmListCache = newCache[FilmList](ttl, 1)
	}
}

func WithFilmPackageCache(ttl time.Duration, capacity int64) ClientOption {
	return func(c *Client) {
		c.filmPackageCache = newCache[FilmPackage](ttl, capacity)
		c.filmPackageListCache = newCache[FilmPackageList](ttl, 1)
	}
}

func WithScreenCache(ttl time.Duration, capacity int64) ClientOption {
	return func(c *Client) {
		c.screenCache = newCache[Screen](ttl, capacity)
		c.screenListCache = newCache[ScreenList](ttl, 1)
	}
}

func WithAttributeCache(ttl time.Duration, capacity int64) ClientOption {
	return func(c *Client) {
		c.attributeCache = newCache[Attribute](ttl, capacity)
		c.attributeListCache = newCache[AttributeList](ttl, 1)
	}
}

func WithSiteCache(ttl time.Duration) ClientOption {
	return func(c *Client) {
		c.siteCache = newCache[Site](ttl, 1)
	}
}

func WithDefaultCaching() ClientOption {
	return func(c *Client) {
		WithSessionCache(30*time.Second, 1000)(c)
		WithFilmCache(5*time.Minute, 500)(c)
		WithFilmPackageCache(5*time.Minute, 500)(c)
		WithScreenCache(1*time.Hour, 100)(c)
		WithAttributeCache(5*time.Minute, 500)(c)
		WithSiteCache(5 * time.Minute)(c)
	}
}

// Create a new Client with the given base URL and token, applying any provided options.
func NewClient(base, token string, opts ...ClientOption) *Client {
	url, err := url.Parse(base)
	if err != nil {
		panic("invalid base URL: " + err.Error())
	}

	client := &Client{
		base:  url,
		token: token,
		http:  http.DefaultClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// getJSON makes a GET request to the Veezi API and parses the JSON response into the provided
// struct. The endpoint should be a path relative to the base URL.
//
// Returns an error if the request fails or if the response cannot be parsed.
func (c *Client) getJSON(ctx context.Context, endpoint string, dest any) error {
	url := c.base.JoinPath(endpoint).String()

	log := zerolog.Ctx(ctx).With().Str("url", url).Str("method", "GET").Logger()
	ctx = log.WithContext(ctx)

	log.Debug().Msg("making API request")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to create request")
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("VeeziAccessToken", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("API request failed")
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status_code", resp.StatusCode).Msg("API request returned non-200 status")
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		log.Error().Err(err).Msg("failed to decode API response")
		return fmt.Errorf("decode response: %w", err)
	}

	log.Debug().Msg("Request completed")
	return nil
}

func (c *Client) InvalidateCachedSession(ctx context.Context, sessionID string) {
	if c.sessionCache != nil {
		c.sessionCache.Delete(ctx, sessionID)
		c.sessionListCache.Clear(ctx)
		c.webSessionListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateAllCachedSessions(ctx context.Context) {
	if c.sessionCache != nil {
		c.sessionCache.Clear(ctx)
		c.sessionListCache.Clear(ctx)
		c.webSessionListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateCachedFilm(ctx context.Context, filmID string) {
	if c.filmCache != nil {
		c.filmCache.Delete(ctx, filmID)
		c.filmListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateAllCachedFilms(ctx context.Context) {
	if c.filmCache != nil {
		c.filmCache.Clear(ctx)
		c.filmListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateCachedFilmPackage(ctx context.Context, packageID string) {
	if c.filmPackageCache != nil {
		c.filmPackageCache.Delete(ctx, packageID)
		c.filmPackageListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateAllCachedFilmPackages(ctx context.Context) {
	if c.filmPackageCache != nil {
		c.filmPackageCache.Clear(ctx)
		c.filmPackageListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateCachedScreen(ctx context.Context, screenID string) {
	if c.screenCache != nil {
		c.screenCache.Delete(ctx, screenID)
		c.screenListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateAllCachedScreens(ctx context.Context) {
	if c.screenCache != nil {
		c.screenCache.Clear(ctx)
		c.screenListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateCachedAttribute(ctx context.Context, attributeID string) {
	if c.attributeCache != nil {
		c.attributeCache.Delete(ctx, attributeID)
		c.attributeListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateAllCachedAttributes(ctx context.Context) {
	if c.attributeCache != nil {
		c.attributeCache.Clear(ctx)
		c.attributeListCache.Clear(ctx)
	}
}

func (c *Client) InvalidateCachedSite(ctx context.Context) {
	if c.siteCache != nil {
		c.siteCache.Clear(ctx)
	}
}

func (c *Client) InvalidateAllCaches(ctx context.Context) {
	if c.sessionCache != nil {
		c.sessionCache.Clear(ctx)
		c.sessionListCache.Clear(ctx)
		c.webSessionListCache.Clear(ctx)
	}
	if c.filmCache != nil {
		c.filmCache.Clear(ctx)
		c.filmListCache.Clear(ctx)
	}
	if c.filmPackageCache != nil {
		c.filmPackageCache.Clear(ctx)
		c.filmPackageListCache.Clear(ctx)
	}
	if c.screenCache != nil {
		c.screenCache.Clear(ctx)
		c.screenListCache.Clear(ctx)
	}
	if c.attributeCache != nil {
		c.attributeCache.Clear(ctx)
		c.attributeListCache.Clear(ctx)
	}
	if c.siteCache != nil {
		c.siteCache.Clear(ctx)
	}
}

// Get a list of all future sessions.
func (c *Client) GetSessions(ctx context.Context) (SessionList, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetSessions").Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (SessionList, error) {
		var sessions []Session
		if err := c.getJSON(ctx, "v1/session", &sessions); err != nil {
			return SessionList{}, err
		}
		return SessionList{Sessions: sessions}, nil
	}

	if c.sessionListCache == nil {
		return fetchRaw()
	}

	if cached, found := c.sessionListCache.Get(ctx, ""); found {
		log.Debug().Msg("returning sessions from cache")
		return cached, nil
	}

	sessions, err := fetchRaw()
	if err != nil {
		return SessionList{}, err
	}

	c.sessionListCache.Set(ctx, "", sessions)
	for _, s := range sessions.Sessions {
		c.sessionCache.Set(ctx, s.ID, s)
	}
	return sessions, nil
}

// Get a list of all future sessions that are available for online booking.
func (c *Client) GetWebSessions(ctx context.Context) (SessionList, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetWebSessions").Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (SessionList, error) {
		var sessions []Session
		if err := c.getJSON(ctx, "v1/websession", &sessions); err != nil {
			return SessionList{}, err
		}
		return SessionList{Sessions: sessions}, nil
	}

	if c.webSessionListCache == nil {
		return fetchRaw()
	}

	if cached, found := c.webSessionListCache.Get(ctx, ""); found {
		log.Debug().Msg("returning web sessions from cache")
		return cached, nil
	}

	sessions, err := fetchRaw()
	if err != nil {
		return SessionList{}, err
	}

	c.webSessionListCache.Set(ctx, "", sessions)
	for _, s := range sessions.Sessions {
		c.sessionCache.Set(ctx, s.ID, s)
	}
	return sessions, nil
}

// Get a specific session by ID.
func (c *Client) GetSession(ctx context.Context, sessionID string) (Session, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetSession").Str("session_id", sessionID).Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (Session, error) {
		var session Session
		if err := c.getJSON(ctx, fmt.Sprintf("v1/session/%s", sessionID), &session); err != nil {
			return Session{}, err
		}
		return session, nil
	}

	if c.sessionCache == nil {
		return fetchRaw()
	}

	if cached, found := c.sessionCache.Get(ctx, sessionID); found {
		log.Debug().Msg("returning session from cache")
		return cached, nil
	}

	session, err := fetchRaw()
	if err != nil {
		return Session{}, err
	}

	c.sessionCache.Set(ctx, sessionID, session)
	return session, nil
}

// Get a list of all films.
func (c *Client) GetFilms(ctx context.Context) (FilmList, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilms").Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (FilmList, error) {
		var films []Film
		if err := c.getJSON(ctx, "v4/film", &films); err != nil {
			return FilmList{}, err
		}
		return FilmList{Films: films}, nil
	}

	if c.filmListCache == nil {
		return fetchRaw()
	}

	if cached, found := c.filmListCache.Get(ctx, ""); found {
		log.Debug().Msg("returning films from cache")
		return cached, nil
	}

	films, err := fetchRaw()
	if err != nil {
		return FilmList{}, err
	}

	c.filmListCache.Set(ctx, "", films)
	for _, f := range films.Films {
		c.filmCache.Set(ctx, f.ID, f)
	}
	return films, nil
}

// Get a specific film by ID.
func (c *Client) GetFilm(ctx context.Context, filmID string) (Film, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilm").Str("film_id", filmID).Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (Film, error) {
		var film Film
		if err := c.getJSON(ctx, fmt.Sprintf("v4/film/%s", filmID), &film); err != nil {
			return Film{}, err
		}
		return film, nil
	}

	if c.filmCache == nil {
		return fetchRaw()
	}

	if cached, found := c.filmCache.Get(ctx, filmID); found {
		log.Debug().Msg("returning film from cache")
		return cached, nil
	}

	film, err := fetchRaw()
	if err != nil {
		return Film{}, err
	}

	c.filmCache.Set(ctx, filmID, film)
	return film, nil
}

// Get a specific film by its exact title. if multiple films have the same title, the first one
// returned by the API will be returned.
func (c *Client) GetFilmByTitle(ctx context.Context, title string) (Film, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmByTitle").Str("title", title).Logger()
	ctx = log.WithContext(ctx)

	films, err := c.GetFilms(ctx)
	if err != nil {
		return Film{}, err
	}

	for _, film := range films.Films {
		if film.Title == title {
			log.Debug().Msg("found film by title")
			return film, nil
		}
	}

	log.Debug().Msg("no film found with given title")
	return Film{}, fmt.Errorf("film not found with title: %s", title)
}

// Get a specific film by its exact ShortName. if multiple films have the same ShortName, the first
// one returned by the API will be returned.
func (c *Client) GetFilmByShortName(ctx context.Context, shortName string) (Film, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmByShortName").Str("short_name", shortName).Logger()
	ctx = log.WithContext(ctx)

	films, err := c.GetFilms(ctx)
	if err != nil {
		return Film{}, err
	}

	for _, film := range films.Films {
		if film.ShortName == shortName {
			log.Debug().Msg("found film by short name")
			return film, nil
		}
	}

	log.Debug().Msg("no film found with given short name")
	return Film{}, fmt.Errorf("film not found with short name: %s", shortName)
}

// Get a specific film by its exact SignageText. if multiple films have the same SignageText, the first one
// returned by the API will be returned.
func (c *Client) GetFilmBySignageText(ctx context.Context, signageText string) (Film, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmBySignageText").Str("signage_text", signageText).Logger()
	ctx = log.WithContext(ctx)

	films, err := c.GetFilms(ctx)
	if err != nil {
		return Film{}, err
	}

	for _, film := range films.Films {
		if film.SignageText == signageText {
			log.Debug().Msg("found film by signage text")
			return film, nil
		}
	}

	log.Debug().Msg("no film found with given signage text")
	return Film{}, fmt.Errorf("film not found with signage text: %s", signageText)
}

// List films that match the given genre.
func (c *Client) GetFilmsByGenre(ctx context.Context, genre string) ([]Film, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmsByGenre").Str("genre", genre).Logger()
	ctx = log.WithContext(ctx)

	films, err := c.GetFilms(ctx)
	if err != nil {
		return nil, err
	}

	var matching []Film
	for _, film := range films.Films {
		if film.Genre == genre {
			matching = append(matching, film)
		}
	}

	log.Debug().Int("matching_count", len(matching)).Msg("found films by genre")
	return matching, nil
}

// List films that match the given distributor.
func (c *Client) GetFilmsByDistributor(ctx context.Context, distributor string) ([]Film, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmsByDistributor").Str("distributor", distributor).Logger()
	ctx = log.WithContext(ctx)

	films, err := c.GetFilms(ctx)
	if err != nil {
		return nil, err
	}

	var matching []Film
	for _, film := range films.Films {
		if film.Distributor == distributor {
			matching = append(matching, film)
		}
	}

	log.Debug().Int("matching_count", len(matching)).Msg("found films by distributor")
	return matching, nil
}

func (c *Client) ListFilmsWithSessionsInTimeRange(ctx context.Context, start, end time.Time) (FilmList, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "ListFilmsWithSessionsInTimeRange").Time("start", start).Time("end", end).Logger()
	ctx = log.WithContext(ctx)

	sessions, err := c.GetSessions(ctx)
	if err != nil {
		return FilmList{}, err
	}

	filmMap := make(map[string]Film)
	for _, session := range sessions.Sessions {
		if session.FeatureStartTime.After(start) && session.FeatureEndTime.Before(end) {
			film, err := session.Film(ctx, c)
			if err != nil {
				return FilmList{}, err
			}
			filmMap[film.ID] = film
		}
	}

	var films []Film
	for _, film := range filmMap {
		films = append(films, film)
	}
	log.Debug().Int("matching_count", len(films)).Msg("found films with sessions in time range")
	return FilmList{Films: films}, nil
}

// List film packages.
func (c *Client) GetFilmPackages(ctx context.Context) (FilmPackageList, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmPackages").Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (FilmPackageList, error) {
		var packages []FilmPackage
		if err := c.getJSON(ctx, "v1/filmpackage", &packages); err != nil {
			return FilmPackageList{}, err
		}
		return FilmPackageList{Packages: packages}, nil
	}

	if c.filmPackageListCache == nil {
		return fetchRaw()
	}

	if cached, found := c.filmPackageListCache.Get(ctx, ""); found {
		log.Debug().Msg("returning film packages from cache")
		return cached, nil
	}

	packages, err := fetchRaw()
	if err != nil {
		return FilmPackageList{}, err
	}

	c.filmPackageListCache.Set(ctx, "", packages)
	for _, p := range packages.Packages {
		c.filmPackageCache.Set(ctx, p.ID, p)
	}
	return packages, nil
}

// Get a specific film package by ID.
func (c *Client) GetFilmPackage(ctx context.Context, packageID string) (FilmPackage, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmPackage").Str("package_id", packageID).Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (FilmPackage, error) {
		var pkg FilmPackage
		if err := c.getJSON(ctx, fmt.Sprintf("v1/filmpackage/%s", packageID), &pkg); err != nil {
			return FilmPackage{}, err
		}
		return pkg, nil
	}

	if c.filmPackageCache == nil {
		return fetchRaw()
	}

	if cached, found := c.filmPackageCache.Get(ctx, packageID); found {
		log.Debug().Msg("returning film package from cache")
		return cached, nil
	}

	pkg, err := fetchRaw()
	if err != nil {
		return FilmPackage{}, err
	}

	c.filmPackageCache.Set(ctx, packageID, pkg)
	return pkg, nil
}

// Get a specific film package by title. if multiple film packages have the same title, the first
// one returned by the API will be returned.
func (c *Client) GetFilmPackageByTitle(ctx context.Context, title string) (FilmPackage, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmPackageByTitle").Str("title", title).Logger()
	ctx = log.WithContext(ctx)

	packages, err := c.GetFilmPackages(ctx)
	if err != nil {
		return FilmPackage{}, err
	}

	for _, pkg := range packages.Packages {
		if pkg.Title == title {
			log.Debug().Msg("found film package by title")
			return pkg, nil
		}
	}

	log.Debug().Msg("no film package found with given title")
	return FilmPackage{}, fmt.Errorf("film package not found with title: %s", title)
}

// Get a list of all film packages containing a specific film ID.
func (c *Client) GetFilmPackagesByFilmID(ctx context.Context, filmID string) ([]FilmPackage, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetFilmPackagesByFilmID").Str("film_id", filmID).Logger()
	ctx = log.WithContext(ctx)

	packages, err := c.GetFilmPackages(ctx)
	if err != nil {
		return nil, err
	}

	var matching []FilmPackage
	for _, pkg := range packages.Packages {
		for _, film := range pkg.Films {
			if film.FilmID == filmID {
				matching = append(matching, pkg)
				break
			}
		}
	}

	log.Debug().Int("matching_count", len(matching)).Msg("found film packages by film ID")
	return matching, nil
}

// List all screens in this site
func (c *Client) GetScreens(ctx context.Context) (ScreenList, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetScreens").Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (ScreenList, error) {
		var screens []Screen
		if err := c.getJSON(ctx, "v1/screen", &screens); err != nil {
			return ScreenList{}, err
		}
		return ScreenList{Screens: screens}, nil
	}

	if c.screenListCache == nil {
		return fetchRaw()
	}

	if cached, found := c.screenListCache.Get(ctx, ""); found {
		log.Debug().Msg("returning screens from cache")
		return cached, nil
	}

	screens, err := fetchRaw()
	if err != nil {
		return ScreenList{}, err
	}

	c.screenListCache.Set(ctx, "", screens)
	for _, s := range screens.Screens {
		c.screenCache.Set(ctx, s.ID, s)
	}
	return screens, nil
}

// Get a specific screen by ID.
func (c *Client) GetScreen(ctx context.Context, screenID string) (Screen, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetScreen").Str("screen_id", screenID).Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (Screen, error) {
		var screen Screen
		if err := c.getJSON(ctx, fmt.Sprintf("v1/screen/%s", screenID), &screen); err != nil {
			return Screen{}, err
		}
		return screen, nil
	}

	if c.screenCache == nil {
		return fetchRaw()
	}

	if cached, found := c.screenCache.Get(ctx, screenID); found {
		log.Debug().Msg("returning screen from cache")
		return cached, nil
	}

	screen, err := fetchRaw()
	if err != nil {
		return Screen{}, err
	}

	c.screenCache.Set(ctx, screenID, screen)
	return screen, nil
}

// Get a specific screen by its number.
func (c *Client) GetScreenByNumber(ctx context.Context, screenNumber string) (Screen, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetScreenByNumber").Str("screen_number", screenNumber).Logger()
	ctx = log.WithContext(ctx)

	screens, err := c.GetScreens(ctx)
	if err != nil {
		return Screen{}, err
	}

	for _, screen := range screens.Screens {
		if screen.ScreenNumber == screenNumber {
			log.Debug().Msg("found screen by number")
			return screen, nil
		}
	}

	log.Debug().Msg("no screen found with given number")
	return Screen{}, fmt.Errorf("screen not found with number: %s", screenNumber)
}

// Get the current site.
func (c *Client) GetSite(ctx context.Context) (Site, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetSite").Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (Site, error) {
		var site Site
		if err := c.getJSON(ctx, "v1/site", &site); err != nil {
			return Site{}, err
		}
		return site, nil
	}

	if c.siteCache == nil {
		return fetchRaw()
	}

	if cached, found := c.siteCache.Get(ctx, ""); found {
		log.Debug().Msg("returning site from cache")
		return cached, nil
	}

	site, err := fetchRaw()
	if err != nil {
		return Site{}, err
	}

	c.siteCache.Set(ctx, "", site)
	return site, nil
}

// Get a list of all attributes.
func (c *Client) GetAttributes(ctx context.Context) (AttributeList, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetAttributes").Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (AttributeList, error) {
		var attributes []Attribute
		if err := c.getJSON(ctx, "v1/attribute", &attributes); err != nil {
			return AttributeList{}, err
		}
		return AttributeList{Attributes: attributes}, nil
	}

	if c.attributeListCache == nil {
		return fetchRaw()
	}

	if cached, found := c.attributeListCache.Get(ctx, ""); found {
		log.Debug().Msg("returning attributes from cache")
		return cached, nil
	}

	attributes, err := fetchRaw()
	if err != nil {
		return AttributeList{}, err
	}

	c.attributeListCache.Set(ctx, "", attributes)
	for _, a := range attributes.Attributes {
		c.attributeCache.Set(ctx, a.ID, a)
	}
	return attributes, nil
}

// Get a specific attribute by ID.
func (c *Client) GetAttribute(ctx context.Context, attributeID string) (Attribute, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetAttribute").Str("attribute_id", attributeID).Logger()
	ctx = log.WithContext(ctx)

	fetchRaw := func() (Attribute, error) {
		var attribute Attribute
		if err := c.getJSON(ctx, fmt.Sprintf("v1/attribute/%s", attributeID), &attribute); err != nil {
			return Attribute{}, err
		}
		return attribute, nil
	}

	if c.attributeCache == nil {
		return fetchRaw()
	}

	if cached, found := c.attributeCache.Get(ctx, attributeID); found {
		log.Debug().Msg("returning attribute from cache")
		return cached, nil
	}

	attribute, err := fetchRaw()
	if err != nil {
		return Attribute{}, err
	}

	c.attributeCache.Set(ctx, attributeID, attribute)
	return attribute, nil
}

// Get a specific attribute by its ShortName. if multiple attributes have the same ShortName,
// the first one returned by the API will be returned.
func (c *Client) GetAttributeByShortName(ctx context.Context, shortName string) (Attribute, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetAttributeByShortName").Str("short_name", shortName).Logger()
	ctx = log.WithContext(ctx)

	attributes, err := c.GetAttributes(ctx)
	if err != nil {
		return Attribute{}, err
	}

	for _, attribute := range attributes.Attributes {
		if attribute.ShortName == shortName {
			log.Debug().Msg("found attribute by short name")
			return attribute, nil
		}
	}

	log.Debug().Msg("no attribute found with given short name")
	return Attribute{}, fmt.Errorf("attribute not found with short name: %s", shortName)
}

// Get a specific attribute by its description. if multiple attributes have the same description,
// the first one returned by the API will be returned.
func (c *Client) GetAttributeByDescription(ctx context.Context, description string) (Attribute, error) {
	log := zerolog.Ctx(ctx).With().Str("method", "GetAttributeByDescription").Str("description", description).Logger()
	ctx = log.WithContext(ctx)

	attributes, err := c.GetAttributes(ctx)
	if err != nil {
		return Attribute{}, err
	}

	for _, attribute := range attributes.Attributes {
		if attribute.Description == description {
			log.Debug().Msg("found attribute by description")
			return attribute, nil
		}
	}

	log.Debug().Msg("no attribute found with given description")
	return Attribute{}, fmt.Errorf("attribute not found with description: %s", description)
}
