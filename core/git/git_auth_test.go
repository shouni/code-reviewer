package git

import "testing"

func TestIsSSHRepoURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "scp style", url: "bbmf@bbmf.git.backlog.jp:/MK/APP.git", want: true},
		{name: "ssh url", url: "ssh://bbmf@bbmf.git.backlog.jp/MK/APP.git", want: true},
		{name: "https url", url: "https://example.com/repo.git", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSSHRepoURL(tt.url); got != tt.want {
				t.Fatalf("isSSHRepoURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestSSHUsernameFromRepoURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{name: "scp style custom user", url: "bbmf@bbmf.git.backlog.jp:/MK/APP.git", want: "bbmf"},
		{name: "scp style git user", url: "git@github.com:owner/repo.git", want: "git"},
		{name: "ssh scheme user", url: "ssh://bbmf@bbmf.git.backlog.jp/MK/APP.git", want: "bbmf"},
		{name: "ssh scheme without user", url: "ssh://bbmf.git.backlog.jp/MK/APP.git", want: "git"},
		{name: "invalid ssh form", url: "bbmf.git.backlog.jp:/MK/APP.git", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sshUsernameFromRepoURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("sshUsernameFromRepoURL(%q) expected error", tt.url)
				}
				return
			}
			if err != nil {
				t.Fatalf("sshUsernameFromRepoURL(%q) unexpected error: %v", tt.url, err)
			}
			if got != tt.want {
				t.Fatalf("sshUsernameFromRepoURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}
