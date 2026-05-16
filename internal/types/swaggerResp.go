package types

import "github.com/serinew/core/internal/model/core"

// 아래 타입들은 OpenAPI(Swagger) 스키마 전용입니다.
// 실제 HTTP 응답은 SuccessDoc/Created 등과 같은 필드명을 갖지만, data 에 구체 타입을 두어 문서에서 펼쳐 보입니다.

type SuccessEnvelopeStr struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       string `json:"data"`
}

type SuccessEnvelopeUser struct {
	Status     string    `json:"status"`
	StatusCode int       `json:"statusCode"`
	Message    string    `json:"message"`
	Data       core.User `json:"data"`
}

type SuccessEnvelopeSignLogin struct {
	Status     string        `json:"status"`
	StatusCode int           `json:"statusCode"`
	Message    string        `json:"message"`
	Data       SignLoginData `json:"data"`
}

type SuccessEnvelopeSignLoginList struct {
	Status     string          `json:"status"`
	StatusCode int             `json:"statusCode"`
	Message    string          `json:"message"`
	Count      *int            `json:"count,omitempty"`
	Data       []SignLoginData `json:"data"`
}

type SuccessEnvelopeMsg struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}
