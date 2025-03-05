<h1 align="center">
S3Scanner
</h1>

<p align="center">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg"/></a>
<a href="https://github.com/sponsors/sa7mon/"><img src="https://img.shields.io/github/sponsors/sa7mon" /></a>
<a href="https://github.com/sa7mon/S3Scanner/issues"><img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat"/></a>
<a href="https://github.com/sa7mon/S3Scanner/releases/latest"><img src="https://img.shields.io/github/v/release/sa7mon/s3scanner" /></a>
</p>
<p align="center">
<a href="#features">Features</a> - <a href="#usage">Usage</a> - <a href="#quick-start">Quick Start</a> - <a href="#installation">Installation</a> - <a href="https://github.com/sa7mon/S3Scanner/discussions">Discuss</a> 
</p>
<br>
A tool to find open S3 buckets in AWS or other cloud providers:

- AWS
- DigitalOcean
- DreamHost
- GCP
- Linode
- Scaleway
- Custom

<img alt="demo" src="https://github.com/sa7mon/S3Scanner/assets/3712226/cfa16801-2a44-4ae9-ad85-9dd466390cd9">

# Features

* ⚡️ Multi-threaded scanning
* 🔭 Supports many built-in S3 storage providers or custom
* 🕵️‍♀️ Scans all bucket permissions to find misconfigurations
* 💾 Save results to Postgres database
* 🐇 Connect to RabbitMQ for automated scanning at scale
* 🐳 Docker support

# Used By

<p align="center">
  <a href="https://github.com/six2dez/reconftw"><img src="https://github.com/six2dez/reconftw/blob/main/images/banner.png" alt="banner for six2dez/reconftw" width="50%"></a>
  <a href="https://github.com/yogeshojha/rengine"><img src="https://github.com/yogeshojha/rengine/blob/master/.github/screenshots/banner.gif" alt="banner for yogeshojha/rengine" width="50%"/></a>
  <a href="https://github.com/pry0cc/axiom"><img src="https://raw.githubusercontent.com/pry0cc/axiom/master/screenshots/axiom_banner.png" alt="banner for pry0cc/axiom - reads 'the dynamic infrastructure framework for everybody'" width="50%" /></a>
</p>

# Usage

```
INPUT: (1 required)
  -bucket        string  Name of bucket to check.
  -bucket-file   string  File of bucket names to check.
  -mq                    Connect to RabbitMQ to get buckets. Requires config file key "mq". Default: "false"

OUTPUT:
  -db       Save results to a Postgres database. Requires config file key "db.uri". Default: "false"
  -json     Print logs to stdout in JSON format instead of human-readable. Default: "false"

OPTIONS:
  -enumerate           Enumerate bucket objects (can be time-consuming). Default: "false"
  -provider    string  Object storage provider: aws, custom, digitalocean, dreamhost, gcp, linode, scaleway - custom requires config file. Default: "aws"
  -threads     int     Number of threads to scan with. Default: "4"

DEBUG:
  -verbose     Enable verbose logging. Default: "false"
  -version     Print version Default: "false"

If config file is required these locations will be searched for config.yml: "." "/etc/s3scanner/" "$HOME/.s3scanner/"
```

# 🚀 Support
If you've found this tool useful, please consider donating to support its development. You can find sponsor options on the side of this repo page or in [FUNDING.yml](.github/FUNDING.yml)

# Quick Start

Scan AWS for bucket names listed in a file, enumerate all objects
  ```shell
  $ s3scanner -bucket-file names.txt -enumerate
   ```

Scan a bucket in GCP, enumerate all objects, and save results to database
  ```shell
  $ s3scanner -provider gcp -db -bucket my-bucket -enumerate
  ```

# Installation

| Platform                  | Version                                                                                                                                                       | Steps                                                                                      |
|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------|
| BlackArch                 | [![BlackArch package](https://repology.org/badge/version-for-repo/blackarch/s3scanner.svg?header=BlackArch)](https://repology.org/project/s3scanner/versions) | `pacman -S s3scanner`                                                                      |
| Docker                    | ![Docker release](https://img.shields.io/github/v/release/sa7mon/s3scanner?label=Docker)                                                                      | `docker run ghcr.io/sa7mon/s3scanner`                                                      |
| Go                        | ![Golang](https://img.shields.io/github/v/release/sa7mon/s3scanner?label=Go)                                                                                  | `go install -v github.com/sa7mon/s3scanner@latest`                                         |
| Kali Linux                | [![Kali package](https://repology.org/badge/version-for-repo/kali_rolling/s3scanner.svg?header=Kali+Linux)](https://repology.org/project/s3scanner/versions)  | `apt install s3scanner`                                                                    |
| MacOS                     | [![homebrew version](https://img.shields.io/homebrew/v/s3scanner)](https://github.com/Homebrew/homebrew-core/blob/master/Formula/s/s3scanner.rb)              | `brew install s3scanner`                                                                   |
| Parrot OS                 | [![Parrot package](https://repology.org/badge/version-for-repo/parrot/s3scanner.svg?header=Parrot+OS)](https://repology.org/project/s3scanner/versions)       | `apt install s3scanner`                                                                    |
| Windows - winget          |                                                                                                                                                               | `winget install s3scanner`                                                                 |
| NixOS stable              | [![nixpkgs unstable package](https://repology.org/badge/version-for-repo/nix_stable_24_05/s3scanner.svg)](https://repology.org/project/s3scanner/versions)    | `nix-shell -p s3scanner`                                                                   |
| NixOS unstable            | [![nixpkgs unstable package](https://repology.org/badge/version-for-repo/nix_unstable/s3scanner.svg)](https://repology.org/project/s3scanner/versions)        | `nix-shell -p s3scanner`                                                                   |
| Other - Build from source | ![GitHub release](https://img.shields.io/github/v/release/sa7mon/s3scanner?label=Git)                                                                         | `git clone git@github.com:sa7mon/S3Scanner.git && cd S3Scanner && go build -o s3scanner .` |

# Config File

If using flags that require config options, `-config` must be used. Refer to the [config documentation]() for details.

# S3 compatible APIs

**Note:** `S3Scanner` currently only supports scanning for anonymous user permissions of non-AWS services

📚 More information on non-AWS APIs can be found [in the project wiki](https://github.com/sa7mon/S3Scanner/wiki/S3-Compatible-APIs).

# License

MIT
