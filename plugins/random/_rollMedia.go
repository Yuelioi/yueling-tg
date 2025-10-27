// package random

// // ----------------------------------------------
// // GIF / 视频处理
// // ----------------------------------------------

// func (rp *RollPlugin) handleMedia(c *context.Context, fileID string) {
// 	url, err := c.Api.GetFileDirectURL(fileID)
// 	if err != nil {
// 		rp.Log.Error().Err(err).Msg("获取文件链接失败")
// 		return
// 	}

// 	data, err := utils.FetchFile(url)
// 	if err != nil {
// 		rp.Log.Error().Err(err).Msg("下载资源失败")
// 		c.Reply("获取资源失败～ 😢")
// 		return
// 	}

// 	mime := http.DetectContentType(data)

// 	switch {
// 	case strings.HasPrefix(mime, "image/gif"):
// 		rp.handleGifData(c, data)
// 	case strings.HasPrefix(mime, "video/") || strings.HasPrefix(mime, "image/"):
// 		rp.handleVideoFrame(c, data)
// 	default:
// 		c.Reply("只支持 GIF 或短视频哦～")
// 	}
// }

// func (rp *RollPlugin) handleGifData(c *context.Context, data []byte) {
// 	img, err := gif.DecodeAll(bytes.NewReader(data))
// 	if err != nil {
// 		rp.Log.Error().Err(err).Msg("解码 GIF 失败")
// 		c.Reply("只支持 GIF 格式图片哦～")
// 		return
// 	}

// 	total := len(img.Image)
// 	if total == 0 {
// 		c.Reply("GIF 文件似乎是空的？😕")
// 		return
// 	}

// 	idx := rand.Intn(total)
// 	frame := img.Image[idx]

// 	buf := new(bytes.Buffer)
// 	if err := png.Encode(buf, frame); err != nil {
// 		rp.Log.Error().Err(err).Msg("编码帧失败")
// 		c.Reply("处理 GIF 帧失败 😭")
// 		return
// 	}

// 	c.SendPhoto(message.NewResourceFromBytes("roll_frame.png", buf.Bytes()))
// }

// func (rp *RollPlugin) handleVideoFrame(c *context.Context, data []byte) {
// 	tmpVideo := "tmp_video.mp4"
// 	if err := os.WriteFile(tmpVideo, data, 0644); err != nil {
// 		c.Reply("处理视频失败 😭")
// 		return
// 	}
// 	defer os.Remove(tmpVideo)

// 	frameFile := "frame.png"
// 	defer os.Remove(frameFile)
// 	frameIdx := rand.Intn(10) // 随机选择前 10 帧

// 	cmd := exec.Command("ffmpeg", "-y", "-i", tmpVideo, "-vf", fmt.Sprintf("select=eq(n\\,%d)", frameIdx), "-vframes", "1", frameFile)
// 	if err := cmd.Run(); err != nil {
// 		c.Reply("提取视频帧失败 😭")
// 		return
// 	}

// 	frameData, err := os.ReadFile(frameFile)
// 	if err != nil {
// 		c.Reply("读取视频帧失败 😭")
// 		return
// 	}

// 	c.SendPhoto(message.NewResourceFromBytes("roll_frame.png", frameData))
// }
