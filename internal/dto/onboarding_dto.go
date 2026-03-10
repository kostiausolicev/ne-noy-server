package dto

type OnboardingDto struct {
	ID       string          `json:"id"`
	Platform string          `json:"platform"`
	Path     string          `json:"path"`
	Data     *OnboardingData `json:"data"`
}

type OnboardingData struct {
	Slides []OnboardingSlide `json:"slides"`
}

type OnboardingSlide struct {
	Media    SlideMedia `json:"media"`
	Title    string     `json:"title"`
	Subtitle string     `json:"subtitle"`
}

type SlideMedia struct {
	Type string `json:"type"`
	Blob string `json:"blob"`
}
