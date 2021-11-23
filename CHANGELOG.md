## Changelog

### 1.7.0
#### Updated
- Updated Chromium to 93.0.4577.0
- Updated Ferret to 0.16.0
- Updated other dependencies

### 1.6.0
#### Updated
- Updated dependencies
- Updated install.sh script
- Updated license badge

### 1.5.0
#### Added
- "version" command
- "attempts" flag

#### Updated
- Ferret to v0.14.1
- Improved time formatting

#### Fixed
- Invalid execution time

### 1.4.0
#### Updated
- Ferret to v0.14.0
- Echo to v4.2.0
- Chromium to v90.0.4427.0

### 1.3.1
#### Fixed
- Files resolution. [#8](https://github.com/MontFerret/lab/pull/8)

### 1.3.0
#### Added
- Added interval time between test cycles. [#6](https://github.com/MontFerret/lab/pull/6)

#### Updated
- Upgraded Ferret to v0.13.0

### 1.2.0
#### Added
- Automatic port allocation. [#4](https://github.com/MontFerret/lab/pull/4)

#### Changed
- Updated dependencies. [#5](https://github.com/MontFerret/lab/pull/5)

### 1.1.1
#### Changed
- Added custom data serialization. [53b638b](https://github.com/MontFerret/lab/commit/53b638bcdd1db745bf738bc2ad1cd387cec9e5df)

### 1.1.0
#### Changed
- Updated Chromium. [e295913](https://github.com/MontFerret/lab/commit/e29591318df984f99e3a6d8a42c307c029c7f0c8)
- Updated sources. [5f0ab56](https://github.com/MontFerret/lab/commit/5f0ab567ee078d955d3f83b97aedb379d1ad9e36)
- Updated workflow. [8f3c76e](https://github.com/MontFerret/lab/commit/8f3c76ecd5a6740da50acbc2bad9807605ee3257)
- Upgraded Ferret to 0.12.1. [0c4f643](https://github.com/MontFerret/lab/commit/0c4f643abe4688b85279504656bcb4d87c56b968)

### 0.7.0
#### Changed
- Updated CLI param names. [87ad185](https://github.com/MontFerret/lab/commit/87ad185b75bdd38f06f13c3919c86cf176135dc9)
- Updated Chromium version in generic Dockerfile. [6ec30f0](https://github.com/MontFerret/lab/commit/6ec30f0e8f224a2b5080009b674ee5a6d48428ef)

### 0.6.0
#### Added
- Support of timeout in .yaml manifest. [7f15d88](https://github.com/MontFerret/lab/commit/7f15d8854bbf06f8500955ea31edd1f00c8eff74)

#### Changed
- Removed --force-gpu-mem-available-mb flag. [d8d58eb](https://github.com/MontFerret/lab/commit/d8d58ebecb0835a3a3634d7bb2bebef9ef397240)
- Updated dockerfile config in .goreleaser. [7e79daf](https://github.com/MontFerret/lab/commit/7e79daff312ca9be22b0a45cde92c86c9baf614a)

### 0.5.0
#### Added
- Access to query context. [fcbf1c6](https://github.com/MontFerret/lab/commit/fcbf1c6f00ed65b97804906e2243092cb5a32d4c)
- Support of suite tests. [ef46ac4](https://github.com/MontFerret/lab/commit/ef46ac4fafdaa3fec6afbcb4cc9bcb0c0d55eb73)
- Timeout for tests. [cd8ef78](https://github.com/MontFerret/lab/commit/cd8ef78a43f99cc1fc62362cf2e91b4d0c12742a)
- Pass cdp address to binary runtime. [ba7cf7e](https://github.com/MontFerret/lab/commit/ba7cf7ec6b3f135de18e313fa9bab1e51617bee3)

#### Fixed
- Outbound IP address retrieval. [40cd9e6](https://github.com/MontFerret/lab/commit/40cd9e655c0a4398b5df142017f9626897be5327)
- Passing multiple args to external binary runtime. [c34dbac](https://github.com/MontFerret/lab/commit/c34dbacc6d131f8d52465eda3ae0025188219cf6)
- Use of invalid file in for loop closure. [ebaac08](https://github.com/MontFerret/lab/commit/ebaac08bec0647bcc9b07eb9a3cc2cb9d29082f8)

#### Changed
- Removed --filter param. [2acb852](https://github.com/MontFerret/lab/commit/2acb852f3e431da96c85c3acb9c1162e0889f1fa)
- Removed Dockerfile tag with v prefix. [706a727](https://github.com/MontFerret/lab/commit/706a7276bd75da1de96fdbf1a610452311141ab6)
- Renamed context to data. [c77cab1](https://github.com/MontFerret/lab/commit/c77cab105ba48eba21ea5da3a189c948d5be3eed)
- Updated env variable name for wait-attempts. [363cc75](https://github.com/MontFerret/lab/commit/363cc75ad043dc745c5eba2e2621ea1c36d3eb33)

### 0.4.0
#### Added
- Support of expected failure (my-test.fail.fql). [1aafc18](https://github.com/MontFerret/lab/commit/1aafc18f5a0789ba272841c47da9ca8e487b4f6c)

### Fixed
- Execution of CDN. [bccb2af](https://github.com/MontFerret/lab/commit/bccb2af2548ef8e4eba2ced860aa67e523e4449d)
- Filtering. [d84ab66](https://github.com/MontFerret/lab/commit/d84ab66e021ccf5e13b86cb6166368b371461795)
- Graceful termination. [1aafc18](https://github.com/MontFerret/lab/commit/1aafc18f5a0789ba272841c47da9ca8e487b4f6c)

### 0.3.0
#### Added
- Installation script. [699e1fb](https://github.com/MontFerret/lab/commit/699e1fb307dba1757e30803917376921cec6bf0f)
- Extended options of resources that can be tested and waited on. [b321f72](https://github.com/MontFerret/lab/commit/b321f72b04ce00db5f07881cd7ca81f0c5c1911d)

#### Changed
- Updated Ferret to v0.10.1. [118a357](https://github.com/MontFerret/lab/commit/118a3576c611974fdb5d2ff82908994b1b3b943b)

#### Fixed
- Directory binding. [888555e](https://github.com/MontFerret/lab/commit/888555efda92903b2392f210a69ae68ad82eb39f)
- Binary lookup for an external runtime. [137999b](https://github.com/MontFerret/lab/commit/137999b8e90340a677d9be2ccaacdaca49db3e08)

#### Removed
- Builtin assertions. [118a357](https://github.com/MontFerret/lab/commit/118a3576c611974fdb5d2ff82908994b1b3b943b)

### 0.2.0
#### Added
- Remote HTTP runtime
- Remote binary runtime