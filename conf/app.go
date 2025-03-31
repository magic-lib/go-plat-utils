package conf

type appStruct struct {
	appId   string
	appName string
}

var currentAppInstance = &appStruct{}

// SetApp 设置App
func SetApp(appId string, appName string) {
	if appId != "" {
		currentAppInstance.appId = appId
	}
	if appName != "" {
		currentAppInstance.appName = appName
	}
}

// AppId 获取AppId
func AppId() string {
	return currentAppInstance.appId
}

// AppName 获取AppName
func AppName() string {
	return currentAppInstance.appName
}
