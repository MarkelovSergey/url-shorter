package handler

import "github.com/stretchr/testify/mock"

const (
	urlShorterServiceGetOriginalURL = "GetOriginalURL"
	urlShorterServiceGenerate       = "Generate"
)

type MockURLShorterService struct {
	mock.Mock
}

func (m *MockURLShorterService) GetOriginalURL(id string) *string {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}

	result := args.String(0)

	return &result
}

func (m *MockURLShorterService) Generate(url string) string {
	args := m.Called(url)

	return args.String(0)
}
