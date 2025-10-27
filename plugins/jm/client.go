package jm

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	APP_TOKEN_SECRET   = "18comicAPP"
	APP_TOKEN_SECRET_2 = "18comicAPPContent"
	APP_DATA_SECRET    = "185Hcomic3PAPP7R"
	APP_VERSION        = "1.7.5"
	API_DOMAIN         = "www.cdnblackmyth.club"
	IMAGE_DOMAIN       = "cdn-msp2.18comic.vip"
)

type ApiPath string

const (
	ApiPathSearch        ApiPath = "/search"
	ApiPathGetComic      ApiPath = "/album"
	ApiPathGetChapter    ApiPath = "/chapter"
	ApiPathGetScrambleId ApiPath = "/chapter_view_template"
	ApiPathGetWeeklyInfo ApiPath = "/week"
	ApiPathGetWeekly     ApiPath = "/week/filter"
)

type SearchSort string

const (
	SearchSortDefault SearchSort = ""
	SearchSortLatest  SearchSort = "mv"
	SearchSortPopular SearchSort = "mp"
)

type ImageFormat string

const (
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatWEBP ImageFormat = "webp"
	ImageFormatGIF  ImageFormat = "gif"
)

type JmClient struct {
	saveDir     string
	concurrency int
	httpClient  *http.Client
}

func NewJmClient(client *http.Client) *JmClient {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &JmClient{
		httpClient:  client,
		concurrency: 5,
	}
}

func (c *JmClient) WithConcurrency(concurrency int) *JmClient {
	if concurrency < 1 {
		c.concurrency = 1
	} else if concurrency > 12 {
		c.concurrency = 12
	}
	c.concurrency = concurrency
	return c
}

func (c *JmClient) WithSaveDir(saveDir string) *JmClient {
	c.saveDir = saveDir
	return c
}

func (c *JmClient) jmRequest(method string, path ApiPath, query map[string]string, ts int64) (*http.Response, error) {
	tokenparam := fmt.Sprintf("%d,%s", ts, APP_VERSION)
	var token string

	if path == ApiPathGetScrambleId {
		token = md5Hex(fmt.Sprintf("%d%s", ts, APP_TOKEN_SECRET_2))
	} else {
		token = md5Hex(fmt.Sprintf("%d%s", ts, APP_TOKEN_SECRET))
	}

	apiUrl := fmt.Sprintf("https://%s%s", API_DOMAIN, path)

	// Add query parameters
	if len(query) > 0 {
		params := url.Values{}
		for k, v := range query {
			params.Add(k, v)
		}
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, params.Encode())
	}

	req, err := http.NewRequest(method, apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("token", token)
	req.Header.Set("tokenparam", tokenparam)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")

	return c.httpClient.Do(req)
}

func (c *JmClient) jmGet(path ApiPath, query map[string]string, ts int64) (*http.Response, error) {
	return c.jmRequest(http.MethodGet, path, query, ts)
}

func decryptData(ts int64, data string) (string, error) {
	// Base64解码
	encryptedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	// 生成密钥
	key := md5Hex(fmt.Sprintf("%d%s", ts, APP_DATA_SECRET))

	// AES-256-ECB解密
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(encryptedData)%aes.BlockSize != 0 {
		return "", fmt.Errorf("encrypted data is not a multiple of block size")
	}

	decrypted := make([]byte, len(encryptedData))

	// ECB mode decryption
	for i := 0; i < len(encryptedData); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], encryptedData[i:i+aes.BlockSize])
	}

	// 去除PKCS#7填充
	paddingLength := int(decrypted[len(decrypted)-1])
	if paddingLength > len(decrypted) || paddingLength > aes.BlockSize {
		return "", fmt.Errorf("invalid padding")
	}

	decrypted = decrypted[:len(decrypted)-paddingLength]

	return string(decrypted), nil
}

// Public API methods

func (c *JmClient) Search(keyword string, page int, sort SearchSort) (interface{}, error) {
	ts := time.Now().Unix()
	query := map[string]string{
		"main_tag":     "0",
		"search_query": keyword,
		"page":         strconv.Itoa(page),
		"o":            string(sort),
	}

	resp, err := c.jmGet(ApiPathSearch, query, ts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索失败，预料之外的状态码(%d): %s", resp.StatusCode, string(body))
	}

	var jmResp JmResp
	if err := json.Unmarshal(body, &jmResp); err != nil {
		return nil, fmt.Errorf("将body解析为JmResp失败: %w", err)
	}

	if jmResp.Code != 200 {
		return nil, fmt.Errorf("搜索失败，预料之外的code: %d", jmResp.Code)
	}

	var dataStr string
	if err := json.Unmarshal(jmResp.Data, &dataStr); err != nil {
		return nil, fmt.Errorf("data字段不是字符串: %w", err)
	}

	decryptedData, err := decryptData(ts, dataStr)
	if err != nil {
		return nil, fmt.Errorf("解密data失败: %w", err)
	}

	// 尝试解析为重定向数据
	var redirectData RedirectRespData
	if err := json.Unmarshal([]byte(decryptedData), &redirectData); err == nil && redirectData.RedirectAid != "" {
		aid, err := strconv.Atoi(redirectData.RedirectAid)
		if err != nil {
			return nil, err
		}
		return c.GetComic(aid)
	}

	// 解析为搜索结果
	var searchData SearchRespData
	if err := json.Unmarshal([]byte(decryptedData), &searchData); err != nil {
		return nil, fmt.Errorf("将解密后的数据解析失败: %w", err)
	}

	return &searchData, nil
}

func (c *JmClient) GetComic(aid int) (*GetComicRespData, error) {
	ts := time.Now().Unix()
	query := map[string]string{
		"id": strconv.Itoa(aid),
	}

	resp, err := c.jmGet(ApiPathGetComic, query, ts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取漫画失败，预料之外的状态码(%d): %s", resp.StatusCode, string(body))
	}

	var jmResp JmResp
	if err := json.Unmarshal(body, &jmResp); err != nil {
		return nil, fmt.Errorf("将body解析为JmResp失败: %w", err)
	}

	if jmResp.Code != 200 {
		return nil, fmt.Errorf("获取漫画失败，预料之外的code: %d", jmResp.Code)
	}

	var dataStr string
	if err := json.Unmarshal(jmResp.Data, &dataStr); err != nil {
		return nil, fmt.Errorf("data字段不是字符串: %w", err)
	}

	decryptedData, err := decryptData(ts, dataStr)
	if err != nil {
		return nil, fmt.Errorf("解密data失败: %w", err)
	}

	var comic GetComicRespData
	if err := json.Unmarshal([]byte(decryptedData), &comic); err != nil {
		return nil, fmt.Errorf("将解密后的data字段解析为GetComicRespData失败: %w", err)
	}

	return &comic, nil
}

func (c *JmClient) GetChapter(id int) (*GetChapterRespData, error) {
	ts := time.Now().Unix()
	query := map[string]string{
		"id": strconv.Itoa(id),
	}

	resp, err := c.jmGet(ApiPathGetChapter, query, ts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取章节失败，预料之外的状态码(%d): %s", resp.StatusCode, string(body))
	}

	var jmResp JmResp
	if err := json.Unmarshal(body, &jmResp); err != nil {
		return nil, fmt.Errorf("将body解析为JmResp失败: %w", err)
	}

	if jmResp.Code != 200 {
		return nil, fmt.Errorf("获取章节失败，预料之外的code: %d", jmResp.Code)
	}

	var dataStr string
	if err := json.Unmarshal(jmResp.Data, &dataStr); err != nil {
		return nil, fmt.Errorf("data字段不是字符串: %w", err)
	}

	decryptedData, err := decryptData(ts, dataStr)
	if err != nil {
		return nil, fmt.Errorf("解密data失败: %w", err)
	}

	var chapter GetChapterRespData
	if err := json.Unmarshal([]byte(decryptedData), &chapter); err != nil {
		return nil, fmt.Errorf("将解密后的data字段解析为GetChapterRespData失败: %w", err)
	}

	return &chapter, nil
}

func (c *JmClient) GetScrambleId(id int) (int, error) {
	ts := time.Now().Unix()
	query := map[string]string{
		"id":            strconv.Itoa(id),
		"v":             strconv.FormatInt(ts, 10),
		"mode":          "vertical",
		"page":          "0",
		"app_img_shunt": "1",
		"express":       "off",
	}

	resp, err := c.jmGet(ApiPathGetScrambleId, query, ts)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("获取scramble_id失败，预料之外的状态码(%d): %s", resp.StatusCode, string(body))
	}

	// 从响应中提取scramble_id
	bodyStr := string(body)
	prefix := "var scramble_id = "
	start := strings.Index(bodyStr, prefix)
	if start == -1 {
		return 220980, nil // 默认值
	}

	start += len(prefix)
	end := strings.Index(bodyStr[start:], ";")
	if end == -1 {
		return 220980, nil
	}

	scrambleIdStr := bodyStr[start : start+end]
	scrambleId, err := strconv.Atoi(strings.TrimSpace(scrambleIdStr))
	if err != nil {
		return 220980, nil
	}

	return scrambleId, nil
}

func (c *JmClient) GetWeeklyInfo() (*GetWeeklyInfoRespData, error) {
	ts := time.Now().Unix()

	resp, err := c.jmGet(ApiPathGetWeeklyInfo, nil, ts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取每周必看信息失败，预料之外的状态码(%d): %s", resp.StatusCode, string(body))
	}

	var jmResp JmResp
	if err := json.Unmarshal(body, &jmResp); err != nil {
		return nil, fmt.Errorf("将body解析为JmResp失败: %w", err)
	}

	if jmResp.Code != 200 {
		return nil, fmt.Errorf("获取每周必看信息失败，预料之外的code: %d", jmResp.Code)
	}

	var dataStr string
	if err := json.Unmarshal(jmResp.Data, &dataStr); err != nil {
		return nil, fmt.Errorf("data字段不是字符串: %w", err)
	}

	decryptedData, err := decryptData(ts, dataStr)
	if err != nil {
		return nil, fmt.Errorf("解密data失败: %w", err)
	}

	var weeklyInfo GetWeeklyInfoRespData
	if err := json.Unmarshal([]byte(decryptedData), &weeklyInfo); err != nil {
		return nil, fmt.Errorf("将解密后的data字段解析为GetWeeklyInfoRespData失败: %w", err)
	}

	return &weeklyInfo, nil
}

func (c *JmClient) GetWeekly(categoryId, typeId string) (*GetWeeklyRespData, error) {
	ts := time.Now().Unix()
	query := map[string]string{
		"id":   categoryId,
		"type": typeId,
	}

	resp, err := c.jmGet(ApiPathGetWeekly, query, ts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取每周必看失败，预料之外的状态码(%d): %s", resp.StatusCode, string(body))
	}

	var jmResp JmResp
	if err := json.Unmarshal(body, &jmResp); err != nil {
		return nil, fmt.Errorf("将body解析为JmResp失败: %w", err)
	}

	if jmResp.Code != 200 {
		return nil, fmt.Errorf("获取每周必看失败，预料之外的code: %d", jmResp.Code)
	}

	var dataStr string
	if err := json.Unmarshal(jmResp.Data, &dataStr); err != nil {
		return nil, fmt.Errorf("data字段不是字符串: %w", err)
	}

	decryptedData, err := decryptData(ts, dataStr)
	if err != nil {
		return nil, fmt.Errorf("解密data失败: %w", err)
	}

	var weekly GetWeeklyRespData
	if err := json.Unmarshal([]byte(decryptedData), &weekly); err != nil {
		return nil, fmt.Errorf("将解密后的data字段解析为GetWeeklyRespData失败: %w", err)
	}

	return &weekly, nil
}

// GetImageData 下载图片数据，如果图片为空会自动重试
func (c *JmClient) GetImageData(imageUrl string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, imageUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 如果图片为空，带上时间戳再次请求
	if len(imageData) == 0 {
		ts := time.Now().Unix()
		retryUrl := fmt.Sprintf("%s?ts=%d", imageUrl, ts)

		req, err := http.NewRequest(http.MethodGet, retryUrl, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("重试下载图片失败，状态码: %d", resp.StatusCode)
		}

		imageData, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return imageData, nil
}

// GetChapterImageInfos 获取章节的所有图片信息（URL和block_num）
func (c *JmClient) GetChapterImageInfos(chapterId int) ([]ImageInfo, error) {
	// 获取scramble_id和章节数据
	scrambleId, err := c.GetScrambleId(chapterId)
	if err != nil {
		return nil, fmt.Errorf("获取scramble_id失败: %w", err)
	}

	chapterData, err := c.GetChapter(chapterId)
	if err != nil {
		return nil, fmt.Errorf("获取章节数据失败: %w", err)
	}

	// 构造图片信息列表
	var imageInfos []ImageInfo
	for i, filename := range chapterData.Images {
		ext := strings.ToLower(filepath.Ext(filename))
		url := fmt.Sprintf("https://%s/media/photos/%d/%s", IMAGE_DOMAIN, chapterId, filename)

		var blockNum uint32
		if ext == ".gif" {
			blockNum = 0
		} else if ext == ".webp" {
			filenameWithoutExt := strings.TrimSuffix(filename, ext)
			blockNum = calculateBlockNum(int64(scrambleId), int64(chapterId), filenameWithoutExt)
		} else {
			// 跳过不支持的格式
			continue
		}

		imageInfos = append(imageInfos, ImageInfo{
			URL:      url,
			BlockNum: blockNum,
			Index:    i,
		})
	}

	return imageInfos, nil
}

// DownloadAndSaveImage 下载图片并保存到指定路径（自动处理图片拼接）
// outputFormat: 输出格式，支持 jpeg, png, webp, gif
func (c *JmClient) DownloadAndSaveImage(imageInfo ImageInfo, savePath string, outputFormat ImageFormat) error {

	imageData, err := c.GetImageData(imageInfo.URL)
	if err != nil {
		return fmt.Errorf("下载图片失败: %w", err)
	}

	// 检测图片格式
	srcFormat, err := detectImageFormat(imageData)
	if err != nil {
		return fmt.Errorf("检测图片格式失败: %w", err)
	}

	// 如果是GIF，直接保存
	if srcFormat == ImageFormatGIF {
		return os.WriteFile(savePath, imageData, 0644)
	}

	// 解码图片
	img, err := decodeImage(imageData, srcFormat)
	if err != nil {
		return fmt.Errorf("解码图片失败: %w", err)
	}

	// 如果需要拼接图片
	if imageInfo.BlockNum > 0 {
		img = stitchImage(img, imageInfo.BlockNum)
	}

	// 保存图片
	return saveImage(img, savePath, outputFormat)
}

// DownloadChapterImages 下载整个章节的所有图片
// outputFormat: 输出格式
func (c *JmClient) DownloadChapterImages(chapterId int, outputFormat ImageFormat) error {

	saveDir := filepath.Join(c.saveDir, strconv.Itoa(chapterId))
	// 创建保存目录
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("创建保存目录失败: %w", err)
	}

	// 获取所有图片信息
	imageInfos, err := c.GetChapterImageInfos(chapterId)
	if err != nil {
		return fmt.Errorf("获取图片信息失败: %w", err)
	}

	// 只保留前25张
	imageInfos = imageInfos[:25]

	// 创建任务通道和错误通道
	tasks := make(chan ImageInfo, len(imageInfos))
	errors := make(chan error, len(imageInfos))

	// 使用 sync.WaitGroup 等待所有goroutine完成
	var wg sync.WaitGroup

	// 启动worker goroutines
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for imageInfo := range tasks {
				// 构造文件名
				filename := fmt.Sprintf("%04d.%s", imageInfo.Index+1, outputFormat)
				savePath := filepath.Join(saveDir, filename)

				// 如果文件已存在，跳过
				if _, err := os.Stat(savePath); err == nil {
					continue
				}

				// 下载并保存图片
				if err := c.DownloadAndSaveImage(imageInfo, savePath, outputFormat); err != nil {
					errors <- fmt.Errorf("下载第%d张图片失败: %w", imageInfo.Index+1, err)
					return
				}
			}
		}()
	}

	// 发送所有任务
	for _, imageInfo := range imageInfos {
		tasks <- imageInfo
	}
	close(tasks)

	// 等待所有任务完成
	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}
