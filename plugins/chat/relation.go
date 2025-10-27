package chat

import "fmt"

// -------------------- 好感度管理 --------------------

func (cp *ChatPlugin) getUserLike(userID int64) int {
	cp.userPrefs.mu.RLock()
	defer cp.userPrefs.mu.RUnlock()

	key := fmt.Sprintf("%d", userID)
	if like, ok := cp.userPrefs.Prefs[key]; ok {
		return like
	}
	return 50 // 默认值
}

func (cp *ChatPlugin) updateUserLike(userID int64, change int) int {
	cp.userPrefs.mu.Lock()
	defer cp.userPrefs.mu.Unlock()

	key := fmt.Sprintf("%d", userID)
	current := cp.userPrefs.Prefs[key]
	if current == 0 {
		current = 50
	}

	newLike := current + change
	if newLike < 0 {
		newLike = 0
	}
	if newLike > 100 {
		newLike = 100
	}

	cp.userPrefs.Prefs[key] = newLike

	// 保存到文件
	go cp.savePrefs()

	return newLike
}

func (cp *ChatPlugin) getRelationshipInfo(userLike int) RelationshipInfo {
	if userLike >= 80 {
		return RelationshipInfo{
			Status:       "很喜欢",
			Attitude:     "亲密撒娇，温柔可爱，像好朋友",
			Mode:         "亲密模式",
			Relationship: "挚友",
		}
	} else if userLike >= 60 {
		return RelationshipInfo{
			Status:       "喜欢",
			Attitude:     "友好温和，偶尔撒娇",
			Mode:         "友好模式",
			Relationship: "好朋友",
		}
	} else if userLike >= 40 {
		return RelationshipInfo{
			Status:       "普通",
			Attitude:     "正常聊天，不冷不热",
			Mode:         "普通模式",
			Relationship: "普通朋友",
		}
	} else if userLike >= 20 {
		return RelationshipInfo{
			Status:       "不太喜欢",
			Attitude:     "有点冷淡，回复简短",
			Mode:         "冷淡模式",
			Relationship: "陌生人",
		}
	} else {
		return RelationshipInfo{
			Status:       "讨厌",
			Attitude:     "明显不耐烦，语气生硬",
			Mode:         "讨厌模式",
			Relationship: "有敌意的陌生人",
		}
	}
}
