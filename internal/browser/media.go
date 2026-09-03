package browser

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ScreenshotOptions configures screenshot capture.
type ScreenshotOptions struct {
	FullPage bool
	Format   string // "png", "jpeg", "webp"
	Quality  int    // 0-100 (for jpeg/webp)
}

// PDFOptions configures PDF printing options.
type PDFOptions struct {
	Landscape         bool
	PrintBackground   bool
	PaperWidth        float64
	PaperHeight       float64
	MarginTop         float64
	MarginBottom      float64
	MarginLeft        float64
	MarginRight       float64
	PreferCSSPageSize bool
}

// CaptureScreenshot captures a screenshot of the current page state.
func CaptureScreenshot(ctx context.Context, page *rod.Page, opts ScreenshotOptions) ([]byte, error) {
	if page == nil {
		return nil, fmt.Errorf("browser: page is nil")
	}

	page = page.Context(ctx)

	req := &proto.PageCaptureScreenshot{}
	switch strings.ToLower(opts.Format) {
	case "jpeg", "jpg":
		req.Format = proto.PageCaptureScreenshotFormatJpeg
		if opts.Quality > 0 && opts.Quality <= 100 {
			q := opts.Quality
			req.Quality = &q
		}
	case "webp":
		req.Format = proto.PageCaptureScreenshotFormatWebp
		if opts.Quality > 0 && opts.Quality <= 100 {
			q := opts.Quality
			req.Quality = &q
		}
	default:
		req.Format = proto.PageCaptureScreenshotFormatPng
	}

	data, err := page.Screenshot(opts.FullPage, req)
	if err != nil {
		return nil, fmt.Errorf("browser: capture screenshot: %w", err)
	}

	return data, nil
}

// CapturePDF generates a PDF document from the current page state.
func CapturePDF(ctx context.Context, page *rod.Page, opts PDFOptions) ([]byte, error) {
	if page == nil {
		return nil, fmt.Errorf("browser: page is nil")
	}

	page = page.Context(ctx)

	req := &proto.PagePrintToPDF{
		Landscape:         opts.Landscape,
		PrintBackground:   opts.PrintBackground,
		PreferCSSPageSize: opts.PreferCSSPageSize,
	}

	if opts.PaperWidth > 0 {
		req.PaperWidth = &opts.PaperWidth
	}
	if opts.PaperHeight > 0 {
		req.PaperHeight = &opts.PaperHeight
	}
	if opts.MarginTop > 0 {
		req.MarginTop = &opts.MarginTop
	}
	if opts.MarginBottom > 0 {
		req.MarginBottom = &opts.MarginBottom
	}
	if opts.MarginLeft > 0 {
		req.MarginLeft = &opts.MarginLeft
	}
	if opts.MarginRight > 0 {
		req.MarginRight = &opts.MarginRight
	}

	reader, err := page.PDF(req)
	if err != nil {
		return nil, fmt.Errorf("browser: print to PDF: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("browser: read PDF stream: %w", err)
	}

	return data, nil
}
