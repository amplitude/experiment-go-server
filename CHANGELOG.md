# [1.5.0](https://github.com/amplitude/experiment-go-server/compare/v1.4.0...v1.5.0) (2024-04-19)


### Features

* fetch v2 for remote eval client ([#26](https://github.com/amplitude/experiment-go-server/issues/26)) ([7f7b073](https://github.com/amplitude/experiment-go-server/commit/7f7b0739ebc952533c4832bfdb7659cbd8b2eaa4))

# [1.4.0](https://github.com/amplitude/experiment-go-server/compare/v1.3.2...v1.4.0) (2024-03-07)


### Features

* add method to access flag metadata ([#22](https://github.com/amplitude/experiment-go-server/issues/22)) ([72ae711](https://github.com/amplitude/experiment-go-server/commit/72ae7116a5d54a30e3ef2abe01a09775c8e616f1))

## [1.3.2](https://github.com/amplitude/experiment-go-server/compare/v1.3.1...v1.3.2) (2024-01-29)


### Bug Fixes

* Improve remote evaluation fetch retry logic ([#20](https://github.com/amplitude/experiment-go-server/issues/20)) ([f844a3c](https://github.com/amplitude/experiment-go-server/commit/f844a3c2b1a3358256516708a6b3e1c3b52a1f09))

## [1.3.1](https://github.com/amplitude/experiment-go-server/compare/v1.3.0...v1.3.1) (2024-01-09)


### Bug Fixes

* add nil check when merging local evaluation config ([eedf3e0](https://github.com/amplitude/experiment-go-server/commit/eedf3e0914916901fbd382ba5c9a12807f9e8962))

# [1.3.0](https://github.com/amplitude/experiment-go-server/compare/v1.2.2...v1.3.0) (2023-12-21)


### Bug Fixes

* assignment track defaults and details ([#19](https://github.com/amplitude/experiment-go-server/issues/19)) ([b46239f](https://github.com/amplitude/experiment-go-server/commit/b46239f94b1abfe1e9bcbe5f37acd7b77e70fb55))


### Features

* Golang Local Evaluation V2 ([#18](https://github.com/amplitude/experiment-go-server/issues/18)) ([f0d4504](https://github.com/amplitude/experiment-go-server/commit/f0d4504fb2099287ad5f579630aaaf74b060055d))

## [1.2.2](https://github.com/amplitude/experiment-go-server/compare/v1.2.1...v1.2.2) (2023-09-19)


### Bug Fixes

* Do not track empty assignment events ([#17](https://github.com/amplitude/experiment-go-server/issues/17)) ([fcb0102](https://github.com/amplitude/experiment-go-server/commit/fcb01021f7cf6e86d40f32ec542c0508f1d5efec))

## [1.2.1](https://github.com/amplitude/experiment-go-server/compare/v1.2.0...v1.2.1) (2023-08-29)


### Bug Fixes

* remove unused logging ([#16](https://github.com/amplitude/experiment-go-server/issues/16)) ([1b2153e](https://github.com/amplitude/experiment-go-server/commit/1b2153ee4cd0c2d68be8f5c51c5f9e658a7840b0))

# [1.2.0](https://github.com/amplitude/experiment-go-server/compare/v1.1.3...v1.2.0) (2023-08-25)


### Features

* Automatic assignment Tracking ([#13](https://github.com/amplitude/experiment-go-server/issues/13)) ([96438e5](https://github.com/amplitude/experiment-go-server/commit/96438e5ac0fc091ea322aec02a4058829c859cc2))

## [1.1.3](https://github.com/amplitude/experiment-go-server/compare/v1.1.2...v1.1.3) (2023-08-04)


### Bug Fixes

* evaluate specific flags, topo sort before evaluate ([#14](https://github.com/amplitude/experiment-go-server/issues/14)) ([6b00979](https://github.com/amplitude/experiment-go-server/commit/6b00979857a5bf772d097d838aa581bd3b324eec))

## [1.1.2](https://github.com/amplitude/experiment-go-server/compare/v1.1.1...v1.1.2) (2023-07-19)


### Bug Fixes

* evaluation 1.1.1 fix SIGABRT on json parse error ([#12](https://github.com/amplitude/experiment-go-server/issues/12)) ([326b9a7](https://github.com/amplitude/experiment-go-server/commit/326b9a77e481fadacc17e61e281ef51733eabd3d))

## [1.1.1](https://github.com/amplitude/experiment-go-server/compare/v1.1.0...v1.1.1) (2023-06-13)


### Bug Fixes

* support multiple client instances based on apiKey ([#11](https://github.com/amplitude/experiment-go-server/issues/11)) ([e365735](https://github.com/amplitude/experiment-go-server/commit/e36573555bd672f778607969cb592dcb76a8d368))

# [1.1.0](https://github.com/amplitude/experiment-go-server/compare/v1.0.0...v1.1.0) (2023-03-14)


### Features

* flag dependencies ([#8](https://github.com/amplitude/experiment-go-server/issues/8)) ([64b11bc](https://github.com/amplitude/experiment-go-server/commit/64b11bc1e657d3b2c9ee4e8a0a33132de73b8455))

# [1.0.0](https://github.com/amplitude/experiment-go-server/compare/v0.6.0...v1.0.0) (2022-12-23)


### chore

* fix lint ([#6](https://github.com/amplitude/experiment-go-server/issues/6)) ([dba2b7e](https://github.com/amplitude/experiment-go-server/commit/dba2b7e042a565a286bf902f2adccff43f9c0afe))


### BREAKING CHANGES

* prepare for GA
