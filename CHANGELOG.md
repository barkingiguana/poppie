# Changelog

## 1.0.0 (2026-03-14)


### Features

* add --live flag to get command with animated multi-label display ([6d940b3](https://github.com/barkingiguana/poppie/commit/6d940b3bfd7f610ec9a444b6565a6f7a95c38e4b))
* add BDD feature specs with godog step definitions ([4fe2360](https://github.com/barkingiguana/poppie/commit/4fe2360bfe0a5461d26197792403eee87f0a51a9))
* add ephemeral in-memory vault and custom vault path options ([47ecd3b](https://github.com/barkingiguana/poppie/commit/47ecd3b53f03211a49e845647b8dde870eede67e))
* add Go and Python SDKs with version negotiation ([e7a66d3](https://github.com/barkingiguana/poppie/commit/e7a66d31fd72a8c8f41da2dcd0ed866c933f6ba9))
* build CLI with cobra — store, get, list, delete, and server commands ([64583e8](https://github.com/barkingiguana/poppie/commit/64583e8a0f209d6fa78fc39f6c4d5faa20ec1b80))
* define protobuf service and messages for TOTP API ([04044c1](https://github.com/barkingiguana/poppie/commit/04044c174ae738dc47b904a298065f412cb3a7e0))
* implement encrypted vault storage with AES-256-GCM ([ef568a8](https://github.com/barkingiguana/poppie/commit/ef568a8a8b97a6b0af3a168bc7bd4fdfd9112d58))
* implement gRPC server with all TOTP operations ([d525ebc](https://github.com/barkingiguana/poppie/commit/d525ebccfc4a38fc6c5279be1123ff976334720f))
* implement RFC 6238 TOTP engine with validation ([b42404c](https://github.com/barkingiguana/poppie/commit/b42404cdb474c3c45eafdf54455f4eddfc67b0f5))
* initial project setup from claude-template ([bdbf3eb](https://github.com/barkingiguana/poppie/commit/bdbf3eb506af1ddc1927e5219071e5044ef54d22))
* set up GitHub Pages with Jekyll and just-the-docs theme ([24f580d](https://github.com/barkingiguana/poppie/commit/24f580d190c250793fa0d448f16b1c851effa76a))
* update CI pipeline and add dm integration guide ([df76c47](https://github.com/barkingiguana/poppie/commit/df76c47eedd003b8b34d7e2b051576dcdfa97200))


### Bug Fixes

* CI coverage now measures cross-package and hits 80% threshold ([246e1d8](https://github.com/barkingiguana/poppie/commit/246e1d8c9d83f6227d82db8d202f0128b62ed7d7))
* handle all error returns flagged by golangci-lint ([804a35f](https://github.com/barkingiguana/poppie/commit/804a35fd912d7c1bab3557e15c9ff78443427785))
* use custom domain for GitHub Pages URL ([5329fda](https://github.com/barkingiguana/poppie/commit/5329fda9d484ba155bf2cb49723e15c1261032b8))
