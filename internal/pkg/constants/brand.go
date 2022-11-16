package constants

const (
	BSCreateSuccess     = 1  // 新建成功
	BSBuilding          = 2  //新建中，待更新状态
	BSCreateDbSuccess   = 3  //新建数据库成功
	BSCreateTbSuccess   = 4  //新建数据表成功
	BSCreateDbFailed    = 5  //新建数据库失败
	BSCreateTbFailed    = 6  //新建数据表失败
	BSInitDataFailed    = 7  //数据初始化失败
	BSPrepareDataFailed = 8  //数据准备失败
	BSDel               = -1 //删除状态
)
