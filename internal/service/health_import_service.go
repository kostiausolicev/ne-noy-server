package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"time"

	"github.com/google/uuid"
)

type healthImportService struct {
	d zip.Decompressor
}

func (h healthImportService) SaveAppleMetadata(ctx context.Context, userVkId int64, eventId uuid.UUID, records *dto.UserActivitiesInfo) error {
	return errors.New("save apple metadata is not implemented")
}

func (h healthImportService) ParceAppleMetadataZip(
	ctx context.Context,
	userVkId int64,
	archive dto.AppleArchiveZipDto,
) (*dto.UserActivitiesInfo, error) {
	b := archive.Archive
	r, _ := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	exportFile := r.File[0]
	rc, _ := exportFile.Open()
	defer rc.Close()
	var hd model.HealthData
	var res []byte
	for {
		buf := make([]byte, 1024)
		n, err := rc.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, buf[:n]...)
	}
	err := xml.Unmarshal(res, &hd)
	if err != nil {
		return nil, err
	}

	d := &dto.UserActivitiesInfo{
		User: nil,
		Activities: []dto.ActivityInfo{
			{
				Activity: "as_test",
				Starts:   time.Now(),
				Ends:     time.Now(),
			},
		},
	}
	return d, nil
}

func (h healthImportService) GetUserActivities(ctx context.Context, userVkId int64, eventId uuid.UUID) ([]dto.UserActivitiesInfo, error) {
	return nil, errors.New("get user activities is not implemented")
}

type HealthImportService interface {
	ParceAppleMetadataZip(ctx context.Context, userVkId int64, archive dto.AppleArchiveZipDto) (*dto.UserActivitiesInfo, error)
	SaveAppleMetadata(ctx context.Context, userVkId int64, eventId uuid.UUID, records *dto.UserActivitiesInfo) error
	GetUserActivities(ctx context.Context, userVkId int64, eventId uuid.UUID) ([]dto.UserActivitiesInfo, error)
}

func NewHealthImportService() HealthImportService {
	return &healthImportService{}
}
