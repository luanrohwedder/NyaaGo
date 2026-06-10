package models

type (
	Anime struct {
		Name     string
		Filepath string
		Episodes []Episode
	}

	Episode struct {
		number string
	}
)
