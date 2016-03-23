package water

// 请在Context处理流程的最后才Write,因为在panic前写入数据会导致Recovery()无法写入正确的HttpStatus.
// water的Context不内嵌Logger,推荐使用功能更强大第三方的Logger.
