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
    "description": "Spotifox is yet another spotify CLI client.",
    "post_install": "Write-Host 'Star✨ please~'",
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
