package core

import (
	"testing"
)

func TestExtractParamsQuark(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  string
		wantPwd string
		wantErr bool
	}{
		{
			name:    "Normal URL",
			url:     "https://pan.quark.cn/s/0592e1dbe475",
			wantID:  "0592e1dbe475",
			wantPwd: "",
			wantErr: false,
		},
		{
			name:    "URL with pwd",
			url:     "https://pan.quark.cn/s/45c6cd59a7f9?pwd=D3eM",
			wantID:  "45c6cd59a7f9",
			wantPwd: "D3eM",
			wantErr: false,
		},
		{
			name:    "Invalid Domain",
			url:     "https://pan.baidu.com/s/123456",
			wantID:  "",
			wantPwd: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotPwd, err := extractParamsQuark(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParamsQuark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("extractParamsQuark() gotID = %v, want %v", gotID, tt.wantID)
			}
			if gotPwd != tt.wantPwd {
				t.Errorf("extractParamsQuark() gotPwd = %v, want %v", gotPwd, tt.wantPwd)
			}
		})
	}
}

func TestExtractParamsAliPan(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  string
		wantErr bool
	}{
		{
			name:    "Normal URL",
			url:     "https://www.alipan.com/s/Xd4HxfpMdVk",
			wantID:  "Xd4HxfpMdVk",
			wantErr: false,
		},
		{
			name:    "URL with params",
			url:     "https://www.alipan.com/s/Xd4HxfpMdVk?folder_id=123",
			wantID:  "Xd4HxfpMdVk",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := extractParamsAliPan(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParamsAliPan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("extractParamsAliPan() gotID = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestExtractParamsYes(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  string
		wantPwd string
		wantErr bool
	}{
		{
			name:    "Normal URL 123684",
			url:     "https://www.123684.com/s/A6xcVv-1jIxh",
			wantID:  "A6xcVv-1jIxh",
			wantPwd: "",
			wantErr: false,
		},
		{
			name:    "Normal URL 123865",
			url:     "https://www.123865.com/s/A6xcVv-1jIxh",
			wantID:  "A6xcVv-1jIxh",
			wantPwd: "",
			wantErr: false,
		},
		{
			name:    "URL with pwd",
			url:     "https://www.123684.com/s/A6xcVv-1jIxh?pwd=abcd",
			wantID:  "A6xcVv-1jIxh",
			wantPwd: "abcd",
			wantErr: false,
		},
		{
			name:    "Invalid Domain",
			url:     "https://www.google.com/s/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotPwd, err := extractParamsYes(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParamsYes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("extractParamsYes() gotID = %v, want %v", gotID, tt.wantID)
			}
			if gotPwd != tt.wantPwd {
				t.Errorf("extractParamsYes() gotPwd = %v, want %v", gotPwd, tt.wantPwd)
			}
		})
	}
}

func TestExtractParamsUc(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  string
		wantErr bool
	}{
		{
			name:    "Normal URL",
			url:     "https://drive.uc.cn/s/9b7941c42f0a4",
			wantID:  "9b7941c42f0a4",
			wantErr: false,
		},
		{
			name:    "URL with query",
			url:     "https://drive.uc.cn/s/9b7941c42f0a4?public=1",
			wantID:  "9b7941c42f0a4",
			wantErr: false,
		},
		{
			name:    "Invalid format",
			url:     "https://drive.uc.cn/t/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := extractParamsUc(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParamsUc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("extractParamsUc() gotID = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestExtractParamsYyw(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantShare   string
		wantReceive string
		wantErr     bool
	}{
		{
			name:        "Normal URL",
			url:         "https://115cdn.com/s/sww0nyf3zv8?password=n865",
			wantShare:   "sww0nyf3zv8",
			wantReceive: "n865",
			wantErr:     false,
		},
		{
			name:        "URL with fragment",
			url:         "https://115cdn.com/s/sww0nyf3zv8#?password=n865",
			wantShare:   "sww0nyf3zv8",
			wantReceive: "n865",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotShare, gotReceive, err := extractParamsYyw(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParamsYyw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShare != tt.wantShare {
				t.Errorf("extractParamsYyw() gotShare = %v, want %v", gotShare, tt.wantShare)
			}
			if gotReceive != tt.wantReceive {
				t.Errorf("extractParamsYyw() gotReceive = %v, want %v", gotReceive, tt.wantReceive)
			}
		})
	}
}
