package protocol

import (
	"bytes"
	"crypto/md5"
	"testing"
)

func TestBuildGetDownloadStatusRequest(t *testing.T) {
	want := []byte{
		0x41, 0x48,
		0x00, 0x02,
		0x00, 0x85,
		0x00, 0x00,
		0x4D, 0x49,
	}

	got :=
		BuildGetDownloadStatusRequest()

	if !bytes.Equal(got, want) {
		t.Fatalf(
			"got : % X\nwant: % X",
			got,
			want,
		)
	}
}

func TestBuildSwitchToDownloadModeRequest(
	t *testing.T,
) {
	want := []byte{
		0x41, 0x48,
		0x00, 0x03,
		0x00, 0x71,
		0x10,
		0x00, 0x00,
		0x4D, 0x49,
	}

	got :=
		BuildSwitchToDownloadModeRequest()

	if !bytes.Equal(got, want) {
		t.Fatalf(
			"got : % X\nwant: % X",
			got,
			want,
		)
	}
}

func TestBuildRequestDownloadRequest(
	t *testing.T,
) {
	fileID :=
		md5.Sum(
			[]byte("MiniTela"),
		)

	got :=
		BuildRequestDownloadRequest(
			0x08100000,
			0x00123456,
			fileID,
		)

	if got[0] != 0x41 ||
		got[1] != 0x48 {
		t.Fatal(
			"cabeçalho inválido",
		)
	}

	if got[4] != 0x00 ||
		got[5] != 0x81 {
		t.Fatal(
			"comando RequestDownload inválido",
		)
	}

	if got[len(got)-2] != 0x4D ||
		got[len(got)-1] != 0x49 {
		t.Fatal(
			"rodapé inválido",
		)
	}
}

func TestBuildDownloadDataRequestPadding(
	t *testing.T,
) {
	got :=
		BuildDownloadDataRequest(
			0x00000100,
			[]byte{
				0x11,
				0x22,
				0x33,
			},
		)

	// Offset 4 bytes + dados 3 bytes +
	// padding 1 byte.
	//
	// command length = 2 + 8 = 10.
	if got[2] != 0x00 ||
		got[3] != 0x0A {
		t.Fatalf(
			"commandLength=%02X%02X esperado=000A",
			got[2],
			got[3],
		)
	}

	if got[4] != 0x00 ||
		got[5] != 0x82 {
		t.Fatal(
			"comando DownloadData inválido",
		)
	}
}

func TestParseDownloadStatus(
	t *testing.T,
) {
	var fileID [16]byte

	for i := range fileID {
		fileID[i] = byte(i)
	}

	content := make(
		[]byte,
		21,
	)

	content[0] =
		DownloadStateActive

	copy(
		content[1:17],
		fileID[:],
	)

	content[17] = 0x00
	content[18] = 0x00
	content[19] = 0x10
	content[20] = 0x00

	got, err :=
		ParseDownloadStatus(
			content,
		)

	if err != nil {
		t.Fatal(err)
	}

	if got.Status !=
		DownloadStateActive {
		t.Fatalf(
			"Status=0x%02X",
			got.Status,
		)
	}

	if got.Offset != 0x1000 {
		t.Fatalf(
			"Offset=0x%X esperado=0x1000",
			got.Offset,
		)
	}

	if got.FileID != fileID {
		t.Fatal(
			"FileID incorreto",
		)
	}
}

func TestParseRequestDownloadResponse(
	t *testing.T,
) {
	content := []byte{
		0x00, 0x00, 0x04, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	got, err :=
		ParseRequestDownloadResponse(
			content,
		)

	if err != nil {
		t.Fatal(err)
	}

	if got.MaxPageSize != 1024 {
		t.Fatalf(
			"MaxPageSize=%d esperado=1024",
			got.MaxPageSize,
		)
	}

	if got.Response != 0 {
		t.Fatalf(
			"Response=0x%08X",
			got.Response,
		)
	}
}

func TestParseUploadResponse(
	t *testing.T,
) {
	frame := []byte{
		0x41, 0x48,

		0x00, 0x06,

		0x00, 0xC2,

		0x00, 0x00, 0x00, 0x00,

		0x00, 0x00,

		0x4D, 0x49,
	}

	content, err :=
		ParseUploadResponse(
			frame,
			CommandDownloadDataResponse,
		)

	if err != nil {
		t.Fatal(err)
	}

	code, err :=
		ParseUploadResultCode(
			content,
		)

	if err != nil {
		t.Fatal(err)
	}

	if code != 0 {
		t.Fatalf(
			"code=0x%08X esperado=0",
			code,
		)
	}
}
