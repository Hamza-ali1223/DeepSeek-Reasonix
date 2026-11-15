package i18n

// ScopeMessages is how the write- and delete-outside-scope cards word their
// reason, and the refusal a declined answer sends back to the model. Messages
// embeds it.
type ScopeMessages struct {
	WriteScopeReasonFmt  string // reason on the write card; %s = the folder a wider answer covers
	ScopeDeclined        string // model-facing denial when the user declines a write or delete outside the scope
	DeleteScopeReasonFmt string // reason on the delete card; %s = the targets
}

var scopeEnglish = ScopeMessages{
	WriteScopeReasonFmt:  "This file is outside the folders the agent may write to. Allowing it for this session, or always allowing it, covers the folder %s and what is below it, not the whole disk.",
	ScopeDeclined:        "the user declined this write or delete outside the workspace - do not retry it; ask how they would like to proceed.",
	DeleteScopeReasonFmt: "This command recursively deletes paths outside the workspace and the folders you granted: %s. It only runs if you allow it, and the answer covers this command alone.",
}

var scopeChinese = ScopeMessages{
	WriteScopeReasonFmt:  "这个文件在 agent 可写的文件夹之外。「本会话允许」和「始终允许」只放开 %s 及其下面的内容，不是整个磁盘。",
	ScopeDeclined:        "用户拒绝了这次工作区之外的写入或删除；不要重试，请询问用户希望如何继续。",
	DeleteScopeReasonFmt: "这条命令会递归删除工作区和你已授权的文件夹之外的路径：%s。只有你允许才会执行，这次答复只对这一条命令有效。",
}

var scopeTraditional = ScopeMessages{
	WriteScopeReasonFmt:  "這個檔案在 agent 可寫入的資料夾之外。「本工作階段允許」和「一律允許」只放開 %s 及其下面的內容，不是整個磁碟。",
	ScopeDeclined:        "使用者拒絕了這次工作區之外的寫入或刪除；不要重試，請詢問使用者希望如何繼續。",
	DeleteScopeReasonFmt: "這條指令會遞迴刪除工作區和你已授權的資料夾之外的路徑：%s。只有你允許才會執行，這次答覆只對這一條指令有效。",
}
