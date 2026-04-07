package veegozi

type PackageFilm struct {
	FilmID          string
	Title           string
	SplitPercent    float32
	TrailerDuration uint32
	CleanupDuration uint32
	Order           uint32
}

type FilmPackage struct {
	ID    string
	Title string
	Films []PackageFilm
}
type FilmPackageList struct {
	Packages []FilmPackage
}
