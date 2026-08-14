package customimage

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twossh/minitela/internal/acf"
)

func writeTestTemplate(
	t *testing.T,
	path string,
) {
	t.Helper()

	data := make(
		[]byte,
		acf.TextureTemplateSize,
	)

	binary.LittleEndian.PutUint32(
		data[len(data)-4:],
		acf.FooterMagic,
	)

	if err := acf.SetChecksum(data); err != nil {
		t.Fatalf(
			"SetChecksum(): %v",
			err,
		)
	}

	if err := os.WriteFile(
		path,
		data,
		0600,
	); err != nil {
		t.Fatalf(
			"WriteFile(): %v",
			err,
		)
	}
}

func TestResolveTemplatePathInternal(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(
		dir,
		"Texture-template.acf",
	)

	writeTestTemplate(
		t,
		path,
	)

	t.Setenv(
		"MINITELA_TEMPLATE",
		path,
	)

	got, err :=
		ResolveTemplatePath()

	if err != nil {
		t.Fatalf(
			"ResolveTemplatePath(): %v",
			err,
		)
	}

	if got != path {
		t.Fatalf(
			"ResolveTemplatePath() = %q; esperado %q",
			got,
			path,
		)
	}
}

func TestResolveTemplatePathMissing(t *testing.T) {
	t.Setenv(
		"MINITELA_TEMPLATE",
		"",
	)

	_, err :=
		ResolveTemplatePath()

	if err == nil {
		t.Fatal(
			"ResolveTemplatePath() deveria falhar sem template interno",
		)
	}

	if !strings.Contains(
		err.Error(),
		"componente interno",
	) {
		t.Fatalf(
			"erro inesperado: %v",
			err,
		)
	}
}

func TestResolveTemplatePathInvalid(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(
		dir,
		"invalid.acf",
	)

	if err := os.WriteFile(
		path,
		[]byte("invalid"),
		0600,
	); err != nil {
		t.Fatalf(
			"WriteFile(): %v",
			err,
		)
	}

	t.Setenv(
		"MINITELA_TEMPLATE",
		path,
	)

	_, err :=
		ResolveTemplatePath()

	if err == nil {
		t.Fatal(
			"ResolveTemplatePath() deveria rejeitar template inválido",
		)
	}

	if !strings.Contains(
		err.Error(),
		"componente interno",
	) {
		t.Fatalf(
			"erro inesperado: %v",
			err,
		)
	}
}
