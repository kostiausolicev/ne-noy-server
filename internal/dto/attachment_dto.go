package dto

// StorageType тип хранилища файла: 1 — база данных, 2 — объектное хранилище.
type StorageType int32

const (
	StorageTypeDatabase      StorageType = 1
	StorageTypeObjectStorage StorageType = 2
)

// UploadFileDTO данные одного загружаемого файла (base64-контент не используется — принимаем bytes через multipart или JSON).
type UploadFileDTO struct {
	StorageType StorageType `json:"storage_type" example:"2"`
	FileName    string      `json:"file_name"    example:"photo.png"`
	FileContent []byte      `json:"file_content"`
}

// PutOneRequest запрос на загрузку одного файла.
type PutOneRequest struct {
	File  UploadFileDTO `json:"file"`
	Async bool          `json:"async" example:"false"`
}

// PutOneResponse ответ с UUID созданного вложения.
type PutOneResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// PutManyRequest запрос на загрузку нескольких файлов.
type PutManyRequest struct {
	Files []UploadFileDTO `json:"files"`
	Async bool            `json:"async" example:"false"`
}

// PutManyResponse ответ со списком UUID загруженных вложений.
type PutManyResponse struct {
	IDs []string `json:"ids"`
}

// GetOneResponse ответ с временной ссылкой на файл.
type GetOneResponse struct {
	URL string `json:"url" example:"https://storage.example.com/file?sign=abc"`
}

// GetManyRequest запрос временных ссылок на несколько файлов.
type GetManyRequest struct {
	IDs []string `json:"ids"`
	TTL int32    `json:"ttl" example:"3600"`
}

// GetManyResponse ответ со списком временных ссылок.
type GetManyResponse struct {
	URLs []string `json:"urls"`
}

// DeleteManyRequest запрос на удаление нескольких файлов.
type DeleteManyRequest struct {
	IDs   []string `json:"ids"`
	Force bool     `json:"force" example:"false"`
}
