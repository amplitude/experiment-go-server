module.exports = {
    "branches": [
        {name: 'beta', prerelease: true},
        "main"
    ],
    "tagFormat": ["v${version}"],
    "plugins": [
        ["@semantic-release/commit-analyzer", {
            "preset": "angular",
            "parserOpts": {
                "noteKeywords": ["BREAKING CHANGE", "BREAKING CHANGES", "BREAKING"]
            }
        }],
        ["@semantic-release/release-notes-generator", {
            "preset": "angular",
        }],
        ["@semantic-release/changelog", {
            "changelogFile": "CHANGELOG.md"
        }],
        "@semantic-release/github",
        [
        "@google/semantic-release-replace-plugin",
        {
            "replacements": [
            {
                "files": ["pkg/experiment/types.go"],
                "from": "VERSION = \".*\"",
                "to": "VERSION = \"${nextRelease.version}\"",
                "results": [
                {
                    "file": "pkg/experiment/types.go",
                    "hasChanged": true,
                    "numMatches": 1,
                    "numReplacements": 1
                }
                ],
                "countMatches": true
            },
            ]
        }
        ],
        ["@semantic-release/git", {
            "assets": ["pkg/experiment/types.go", "CHANGELOG.md"],
            "message": "chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}"
        }],
        "@semantic-release/github",
    ],
}
