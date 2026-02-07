package utils

import (
	"bytes"
	"slices"
)

// DetectMediaFormat 通过文件魔数（Magic Number）检测媒体格式
// 支持常见图片格式：JPEG, PNG, GIF, WebP, AVIF, HEIC, BMP, ICO, TIFF
func DetectMediaFormat(data []byte) string {
	if len(data) < 2 {
		return ".jpg" // 默认返回 jpg
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return ".png"
	}

	// JPEG: FF D8 FF
	if len(data) >= 3 && bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return ".jpg"
	}

	// GIF: GIF87a 或 GIF89a
	if len(data) >= 6 && (bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a"))) {
		return ".gif"
	}

	// WebP: RIFF....WEBP
	if len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.HasPrefix(data[8:], []byte("WEBP")) {
		return ".webp"
	}

	// AVIF: ....ftypavif
	if len(data) >= 12 && bytes.HasPrefix(data[4:], []byte("ftypavif")) {
		return ".avif"
	}

	// HEIC/HEIF: ....ftypheic 或 ....ftypheif
	if len(data) >= 12 && (bytes.HasPrefix(data[4:], []byte("ftypheic")) || bytes.HasPrefix(data[4:], []byte("ftypheif"))) {
		return ".heic"
	}

	// BMP: 42 4D (BM)
	if bytes.HasPrefix(data, []byte{0x42, 0x4D}) {
		return ".bmp"
	}

	// ICO: 00 00 01 00
	if len(data) >= 4 && bytes.HasPrefix(data, []byte{0x00, 0x00, 0x01, 0x00}) {
		return ".ico"
	}

	// TIFF: 49 49 2A 00 (little-endian) 或 4D 4D 00 2A (big-endian)
	if len(data) >= 4 {
		if bytes.HasPrefix(data, []byte{0x49, 0x49, 0x2A, 0x00}) || bytes.HasPrefix(data, []byte{0x4D, 0x4D, 0x00, 0x2A}) {
			return ".tiff"
		}
	}

	// MP4/MOV/M4V: ....ftyp (ftyp 前面有4字节表示 box size)
	// 常见的 ftyp 品牌包括: isom, mp41, mp42, avc1, M4V , qt  , 3gp, etc.
	if len(data) >= 12 && bytes.Equal(data[4:8], []byte("ftyp")) {
		// 检查是否为 MP4 相关格式
		ftypBrand := data[8:12]
		if bytes.HasPrefix(ftypBrand, []byte("isom")) ||
			bytes.HasPrefix(ftypBrand, []byte("mp4")) ||
			bytes.HasPrefix(ftypBrand, []byte("avc1")) ||
			bytes.HasPrefix(ftypBrand, []byte("M4V")) {
			return ".mp4"
		}
		// QuickTime MOV
		if bytes.HasPrefix(ftypBrand, []byte("qt  ")) {
			return ".mov"
		}
		// 3GP 移动视频格式
		if bytes.HasPrefix(ftypBrand, []byte("3gp")) {
			return ".3gp"
		}
		// 其他 ftyp 格式默认返回 mp4
		return ".mp4"
	}

	// AVI: RIFF....AVI (RIFF chunk 包含 AVI LIST)
	if len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:11], []byte("AVI")) {
		return ".avi"
	}

	// WebM: 1A 45 DF A3 (EBML 格式，Matroska 容器)
	if len(data) >= 4 && bytes.HasPrefix(data, []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		return ".webm"
	}

	// FLV: 46 4C 56 01 (FLV\x01)
	if len(data) >= 4 && bytes.HasPrefix(data, []byte{0x46, 0x4C, 0x56, 0x01}) {
		return ".flv"
	}

	// WMV/ASF: 30 26 B2 75 8E 66 CF 11
	if len(data) >= 8 && bytes.HasPrefix(data, []byte{0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11}) {
		return ".wmv"
	}

	// MKV: 1A 45 DF A3 后跟 Matroska 标识
	// 注意：WebM 也是基于 Matroska，但通过后续内容区分
	if len(data) >= 40 && bytes.HasPrefix(data, []byte{0x1A, 0x45, 0xDF, 0xA3}) {
		// 简单检查，如果包含 "matroska" 字符串则判定为 mkv
		if bytes.Contains(data[:40], []byte("matroska")) {
			return ".mkv"
		}
	}

	return ".jpg" // 默认返回 jpg
}

func IsVideo(data []byte) bool {
	format := DetectMediaFormat(data)
	videoFormats := []string{".mp4", ".mov", ".avi", ".webm", ".flv", ".wmv", ".mkv", ".3gp"}
	return slices.Contains(videoFormats, format)
}
