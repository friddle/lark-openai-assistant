package feishu

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/88250/lute"
	"github.com/chyroc/lark"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func (c *FeishuClient) GetDocContent(docToken string, options ...lark.MethodOptionFunc) (*lark.DocContent, error) {
	resp, _, err := c.LarkClient.Drive.GetDriveDocContent(c.Ctx, &lark.GetDriveDocContentReq{
		DocToken: docToken,
	}, options...)
	if err != nil {
		return nil, err
	}
	doc := &lark.DocContent{}
	err = json.Unmarshal([]byte(resp.Content), doc)
	if err != nil {
		return doc, err
	}

	return doc, nil
}

func (c *FeishuClient) GetSheetDoc(docToken string, options ...lark.MethodOptionFunc) ([]*lark.GetSheetRespSheet, error) {
	metaRsp, _, _ := c.LarkClient.Drive.GetSheetMeta(c.Ctx, &lark.GetSheetMetaReq{
		SpreadSheetToken: docToken,
	})
	sheetContents := []*lark.GetSheetRespSheet{}
	for _, sheet := range metaRsp.Sheets {
		resp, _, err := c.LarkClient.Drive.GetSheet(c.Ctx, &lark.GetSheetReq{
			SpreadSheetToken: docToken,
			SheetID:          sheet.SheetID,
		}, options...)
		if err != nil {
			continue
		}
		sheetContents = append(sheetContents, resp.Sheet)
	}

	return sheetContents, nil

}

func (c *FeishuClient) DownloadImage(imgToken string, options ...lark.MethodOptionFunc) (string, error) {
	resp, _, err := c.LarkClient.Drive.DownloadDriveMedia(c.Ctx, &lark.DownloadDriveMediaReq{
		FileToken: imgToken,
	}, options...)
	if err != nil {
		return imgToken, err
	}
	imgDir := c.CacheDir + "/img"
	fileext := filepath.Ext(resp.Filename)
	filename := fmt.Sprintf("%s/%s%s", imgDir, imgToken, fileext)
	err = os.MkdirAll(filepath.Dir(filename), 0o755)
	if err != nil {
		return imgToken, err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return imgToken, err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.File)
	if err != nil {
		return imgToken, err
	}
	return filename, nil
}

func (c *FeishuClient) GetDocxContent(docToken string, options ...lark.MethodOptionFunc) (*lark.DocxDocument, []*lark.DocxBlock, error) {
	resp, _, err := c.LarkClient.Drive.GetDocxDocument(c.Ctx, &lark.GetDocxDocumentReq{
		DocumentID: docToken,
	}, options...)
	if err != nil {
		return nil, nil, err
	}
	docx := &lark.DocxDocument{
		DocumentID: resp.Document.DocumentID,
		RevisionID: resp.Document.RevisionID,
		Title:      resp.Document.Title,
	}
	var blocks []*lark.DocxBlock
	var pageToken *string
	for {
		resp2, _, err := c.LarkClient.Drive.GetDocxBlockListOfDocument(c.Ctx, &lark.GetDocxBlockListOfDocumentReq{
			DocumentID: docx.DocumentID,
			PageToken:  pageToken,
		}, options...)
		if err != nil {
			return docx, nil, err
		}
		blocks = append(blocks, resp2.Items...)
		pageToken = &resp2.PageToken
		if !resp2.HasMore {
			break
		}
	}

	return docx, blocks, nil
}

func (c *FeishuClient) GetWikiNodeInfo(token string, options ...lark.MethodOptionFunc) (*lark.GetWikiNodeRespNode, error) {
	resp, _, err := c.LarkClient.Drive.GetWikiNode(c.Ctx, &lark.GetWikiNodeReq{
		Token: token,
	}, options...)
	if err != nil {
		return nil, err
	}
	return resp.Node, nil
}

func (c *FeishuClient) GetWikiNodeList(spaceId string, parentNodeToken *string, pageToken *string, options ...lark.MethodOptionFunc) ([]*lark.GetWikiNodeListRespItem, *string, error) {
	size := int64(50)
	resp, _, err := c.LarkClient.Drive.GetWikiNodeList(c.Ctx, &lark.GetWikiNodeListReq{
		SpaceID:         spaceId,
		PageSize:        &size,
		PageToken:       pageToken,
		ParentNodeToken: parentNodeToken,
	}, options...)
	if err != nil {
		return nil, nil, err
	}
	return resp.Items, nil, nil
}

func (c *FeishuClient) GetDocumentByUrl(url string) (string, string, error) {
	reg := regexp.MustCompile("^https://[a-zA-Z0-9-]+.(feishu.cn|larksuite.com)/(docs|docx|wiki)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(url)
	if matchResult == nil || len(matchResult) != 4 {
		return "", "", errors.New(fmt.Sprintf("Invalid feishu/larksuite URL format %s", url))
	}
	docType := matchResult[2]
	nodeToken := matchResult[3]
	parser := NewParser(c.Ctx)
	if docType == "wiki" {
		node, err := c.GetWikiNodeInfo(nodeToken)
		if err != nil {
			return "", "", err
		}
		docType = node.ObjType
		nodeToken = node.ObjToken
	}

	markdown := ""
	title := ""
	if docType == "doc" {
		doc, err := c.GetDocContent(nodeToken)
		if err != nil {
			if strings.Contains(fmt.Sprintf("%v", err), "forBidden") {
				log.Printf("download docx failed %v,doc:%s", err, nodeToken)
			}
			log.Fatalf("download docx failed %v,doc:%s", err, nodeToken)
			return "", "", err
		}
		markdown = parser.ParseDocContent(doc)
		for _, element := range doc.Title.Elements {
			title += element.TextRun.Text
		}
	}
	if docType == "docx" {
		docx, blocks, err := c.GetDocxContent(nodeToken)
		if err != nil {
			if strings.Contains(fmt.Sprintf("%v", err), "forBidden") {
				log.Printf("download docx failed %v,doc:%s", err, nodeToken)
			}
			log.Fatalf("download docx failed %v,doc:%s", err, nodeToken)
			return "", "", err
		}
		markdown = parser.ParseDocxContent(docx, blocks)
		title = docx.Title
	}

	for _, imgToken := range parser.ImgTokens {
		localLink, err := c.DownloadImage(imgToken)
		if err != nil {
			return "", "", err
		}
		markdown = strings.Replace(markdown, imgToken, localLink, 1)
	}

	engine := lute.New(func(l *lute.Lute) {
		l.RenderOptions.AutoSpace = true
	})
	result := engine.FormatStr("md", markdown)
	if title == "" {
		title = nodeToken
	}
	mdName := fmt.Sprintf("%s.md", title)
	mdPath := path.Join(c.CacheDir, mdName)
	os.MkdirAll(c.CacheDir, 0o766)
	if err := os.WriteFile(mdPath, []byte(result), 0o777); err != nil {
		log.Fatalf("write markdown error %v", err)
	}
	fmt.Printf("Downloaded markdown file to %s\n", mdName)
	return mdName, mdPath, nil
}

func TestDocuments() {
	//u-cpYGwuIMF6dHNajzw55EXkh54XXxgh3bogw0gk.803d8

}
