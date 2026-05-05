package cli

import "testing"

func TestResolveDestPath(t *testing.T) {
	cases := []struct {
		name     string
		srcPath  string
		destPath string
		want     string
	}{
		{
			name:     "trailing slash appends source filename",
			srcPath:  "/cmdb/archive.zip",
			destPath: "/home/azureuser/data_file/",
			want:     "/home/azureuser/data_file/archive.zip",
		},
		{
			name:     "root-level trailing slash",
			srcPath:  "/data/report.pdf",
			destPath: "/backup/",
			want:     "/backup/report.pdf",
		},
		{
			name:     "flat filename with trailing slash",
			srcPath:  "archive.zip",
			destPath: "/dest/",
			want:     "/dest/archive.zip",
		},
		{
			name:     "deeply nested source keeps only filename",
			srcPath:  "/a/b/c/d/file.tar.gz",
			destPath: "/out/",
			want:     "/out/file.tar.gz",
		},
		{
			name:     "explicit filename unchanged",
			srcPath:  "/src/archive.zip",
			destPath: "/dest/output.zip",
			want:     "/dest/output.zip",
		},
		{
			name:     "explicit filename different from source unchanged",
			srcPath:  "/src/file.txt",
			destPath: "/dst/renamed.txt",
			want:     "/dst/renamed.txt",
		},
		{
			name:     "no trailing slash no extension unchanged",
			srcPath:  "/src/file.txt",
			destPath: "/dst/nodot",
			want:     "/dst/nodot",
		},
		{
			name:     "empty dest unchanged",
			srcPath:  "/src/file.txt",
			destPath: "",
			want:     "",
		},
		{
			name:     "relative dest with trailing slash",
			srcPath:  "/data/image.png",
			destPath: "uploads/",
			want:     "uploads/image.png",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveDestPath(tc.srcPath, tc.destPath)
			if got != tc.want {
				t.Errorf("resolveDestPath(%q, %q) = %q, want %q",
					tc.srcPath, tc.destPath, got, tc.want)
			}
		})
	}
}
