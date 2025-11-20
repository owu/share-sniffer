package core

import (
	"testing"
)

func TestExtractParamsTelecom(t *testing.T) {
	// 定义测试用例
	testCases := []struct {
		name          string
		url           string
		expectedCode  string
		expectedError bool
	}{
		{
			name:          "web share simple",
			url:           "https://cloud.189.cn/web/share?code=nEniUnvIBBBv",
			expectedCode:  "nEniUnvIBBBv",
			expectedError: false,
		},
		{
			name:          "web share with encoded suffix",
			url:           "https://cloud.189.cn/web/share?code=7BfYRjRZvYBz%EF%BC%88%E8%AE%BF%E9%97%AE%E7%A0%81%EF%BC%9Ac0jt%EF%BC%89",
			expectedCode:  "7BfYRjRZvYBz",
			expectedError: false,
		},
		{
			name:          "web share with chinese suffix",
			url:           "https://cloud.189.cn/web/share?code=7BfYRjRZvYBz（访问码：c0jt）",
			expectedCode:  "7BfYRjRZvYBz",
			expectedError: false,
		},
		{
			name:          "t prefix with encoded suffix",
			url:           "https://cloud.189.cn/t/6FjeIfQvMRba%EF%BC%88%E8%AE%BF%E9%97%AE%E7%A0%81%EF%BC%9A2jio%EF%BC%89",
			expectedCode:  "6FjeIfQvMRba",
			expectedError: false,
		},
		{
			name:          "t prefix with chinese suffix",
			url:           "https://cloud.189.cn/t/6FjeIfQvMRba（访问码：2jio）",
			expectedCode:  "6FjeIfQvMRba",
			expectedError: false,
		},
		{
			name:          "t prefix simple",
			url:           "https://cloud.189.cn/t/bm2iuqZZj632",
			expectedCode:  "bm2iuqZZj632",
			expectedError: false,
		},
	}

	// 运行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, _, err := extractParamsTelecom(tc.url)

			if tc.expectedError {
				if err == nil {
					t.Errorf("预期会产生错误，但实际没有错误")
				}
				return
			}

			if err != nil {
				t.Errorf("预期不会产生错误，但实际产生了错误: %v", err)
				return
			}

			if code != tc.expectedCode {
				t.Errorf("预期 code 为 %s，但实际为 %s", tc.expectedCode, code)
			}
		})
	}
}
