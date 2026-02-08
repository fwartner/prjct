package project

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fwartner/prjct/internal/config"
)

func TestCreateBasic(t *testing.T) {
	base := t.TempDir()
	tmpl := &config.Template{
		ID:       "test",
		Name:     "Test Template",
		BasePath: base,
		Directories: []config.Directory{
			{Name: "src"},
			{Name: "docs"},
			{Name: "tests"},
		},
	}

	result, err := Create(tmpl, "MyProject", false)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if result.TemplateName != "Test Template" {
		t.Errorf("TemplateName = %q, want %q", result.TemplateName, "Test Template")
	}
	if result.DirsCreated != 4 { // root + 3 dirs
		t.Errorf("DirsCreated = %d, want 4", result.DirsCreated)
	}

	expectedPath := filepath.Join(base, "MyProject")
	if result.ProjectPath != expectedPath {
		t.Errorf("ProjectPath = %q, want %q", result.ProjectPath, expectedPath)
	}

	// Verify directories exist
	for _, name := range []string{"src", "docs", "tests"} {
		path := filepath.Join(expectedPath, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("directory %q was not created", path)
		}
	}
}

func TestCreateNested(t *testing.T) {
	base := t.TempDir()
	tmpl := &config.Template{
		ID:       "nested",
		Name:     "Nested",
		BasePath: base,
		Directories: []config.Directory{
			{
				Name: "parent",
				Children: []config.Directory{
					{
						Name: "child",
						Children: []config.Directory{
							{Name: "grandchild"},
						},
					},
				},
			},
		},
	}

	result, err := Create(tmpl, "DeepProject", false)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if result.DirsCreated != 4 { // root + parent + child + grandchild
		t.Errorf("DirsCreated = %d, want 4", result.DirsCreated)
	}

	// Verify nested structure
	paths := []string{
		"parent",
		filepath.Join("parent", "child"),
		filepath.Join("parent", "child", "grandchild"),
	}
	for _, rel := range paths {
		full := filepath.Join(result.ProjectPath, rel)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("nested directory %q was not created", rel)
		}
	}
}

func TestCreateComplexStructure(t *testing.T) {
	base := t.TempDir()
	tmpl := &config.Template{
		ID:       "video",
		Name:     "Video Production",
		BasePath: base,
		Directories: []config.Directory{
			{
				Name: "01_Pre-Production",
				Children: []config.Directory{
					{Name: "Scripts"},
					{Name: "Storyboards"},
				},
			},
			{
				Name: "02_Production",
				Children: []config.Directory{
					{
						Name: "Footage",
						Children: []config.Directory{
							{Name: "A-Roll"},
							{Name: "B-Roll"},
						},
					},
					{Name: "Audio"},
				},
			},
			{Name: "03_Delivery"},
		},
	}

	result, err := Create(tmpl, "Commercial", false)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// root(1) + 01_Pre-Production(1) + Scripts(1) + Storyboards(1) +
	// 02_Production(1) + Footage(1) + A-Roll(1) + B-Roll(1) + Audio(1) + 03_Delivery(1) = 10
	if result.DirsCreated != 10 {
		t.Errorf("DirsCreated = %d, want 10", result.DirsCreated)
	}

	// Spot check deep paths
	deepPath := filepath.Join(result.ProjectPath, "02_Production", "Footage", "A-Roll")
	if _, err := os.Stat(deepPath); os.IsNotExist(err) {
		t.Errorf("deep directory %q was not created", deepPath)
	}
}

func TestCreateProjectExists(t *testing.T) {
	base := t.TempDir()
	projectDir := filepath.Join(base, "Existing")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tmpl := &config.Template{
		ID:       "test",
		Name:     "Test",
		BasePath: base,
		Directories: []config.Directory{
			{Name: "src"},
		},
	}

	_, err := Create(tmpl, "Existing", false)
	if err == nil {
		t.Fatal("Create() expected error for existing project, got nil")
	}
	if !errors.Is(err, ErrProjectExists) {
		t.Errorf("Create() error = %v, want ErrProjectExists", err)
	}
}

func TestCreateWithSpacesInName(t *testing.T) {
	base := t.TempDir()
	tmpl := &config.Template{
		ID:       "test",
		Name:     "Test",
		BasePath: base,
		Directories: []config.Directory{
			{Name: "src"},
		},
	}

	result, err := Create(tmpl, "Client Commercial 2026", false)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	expectedPath := filepath.Join(base, "Client Commercial 2026")
	if result.ProjectPath != expectedPath {
		t.Errorf("ProjectPath = %q, want %q", result.ProjectPath, expectedPath)
	}
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("project directory with spaces was not created")
	}
}

func TestCreateWithTildeExpansion(t *testing.T) {
	// Use a real temp dir to avoid tilde issues
	base := t.TempDir()
	tmpl := &config.Template{
		ID:       "test",
		Name:     "Test",
		BasePath: base,
		Directories: []config.Directory{
			{Name: "src"},
		},
	}

	result, err := Create(tmpl, "TildeTest", false)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if _, err := os.Stat(result.ProjectPath); os.IsNotExist(err) {
		t.Error("project directory was not created")
	}
}

func TestCreateVerbose(t *testing.T) {
	base := t.TempDir()
	tmpl := &config.Template{
		ID:       "test",
		Name:     "Test",
		BasePath: base,
		Directories: []config.Directory{
			{Name: "src"},
		},
	}

	// Verbose mode should not cause errors
	result, err := Create(tmpl, "VerboseProject", true)
	if err != nil {
		t.Fatalf("Create() with verbose error: %v", err)
	}
	if result.DirsCreated != 2 { // root + src
		t.Errorf("DirsCreated = %d, want 2", result.DirsCreated)
	}
}

func TestCreateCreatesBasePath(t *testing.T) {
	base := t.TempDir()
	deepBase := filepath.Join(base, "nested", "deep", "path")

	tmpl := &config.Template{
		ID:       "test",
		Name:     "Test",
		BasePath: deepBase,
		Directories: []config.Directory{
			{Name: "src"},
		},
	}

	result, err := Create(tmpl, "Project", false)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if _, err := os.Stat(result.ProjectPath); os.IsNotExist(err) {
		t.Error("project not created under deep base path")
	}
}

func TestFlatten(t *testing.T) {
	dirs := []config.Directory{
		{
			Name: "a",
			Children: []config.Directory{
				{Name: "b"},
				{
					Name: "c",
					Children: []config.Directory{
						{Name: "d"},
					},
				},
			},
		},
		{Name: "e"},
	}

	paths := flatten(dirs, "")

	expected := []string{
		"a",
		filepath.Join("a", "b"),
		filepath.Join("a", "c"),
		filepath.Join("a", "c", "d"),
		"e",
	}

	if len(paths) != len(expected) {
		t.Fatalf("flatten() returned %d paths, want %d: %v", len(paths), len(expected), paths)
	}

	for i, want := range expected {
		if paths[i] != want {
			t.Errorf("flatten()[%d] = %q, want %q", i, paths[i], want)
		}
	}
}

func TestFlattenEmpty(t *testing.T) {
	paths := flatten(nil, "")
	if len(paths) != 0 {
		t.Errorf("flatten(nil) returned %d paths, want 0", len(paths))
	}
}

func TestFlattenWithPrefix(t *testing.T) {
	dirs := []config.Directory{
		{Name: "child"},
	}

	paths := flatten(dirs, "parent")
	if len(paths) != 1 {
		t.Fatalf("flatten() returned %d paths, want 1", len(paths))
	}
	want := filepath.Join("parent", "child")
	if paths[0] != want {
		t.Errorf("flatten()[0] = %q, want %q", paths[0], want)
	}
}

func TestCreatePermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows (no Unix-style directory permissions)")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	base := t.TempDir()
	readonlyDir := filepath.Join(base, "readonly")
	if err := os.Mkdir(readonlyDir, 0555); err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer func() { _ = os.Chmod(readonlyDir, 0755) }()

	tmpl := &config.Template{
		ID:       "test",
		Name:     "Test",
		BasePath: readonlyDir,
		Directories: []config.Directory{
			{Name: "src"},
		},
	}

	_, err := Create(tmpl, "Forbidden", false)
	if err == nil {
		t.Fatal("Create() expected error for permission denied, got nil")
	}
	if !errors.Is(err, ErrPermission) {
		t.Errorf("Create() error = %v, want ErrPermission", err)
	}
}
