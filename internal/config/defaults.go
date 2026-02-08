package config

// DefaultConfigYAML returns the example configuration YAML string.
// This is used by `prjct install` to write the initial config file.
func DefaultConfigYAML() string {
	return `# prjct configuration file
# Docs: https://github.com/fwartner/prjct

templates:
  - id: video
    name: "Video Production"
    base_path: "~/Projects/Video"
    directories:
      - name: "01_Pre-Production"
        children:
          - name: "Scripts"
          - name: "Storyboards"
          - name: "References"
          - name: "Mood Boards"
      - name: "02_Production"
        children:
          - name: "Footage"
            children:
              - name: "A-Roll"
              - name: "B-Roll"
              - name: "Interviews"
          - name: "Audio"
            children:
              - name: "Voiceover"
              - name: "Music"
              - name: "SFX"
          - name: "Photos"
      - name: "03_Post-Production"
        children:
          - name: "Project Files"
          - name: "Edits"
            children:
              - name: "Rough Cut"
              - name: "Fine Cut"
              - name: "Final"
          - name: "Graphics"
          - name: "Color"
          - name: "Sound Design"
      - name: "04_Export"
        children:
          - name: "Masters"
          - name: "Proxies"
          - name: "Thumbnails"
      - name: "05_Delivery"
      - name: "06_Archive"

  - id: photo
    name: "Photography"
    base_path: "~/Projects/Photo"
    directories:
      - name: "RAW"
      - name: "Selects"
      - name: "Edits"
      - name: "Export"
        children:
          - name: "Web"
          - name: "Print"
          - name: "Social"
      - name: "BTS"

  - id: dev
    name: "Software Development"
    base_path: "~/Projects/Dev"
    directories:
      - name: "docs"
      - name: "src"
      - name: "tests"
      - name: "scripts"
`
}
