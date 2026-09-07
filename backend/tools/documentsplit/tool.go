package documentsplit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"developer-toolbox/backend/models"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

var supportedExtensions = map[string]bool{
	".pdf": true,
	".doc": true, ".docx": true,
	".ppt": true, ".pptx": true,
}

// Result describes the printable odd/even PDF files created by the tool.
type Result struct {
	InputFile  string `json:"inputFile"`
	TotalPages int    `json:"totalPages"`
	OddPages   int    `json:"oddPages"`
	EvenPages  int    `json:"evenPages"`
	OddFile    string `json:"oddFile"`
	EvenFile   string `json:"evenFile,omitempty"`
	Converted  bool   `json:"converted"`
	Converter  string `json:"converter,omitempty"`
}

type Tool struct{}

func NewTool() *Tool { return &Tool{} }

func (t *Tool) Info() models.Tool {
	return models.Tool{
		ID:          "document-split",
		Name:        "Document Page Splitter",
		Description: "Split PDF, Word, and PowerPoint files into odd-page and even-page PDFs for duplex printing.",
		Category:    "document",
		Icon:        "files",
		Version:     "1.0.0",
		Capabilities: models.ToolCapabilities{
			FileInput: true, FileOutput: true,
		},
		Keywords: []string{"pdf", "word", "docx", "powerpoint", "pptx", "odd", "even", "duplex", "print"},
	}
}

func (t *Tool) Execute(toolCtx models.ToolContext, input models.ToolInput) models.ToolOutput {
	filePath, _ := input.Payload["filePath"].(string)
	outputDir, _ := input.Payload["outputDir"].(string)
	result, err := Split(toolCtx.Context, filePath, outputDir)
	if err != nil {
		return models.FailureWithDetail("DOCUMENT_SPLIT_FAILED", "无法拆分文件", err.Error())
	}
	return models.ToolOutput{Success: true, Data: result}
}

// Split renders supported Office files to PDF and writes odd/even page PDFs.
// The source file is never modified.
func Split(ctx context.Context, filePath, outputDir string) (Result, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return Result{}, errors.New("请选择 PDF、Word 或 PowerPoint 文件")
	}
	absInput, err := filepath.Abs(filePath)
	if err != nil {
		return Result{}, fmt.Errorf("解析源文件路径: %w", err)
	}
	info, err := os.Stat(absInput)
	if err != nil {
		return Result{}, fmt.Errorf("读取源文件: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Result{}, errors.New("源路径不是普通文件")
	}

	ext := strings.ToLower(filepath.Ext(absInput))
	if !supportedExtensions[ext] {
		return Result{}, fmt.Errorf("不支持 %s 文件，仅支持 PDF、DOC、DOCX、PPT 和 PPTX", ext)
	}
	if strings.TrimSpace(outputDir) == "" {
		outputDir = filepath.Dir(absInput)
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return Result{}, fmt.Errorf("解析输出目录: %w", err)
	}
	if outInfo, statErr := os.Stat(absOutput); statErr != nil {
		return Result{}, fmt.Errorf("读取输出目录: %w", statErr)
	} else if !outInfo.IsDir() {
		return Result{}, errors.New("输出路径不是目录")
	}

	pdfPath := absInput
	converted := ext != ".pdf"
	converter := ""
	if converted {
		tempDir, tempErr := os.MkdirTemp("", "developer-toolbox-document-split-*")
		if tempErr != nil {
			return Result{}, fmt.Errorf("创建临时目录: %w", tempErr)
		}
		defer os.RemoveAll(tempDir)
		pdfPath = filepath.Join(tempDir, "rendered.pdf")
		converter, err = convertToPDF(ctx, absInput, pdfPath)
		if err != nil {
			return Result{}, err
		}
	}

	pageCount, err := api.PageCountFile(pdfPath)
	if err != nil {
		return Result{}, fmt.Errorf("读取 PDF 页数: %w", err)
	}
	if pageCount < 1 {
		return Result{}, errors.New("文件没有可拆分的页面")
	}

	base := strings.TrimSuffix(filepath.Base(absInput), filepath.Ext(absInput))
	oddFile := nextAvailablePath(absOutput, base+"_odd", ".pdf")
	evenFile := ""
	if err := api.TrimFile(pdfPath, oddFile, []string{"odd"}, nil); err != nil {
		return Result{}, fmt.Errorf("生成奇数页 PDF: %w", err)
	}
	if pageCount > 1 {
		evenFile = nextAvailablePath(absOutput, base+"_even", ".pdf")
		if err := api.TrimFile(pdfPath, evenFile, []string{"even"}, nil); err != nil {
			_ = os.Remove(oddFile)
			return Result{}, fmt.Errorf("生成偶数页 PDF: %w", err)
		}
	}

	return Result{
		InputFile: absInput, TotalPages: pageCount,
		OddPages: (pageCount + 1) / 2, EvenPages: pageCount / 2,
		OddFile: oddFile, EvenFile: evenFile,
		Converted: converted, Converter: converter,
	}, nil
}

func nextAvailablePath(dir, base, ext string) string {
	candidate := filepath.Join(dir, base+ext)
	if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
		return candidate
	}
	for i := 2; ; i++ {
		candidate = filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

func convertToPDF(ctx context.Context, input, output string) (string, error) {
	var conversionErrors []error
	if runtime.GOOS == "windows" {
		if err := convertWithMicrosoftOffice(ctx, input, output); err == nil {
			if err := validateConvertedPDF(output); err == nil {
				return "Microsoft Office", nil
			} else {
				conversionErrors = append(conversionErrors, err)
			}
		} else {
			conversionErrors = append(conversionErrors, err)
		}
	}
	if err := convertWithLibreOffice(ctx, input, output); err == nil {
		if err := validateConvertedPDF(output); err == nil {
			return "LibreOffice", nil
		} else {
			conversionErrors = append(conversionErrors, err)
		}
	} else {
		conversionErrors = append(conversionErrors, err)
	}
	return "", fmt.Errorf("无法将 Office 文件转换为 PDF（请安装 Microsoft Office 或 LibreOffice）: %w", errors.Join(conversionErrors...))
}

func validateConvertedPDF(output string) error {
	info, err := os.Stat(output)
	if err != nil {
		return fmt.Errorf("转换程序未生成 PDF: %w", err)
	}
	if info.Size() == 0 {
		return errors.New("转换程序生成了空 PDF")
	}
	return nil
}

const officeConversionScript = `$ErrorActionPreference = 'Stop'
$inputPath = $env:DEVTOOLBOX_SPLIT_INPUT
$outputPath = $env:DEVTOOLBOX_SPLIT_OUTPUT
$extension = [IO.Path]::GetExtension($inputPath).ToLowerInvariant()
if ($extension -eq '.doc' -or $extension -eq '.docx') {
  $application = New-Object -ComObject Word.Application
  try {
    $application.Visible = $false
    $application.DisplayAlerts = 0
	$application.AutomationSecurity = 3
    $document = $application.Documents.Open($inputPath, $false, $true)
    try { $document.ExportAsFixedFormat($outputPath, 17) } finally { $document.Close($false) }
  } finally { $application.Quit() }
} elseif ($extension -eq '.ppt' -or $extension -eq '.pptx') {
  $application = New-Object -ComObject PowerPoint.Application
  try {
	$application.AutomationSecurity = 3
    $presentation = $application.Presentations.Open($inputPath, $true, $false, $false)
    try { $presentation.SaveAs($outputPath, 32) } finally { $presentation.Close() }
  } finally { $application.Quit() }
} else { throw "Unsupported Office extension: $extension" }
if (-not (Test-Path -LiteralPath $outputPath)) { throw 'Office did not create the PDF output.' }`

func convertWithMicrosoftOffice(ctx context.Context, input, output string) error {
	powerShell, err := exec.LookPath("powershell.exe")
	if err != nil {
		return errors.New("未找到 Windows PowerShell")
	}
	cmd := exec.CommandContext(ctx, powerShell, "-NoLogo", "-NoProfile", "-NonInteractive", "-STA", "-Command", officeConversionScript)
	cmd.Env = append(os.Environ(), "DEVTOOLBOX_SPLIT_INPUT="+input, "DEVTOOLBOX_SPLIT_OUTPUT="+output)
	if raw, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("Microsoft Office 转换失败: %w (%s)", runErr, strings.TrimSpace(string(raw)))
	}
	return nil
}

func convertWithLibreOffice(ctx context.Context, input, output string) error {
	executable, err := exec.LookPath("soffice")
	if err != nil {
		executable, err = exec.LookPath("libreoffice")
	}
	if err != nil {
		return errors.New("未找到 LibreOffice")
	}
	outDir := filepath.Dir(output)
	cmd := exec.CommandContext(ctx, executable, "--headless", "--convert-to", "pdf", "--outdir", outDir, input)
	if raw, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("LibreOffice 转换失败: %w (%s)", runErr, strings.TrimSpace(string(raw)))
	}
	generated := filepath.Join(outDir, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))+".pdf")
	if generated != output {
		if err := os.Rename(generated, output); err != nil {
			return fmt.Errorf("整理 LibreOffice 输出: %w", err)
		}
	}
	return nil
}
