# lm-cli — LINE Messaging API CLI

LINE Messaging API를 위한 커맨드라인 인터페이스입니다. DevOps 엔지니어, 봇 개발자, 마케터가
터미널에서 LINE 공식 계정을 관리할 수 있습니다.

## 특징

- **3가지 인증 모드**: 장기 토큰, 스테이트리스 (CI/CD 친화적), JWT v2
- **모든 메시지 타입**: push, multicast, broadcast, narrowcast, reply
- **출력 형식**: table / json / yaml / csv
- **CI/CD 친화적**: `--no-input`, `LM_TOKEN` 환경 변수, 명확한 종료 코드
- **다중 프로파일**: 프로덕션/스테이징/테스트 환경 관리

## 설치

```bash
# Go로 설치
go install github.com/crowdy/lm-cli@latest

# 바이너리 (GitHub Releases)
# https://github.com/crowdy/lm-cli/releases
```

## 빠른 시작

```bash
# 인증 (장기 토큰)
lm auth login

# 봇 정보 확인
lm bot info

# 메시지 전송
lm message push <userId> "안녕하세요!"

# JSON 출력
lm bot info --format json

# CI/CD에서 환경 변수 사용
LM_TOKEN=xxx lm message push <userId> "배포 완료"
```

## 명령어 목록

| 명령어 | 설명 |
|--------|------|
| `lm auth login` | 인증 설정 |
| `lm auth status` | 인증 상태 확인 |
| `lm message push <userId> <text>` | 사용자에게 메시지 전송 |
| `lm message broadcast <text>` | 모든 팔로워에게 전송 |
| `lm bot info` | 봇 프로파일 조회 |
| `lm webhook get` | Webhook URL 조회 |
| `lm webhook set <url>` | Webhook URL 설정 |
| `lm richmenu list` | 리치 메뉴 목록 |
| `lm insight followers` | 팔로워 통계 |

## 환경 변수

| 변수 | 설명 |
|------|------|
| `LM_TOKEN` | 직접 토큰 지정 (인증 우회) |
| `LM_CHANNEL_ID` | 채널 ID |
| `LM_CHANNEL_SECRET` | 채널 시크릿 |
| `LM_FORMAT` | 출력 형식 (table/json/yaml/csv) |
| `LM_NO_INPUT` | 비대화형 모드 (1 또는 true) |

## 종료 코드

| 코드 | 의미 |
|------|------|
| 0 | 성공 |
| 1 | 일반 오류 |
| 2 | 인증 오류 (401/403) |
| 3 | 리소스 없음 (404) |
| 4 | 유효성 검사 오류 |
| 5 | LINE API 오류 |
| 6 | 네트워크 오류 |
| 11 | 속도 제한 (429) |

## 라이선스

MIT
