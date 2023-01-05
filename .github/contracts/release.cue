schemaVersion: "v1"
materials: [
	{type: "ARTIFACT", name: "goket-linux-amd64", output: true},
	{type: "ARTIFACT", name: "goket-linux-arm", output: true},
	{type: "ARTIFACT", name: "goket-linux-arm64", output: true},
	// Software Bill Of Materials for the generated binaries
	{type: "SBOM_CYCLONEDX_JSON", name: "sbom"},
]
runner: type: "GITHUB_ACTION"
