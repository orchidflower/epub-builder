package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/bmaupin/go-epub"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	decoder    *encoding.Decoder
	cssContent = `
.title {text-align:center}
.content {text-indent: 2em}
`
	htmlPStart     = `<p class="content">`
	htmlPEnd       = "</p>"
	htmlTitleStart = `<h3 class="title">`
	htmlTitleEnd   = "</h3>"
	defaultRegStr  = "^.{0,8}(第.{1,20}(章|节)|(S|s)ection.{1,20}|(C|c)hapter.{1,20}|(P|p)age.{1,20})|^\\d{1,4}.{0,20}$|^引子|^楔子|^章节目录"
	defaultMax     = 35
)

type EPubBuilder struct {
	FileName    string // 文件名
	BookName    string // 书名
	Cover       string // 封面图片
	Author      string // 作者
	TitleRegExp string // 章节匹配正则表达式
	TitleMax    uint   // 章节标题最大字数
	Lang        string // 语言
}

func (b *EPubBuilder) Before(c *cli.Context) error {
	return nil
}

func readAndDecode(fileName string) *bufio.Reader {
	f, err := os.Open(fileName)
	if err != nil {
		log.Printf("Failed to read file: %s", fileName)
		os.Exit(1)
	}
	temBuf := bufio.NewReader(f)
	bs, _ := temBuf.Peek(1024)
	encodig, encodename, _ := charset.DetermineEncoding(bs, "text/plain")
	if encodename != "utf-8" {
		f.Seek(0, 0)
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("读取文件出错: ", err.Error())
			os.Exit(1)
		}
		var buf bytes.Buffer
		decoder = encodig.NewDecoder()
		if encodename == "windows-1252" {
			decoder = simplifiedchinese.GB18030.NewDecoder()
		}
		bs, _, _ = transform.Bytes(decoder, bs)
		buf.Write(bs)
		return bufio.NewReader(&buf)
	} else {
		f.Seek(0, 0)
		buf := bufio.NewReader(f)
		return buf
	}
}

func AddPart(buff *bytes.Buffer, content string) {
	if strings.HasSuffix(content, "==") ||
		strings.HasSuffix(content, "**") ||
		strings.HasSuffix(content, "--") ||
		strings.HasSuffix(content, "//") {
		buff.WriteString(content)
		return
	}
	buff.WriteString(htmlPStart)
	buff.WriteString(content)
	buff.WriteString(htmlPEnd)
}

func (b *EPubBuilder) Build(c *cli.Context) error {
	reg, err := regexp.Compile(defaultRegStr)
	if err != nil {
		fmt.Printf("Failed to build regexp: %s\n%s\n", defaultRegStr, err.Error())
		return err
	}

	tempDir, err := ioutil.TempDir("", "epub-builder")
	if err != nil {
		log.Printf("Failed to create temp directory: %s. %e", tempDir, err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			panic(fmt.Sprintf("Failed to remove temp directory: %s", err))
		}
	}()
	pageStylesFile := path.Join(tempDir, "page_styles.css")
	err = ioutil.WriteFile(pageStylesFile, []byte(cssContent), 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed to write css style file: %s", err))
	}

	e := epub.NewEpub(b.BookName)
	e.SetLang("zh")
	// Set the author
	e.SetAuthor("Orchid")
	e.SetTitle(b.BookName)
	e.SetDescription("Hello")
	e.SetIdentifier("1234567890-ABCDEF")
	css, err := e.AddCSS(pageStylesFile, "")
	if err != nil {
		panic(fmt.Sprintf("Failed to add css style to epub: %s", err))
	}
	// 添加封面图片
	cover, err := e.AddImage(b.Cover, "")
	if err != nil {
		log.Fatalf("Failed to add cover image. %s", err)
	}
	e.SetCover(cover, "")

	fmt.Println("正在读取txt文件...")
	buf := readAndDecode(b.FileName)
	var title string
	var content bytes.Buffer

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					if line = strings.TrimSpace(line); line != "" {
						AddPart(&content, line)
					}
				}
				e.AddSection(content.String(), title, "", css)
				content.Reset()
				break
			}
			fmt.Println("读取文件出错:", err.Error())
			os.Exit(1)
		}
		line = strings.TrimSpace(line)
		// 空行直接跳过
		if len(line) == 0 {
			continue
		}
		// 处理标题
		if utf8.RuneCountInString(line) <= defaultMax && reg.MatchString(line) {
			if content.Len() == 0 {
				continue
			}
			e.AddSection(content.String(), title, "", css)
			title = line
			content.Reset()
			content.WriteString(htmlTitleStart)
			content.WriteString(title)
			content.WriteString(htmlTitleEnd)
			continue
		}
		AddPart(&content, line)
	}
	// 没识别到章节又没识别到 EOF 时，把所有的内容写到最后一章
	if content.Len() != 0 {
		if title == "" {
			title = "章节正文"
		}
		e.AddSection(content.String(), title, "", "")
	}

	// Write the EPUB
	fmt.Println("正在生成电子书...")
	epubName := b.BookName + ".epub"
	err = e.Write(epubName)
	if err != nil {
		// handle error
	}

	return nil
}

func (b *EPubBuilder) Split(c *cli.Context) error {
	return nil
}
