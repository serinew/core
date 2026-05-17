# Serinew Core

Serinew **Core HTTP API** 서버입니다.

## 기술 스택

`go.mod` 기준으로 이 프로젝트에서 쓰는 주요 기술입니다.

| 분류 | 기술 | 용도 |
|------|------|------|
| 언어 | **Go** 1.25 | 서버 구현 |
| HTTP 프레임워크 | **[Gin](https://github.com/gin-gonic/gin)** | 라우팅·미들웨어·JSON API |
| 데이터베이스 | **PostgreSQL** | 영구 데이터 저장 |
| ORM / DB 접근 | **[GORM](https://gorm.io/)**, [`gorm.io/driver/postgres`](https://github.com/go-gorm/postgres) | Postgres 연동·모델 매핑 |
| 드라이버 | **[jackc/pgx](https://github.com/jackc/pgx)** v5 | GORM Postgres 드라이버 하위 의존성 |
| 캐시 / 보조 저장 | **Redis**, **[go-redis](https://github.com/redis/go-redis)** v9 | 세션(토큰) 캐시, 레이트리밋 등 |
| 설정 | **[godotenv](https://github.com/joho/godotenv)** | `.env` 로드 |
| API 문서 | **[swag](https://github.com/swaggo/swag)**, **[gin-swagger](https://github.com/swaggo/gin-swagger)** | OpenAPI 주석 → Swagger UI (`/v1/docs`) |
| 검증 | **[go-playground/validator](https://github.com/go-playground/validator)** v10 | 요청 DTO 유효성 검사 |
| 식별자 | **[google/uuid](https://github.com/google/uuid)** | 사용자 등 UUID |
| 보안 | **[golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto)** | 비밀번호 등 크립토(서비스 레이어) |
| 인증 방식 | **HttpOnly 쿠키 + HMAC 서명 토큰** | `accessToken`, `TOKEN_SECRET` (`internal/service/token.service.go`) |
| 컨테이너 | **Docker** (multi-stage, Alpine) | 이미지 빌드·배포 (`Dockerfile`) |

Gin이 내부적으로 **sonic / go-json** 등 JSON 직렬화 라이브러리를 간접 의존합니다.

## 요구사항

- Go **1.25** 이상 (`go.mod` 기준)
- PostgreSQL
- Redis(선택 — 연결 실패 시 로그만 남기고 일부 기능은 인메모리 폴백)

## 로컬 실행

프로젝트 루트에 `.env`를 두면 [godotenv](https://github.com/joho/godotenv)로 자동 로드됩니다. 없으면 프로세스 환경 변수만 사용합니다.

```bash
go run ./cmd/server
```

기본 포트는 **8080** (`PORT` 미설정 시)입니다. 프로덕션 모드는 `NODE_ENV=production` 또는 `GIN_MODE=release`입니다.

## 폴더 구조

| 경로 | 역할 |
|------|------|
| `cmd/server` | HTTP 서버 진입점(`main`), Swagger 메타 주석(`@title`, `@BasePath` 등) |
| `internal/app` | 최상위 라우팅(`router.go`) — `/`, `/v1` 마운트 |
| `internal/app/v1` | API v1 그룹: `route.go`에서 하위 도메인 라우트 조립 |
| `internal/app/v1/<도메인>` | 도메인별 `route.go`(예: `sign`, `users`) |
| `internal/app/v1/<도메인>/<경로파라미터>` | 경로 파라미터 단위(예: `users/userId`) |
| `internal/config` | 환경 변수 로드(`Load`) |
| `internal/middleware` | 복구, 레이트리밋, JWT/쿠키 검증, Swagger 관리자 등 |
| `internal/repository` | DB·Redis 스토어, CRUD 헬퍼 |
| `internal/service` | 비즈니스 로직(회원, 토큰, 사용자 등) |
| `internal/model` | GORM 모델(`core`, `sso` 등) |
| `internal/types` | 요청/응답 DTO, 바인딩·검증, 공통 응답 타입 |
| `internal/util` | HTTP 응답 래퍼 등 공용 유틸 |
| `docs/` | Swagger 생성물(`swag`로 재생성) |

## 환경 변수 예시 (`.env`)

아래는 `internal/config/config.go`에서 읽는 변수입니다. 값은 환경에 맞게 바꾸면 됩니다.

```env
# HTTP
PORT=8080

# PostgreSQL (DSN은 내부에서 조합됨)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=postgres
DB_SSLMODE=disable

# Redis
REDIS_URL=redis://localhost:6379

# accessToken 쿠키 서명용 — 최소 16바이트 이상 권장 (미만이면 로그인 토큰 발급 실패)
TOKEN_SECRET=your-secret-at-least-16-chars

# true면 Set-Cookie에 Secure (HTTPS 배포 시)
COOKIE_SECURE=false

# 프로덕션에서 Gin 릴리스 모드 (둘 중 하나면 됨)
# NODE_ENV=production
# GIN_MODE=release

# 라우트 단위 버스트 레이트리밋 (Redis 권장)
RATE_LIMIT_ENABLED=false
RATE_LIMIT_MAX=120
RATE_LIMIT_WINDOW_SEC=60
# 429 연속 시 같은 토큰으로 추가 시도하면 토큰 무효화(0이면 비활성)
RATE_LIMIT_REVOKE_AFTER_STRIKES=0
```

## HTTP 라우트

**Base URL:** 서버 루트 (예: `http://localhost:8080`)

| 메서드 | 경로 | 인증 | 설명 |
|--------|------|------|------|
| `GET` | `/` | 없음 | 서비스 식별용 짧은 응답 |
| `GET` | `/v1/` | 없음 | v1 그룹 동작 확인 |
| `POST` | `/v1/sign/register` | 없음 | 회원 가입 |
| `POST` | `/v1/sign/login` | 없음 | 로그인 — `accessToken` HttpOnly 쿠키 설정 |
| `POST` | `/v1/sign/password` | 없음 | 비밀번호 변경 |
| `GET` | `/v1/sign`, `/v1/sign/` | `accessToken` 쿠키 | 현재 세션 |
| `GET` | `/v1/users/` | `accessToken` 쿠키 | 사용자 목록(페이지·검색) |
| `GET` | `/v1/users/:userId` | `accessToken` 쿠키 | 사용자 단건 |
| `GET` | `/v1/docs/*` | `accessToken` + **core.admin** | Swagger UI (`/v1/docs` → `index.html` 리다이렉트) |

인증이 필요한 API는 **`accessToken` 쿠키**(이름 상수: `service.AccessTokenCookieName`)를 사용합니다.

## 개발 컨벤션

- **라우트:** 도메인별로 `internal/app/v1/<이름>/route.go`에 `*Routes` 함수로 그룹을 등록합니다. 경로 파라미터가 있는 세그먼트는 별도 폴더(예: `userId`)로 나눕니다.
- **의존성 방향:** `app` → `middleware`, `service`, `repository` — 핸들러는 서비스를 호출하고, 저장소는 `repository`에 둡니다.
- **요청 바디/쿼리:** `internal/types`의 DTO와 `types.BindJSON`, `types.FetchListQuery` 등으로 바인딩합니다.
- **응답:** `internal/types/response.go` 및 `internal/util/http`의 `Http.*` 헬퍼로 상태 코드·본문을 맞춥니다.
- **문서화:** 공개 핸들러에 `swag` 주석(`@Summary`, `@Router`, `@Param` 등)을 유지합니다. 진입점 `@BasePath`는 **`/v1`** 입니다.
- **에러 매핑:** 도메인별로 `errors.Is`로 서비스 에러를 HTTP 상태로 매핑하는 패턴을 따릅니다(예: `sign/route.go`의 `writeSignErr`).

## Swagger (OpenAPI)

- UI: 로그인 후 **`core.admin`** 테이블에 사용자 UUID가 있어야 `/v1/docs`에 접근할 수 있습니다.
- 스펙 재생성 예시 (도구 설치 후):

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go -o docs
```

## Docker

```bash
docker build -t serinew-core .
docker run --env-file .env -p 8080:8080 serinew-core
```

CI/BuildKit 캐시를 쓰는 빌드 스크립트는 `scripts/docker-build.sh`를 참고하세요.
