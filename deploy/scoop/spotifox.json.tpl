{
    "version": "${SCOOP_VERSION}",
    "architecture": {
        "64bit": {
            "url": "https://github.com/go-musicfox/spotifox/releases/download/v${SCOOP_VERSION}/spotifox_${SCOOP_VERSION}_windows_amd64.zip",
            "hash": "${SCOOP_HASH}",
            "extract_dir": "spotifox_${SCOOP_VERSION}_windows_amd64"
        }
    },
    "bin": "spotifox.exe",
    "homepage": "https://github.com/go-musicfox/spotifox",
    "license": "MIT",
    "description": "go-musicfox是用Go写的又一款网易云音乐命令行客户端，支持UnblockNeteaseMusic、各种音质级别、lastfm、MPRIS...",
    "post_install": "Write-Host '好用记得给go-musicfox一个star✨哦~'",
    "env_set": {
        "SPOTIFOX_ROOT": "\$dir\\\\data"
    },
    "persist": "data",
    "checkver": "github",
    "autoupdate": {
        "architecture": {
            "64bit": {
                "url": "https://github.com/go-musicfox/spotifox/releases/download/v\$version/spotifox_\$version_windows_amd64.zip",
                "extract_dir": "spotifox_\$version_windows_amd64"
            }
        }
    }
}
